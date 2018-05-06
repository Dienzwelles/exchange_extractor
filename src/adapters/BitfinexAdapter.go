package adapters

import (
	"../adapters/custom/bitfinexcustom"
	"../models"
	"time"
	"log"
	"../datastorage"
	_ "fmt"
	"strings"
	_"github.com/shopspring/decimal"
	_"github.com/thrasher-/gocryptotrader/exchanges/bitfinex"
	"github.com/bitfinexcom/bitfinex-api-go/v1"
	"fmt"
	"../properties"
	"strconv"
	"golang.org/x/sync/syncmap"
)


const BITFINEX  = "Bitfinex"


//contiene timestamp ultima richiesta fatta per symbol
//sposto i ragionamenti prima fatti su StartMs su questa
//struttura
var ts_bitfinex_transactions = map[string]int64{}
var lastLot int64
var lastLotSymbol string
var bookmap *syncmap.Map


type BitfinexAdapter struct{
	AbstractAdapter
	StartMs int64
	initBook bool
	chanbook chan []models.AggregateBooks
}

func NewBitfinexAdapter() AdapterInterface {
	return BitfinexAdapter{initBook:false}
}

func asyncExtractAll(chanchannels [] chan []float64, synchs [] chan int, symbols []string, channel string, wss *bitfinexcustom.WebSocketService){
	wss.Connect()

	for i := 0; i < len(symbols); i++ {
		s :=bitfinexcustom.SubscribeToChannelData{
			Channel: channel,
			Pair: strings.ToUpper(symbols[i]),
			Chan: chanchannels[i],
			Precision: "R0",
			Length: "25",
		}
		wss.AddSubscribeFull(s)
		//wss.AddSubscribePrecision(channel, strings.ToUpper(symbols[i]), "R0", chanchannels[i])
	}

	for i := 0; i < len(synchs); i++ {
		<- synchs[i]
	}

	wss.Subscribe()
}

func extractAll(chanchannels [] chan []float64, synchs [] chan int, symbols []string, channel string) *bitfinexcustom.WebSocketService{
	ac := properties.GetInstance()
	client := bitfinex.NewClient().Auth(ac.Bitfinex.Key, ac.Bitfinex.Secret)
	/*
	trades, err:= client.Trades.All("BTCUSD", time.Now(), 50)

	if err != nil{
		panic(err)
	}

	print(trades)
	*/
	wss := bitfinexcustom.NewWebSocketService(client)

	go asyncExtractAll(chanchannels, synchs, symbols, channel, wss)

	return wss
}

func asyncExtract(chanchannel chan []float64, synch chan int, symbol string, channel string, wss *bitfinex.WebSocketService){
	wss.AddSubscribe(channel, strings.ToUpper(symbol), chanchannel)

	<- synch

	wss.Subscribe()
}

func extract(chanchannel chan []float64, synch chan int, symbol string, channel string) *bitfinex.WebSocketService{
	ac := properties.GetInstance()
	client := bitfinex.NewClient().Auth(ac.Bitfinex.Key, ac.Bitfinex.Secret)

	trades, err:= client.Trades.All("BTCUSD", time.Now(), 50)

	if err != nil{
		panic(err)
	}

	print(trades)

	wss := bitfinex.NewWebSocketService(client)
	wss.Connect()

	go asyncExtract(chanchannel, synch, symbol, channel, wss)

	return wss
}

func waitTrades(chantrade chan []models.Trade, chanchannel chan []float64, synch chan int, exchangeId string, symbol string){
	synch <- 1

	for {
		rawtrade := <-chanchannel

		//diff := 1
		tid := "0"
		if len(rawtrade) == 4 {
			//diff = 0
			tid = strconv.FormatFloat(rawtrade[0], 'G', 40, 64)
			chantrade <- []models.Trade{models.Trade{Exchange_id: exchangeId, Symbol: strings.ToUpper(symbol),
				Trade_ts: time.Unix(int64(rawtrade[1 /*- diff*/]), 0), Amount: rawtrade[3 /*- diff*/], Price: rawtrade[2 /*- diff*/],
				Tid: tid}}
		}

	}
}

func waitBooks(chanbook chan []models.AggregateBooks, chanchannel chan []float64, synch chan int, reset [] chan int, exchangeId string, symbol string){
	synch <- 1

	for {
		rawbook := <-chanchannel

		if(len(rawbook) != 3){
			//println("pluto: " + symbol)
			lastLot++

			retChanbooks := []models.AggregateBooks{}

			bookmap.Range(func(ki, vi interface{}) bool {
				_, v := ki.(float64), vi.(models.AggregateBooks)

				v.Lot = lastLot
				retChanbooks = append(retChanbooks, v)

				return true
			})

			//print("lunghezza array: ")
			//println(len(retChanbooks))

			chanbook <- retChanbooks
			//chanbooks <- []models.AggregateBook{models.AggregateBook{Exchange_id: exchangeId, Symbol: strings.ToUpper(symbol), Price: 0, Count_number: 0, Amount: rawbook[0], Lot: lastLot, Obsolete: true}}
		} else {
			aggregateBook := models.AggregateBooks{Exchange_id: exchangeId, Symbol: strings.ToUpper(symbol), Price: rawbook[1], Count_number: 1, Amount: rawbook[2], Lot: lastLot, Obsolete: false}

			if aggregateBook.Price == 0 {
				bookmap.Delete(rawbook[0])
			} else {
				bookmap.Store(rawbook[0], aggregateBook)
			}

		}
	}
}

/*
func closeWss(wss *bitfinex.WebSocketService, closed bool){
	if !closed{
		wss.ClearSubscriptions()
		wss.Close()
	}
}
*/
func (ba BitfinexAdapter) instantiateExtracts(symbols []string, chanbook chan []models.AggregateBooks, chanchannels []chan []float64, synchs []chan int, reset [] chan int){
	//for {
	print("1:extract-open lot:")
	println(lastLot)
	wss := extractAll(chanchannels, synchs, symbols, "book")
	for i := 0; i < len(symbols); i++ {
		ba.instantiateExtract(symbols[i], chanbook, chanchannels[i], synchs[i], reset, i)
	}

	for i := 0; i < len(reset); i++ {
		<-reset[i]
	}

	println("2:extract-closing ")
	wss.ClearSubscriptions()
	wss.Close()
	print("3:extractclosed ")
	println(lastLot)
	time.Sleep(200 * time.Millisecond)
	//}
}

func (ba BitfinexAdapter) instantiateExtract(symbol string, chanbook chan []models.AggregateBooks, chanchannel chan []float64, synch chan int, reset [] chan int, i int){
	go waitBooks(chanbook, chanchannel, synch, reset, ba.ExchangeId, symbol)
}

func (ba BitfinexAdapter) getTrade() [] chan []models.Trade {

	//ritorna la lista dei mercati attualmente attivi
	symbols := datastorage.GetMarkets(BITFINEX)
	var chanchannels [] chan []float64
	var chantrades [] chan []models.Trade
	var synchs [] chan int

	for i := 0; i < len(symbols); i++ {
		chanchannels = append(chanchannels, make(chan []float64))
		synchs = append(synchs, make(chan int))
		chantrades = append(chantrades, make(chan []models.Trade))
	}

	for i := 0; i < len(synchs); i++ {
		extract(chanchannels[i], synchs[i], symbols[i], "trades")
		go waitTrades(chantrades[i], chanchannels[i], synchs[i], ba.ExchangeId, symbols[i])
	}

	return chantrades
}

func (ba BitfinexAdapter) getAggregateBooks() (chan []models.AggregateBooks, chan int) {

	lastLot = datastorage.GetLastLot(ba.ExchangeId, ba.Symbol)
	//log.Println(ba.Symbol, lastLot)
	lastLot++
	//log.Println(ba.Symbol , lastLot)

	symbols := datastorage.GetMarkets(BITFINEX)
	var chanchannels [] chan []float64
	var synchs [] chan int
	var reset [] chan int
	outReset := make(chan int)

	if !ba.initBook {
		ba.chanbook = make(chan []models.AggregateBooks)
	}

	for i := 0; i < len(symbols); i++ {
		chanchannels = append(chanchannels, make(chan []float64))
		synchs = append(synchs, make(chan int))
		reset = append(reset, make(chan int))
	}

	bookmap = new(syncmap.Map)

	ba.initBook = true

	//for i := 0; i < len(synchs); i++ {
	go ba.instantiateExtracts(symbols, ba.chanbook, chanchannels, synchs, reset)
	//}


	return ba.chanbook, outReset
}

func (ba BitfinexAdapter) instantiateDefault(symbol string) AdapterInterface {
	ba.ExchangeId = BITFINEX
	aa := ba.abstractInstantiateDefault(symbol)
	ba.AbstractAdapter = aa
	return ba
}

func (ba BitfinexAdapter) instantiate(Symbol string, FetchSize int, ReloadInterval int) AdapterInterface {
	ba.ExchangeId = BITFINEX
	aa := ba.abstractInstantiate(Symbol, FetchSize, ReloadInterval)
	ba.AbstractAdapter = aa
	return ba
}


func CheckBitfinexRecord (symbol string, time_last time.Time, quantity float64) bool {
	conn := datastorage.NewConnection()
	db := datastorage.GetConnection(conn)
	defer db.Close()

	type el struct {
		sy string
		ts time.Time
		am float64
	}

	var record el

	rows, err := db.Query("select symbol, trade_ts, amount from trades where exchange_id = '" + BITFINEX + "' and symbol = ? and trade_ts = ? group by symbol, trade_ts, amount ", symbol, time_last)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next(){
		err:=rows.Scan(&record.sy, &record.ts, &record.am)
		if err != nil {
			log.Fatal(err)
		}
		if record.am == quantity {
			return false
		}
	}

	return true
}

func (ba BitfinexAdapter) executeArbitrage(arbitrage models.Arbitrage) bool  {
	fmt.Println("attivata funzione arbitraggio")
	//historicalArbitrage := models.HistoricalArbitrage{Exchange_id:BITFINEX, SymbolStart: arbitrage.SymbolStart, SymbolTransitory: arbitrage.SymbolTransitory, SymbolEnd: arbitrage.SymbolEnd  }

	//ottengo application context
	ac := properties.GetInstance()
	client := bitfinex.NewClient().Auth(ac.Bitfinex.Key, ac.Bitfinex.Secret)

	valueTrade := arbitrage.PriceStart * arbitrage.AmountStart

	//condition to test limit valuo for the trading
	if valueTrade < (20 * (1 + 0.006)) {
		fmt.Println("funzione arbitraggio - valore vendita troppo basso")
		return false
	}

	// case buy
	if ac.Bitfinex.ExecArbitrage == "S" {

		fmt.Println("funzione arbitraggio vendita - esecuzione trade")
		//return the max available quantity within the wallet for the first cross side b
		//q0 := GetAvailableQuantity(arbitrage.SymbolStart, client, true)


		ratio := 1.0
		//check if quantity is ok
		//if valueTrade  > q0 {
			fmt.Println("funzione arbitraggio - valore da scambiare eccede la dispomibilità massima")
			ratio = 1.0 * 70 / valueTrade
			//return false
		//}

		print("ratio: ", ratio)


		fmt.Println("funzione arbitraggio - step 1 valore da scambiare: ", arbitrage.AmountStart *ratio)
		//execute first trade
		order, err := client.Orders.Create(arbitrage.SymbolStart, arbitrage.AmountStart * ratio, 3, bitfinex.OrderTypeExchangeMarket)
		if err != nil {
			fmt.Println("errore durante trade 1")
			fmt.Println(err)
			return false
		} else {
			fmt.Println("trade 1 avvenuto")
			//time.Sleep(1000 * time.Millisecond)
			fmt.Println(order)
		}


		if arbitrage.AmountTransitory != 0 && arbitrage.AmountEnd != 0 {

			fmt.Println("funzione arbitraggio - step 2 da scambiare: ", arbitrage.AmountTransitory * ratio * 0.099785)
			order1, err := client.Orders.Create(arbitrage.SymbolTransitory, arbitrage.AmountTransitory * ratio * 0.099785, 3, bitfinex.OrderTypeExchangeMarket)
			if err != nil {
				fmt.Println("errore durante trade 2")
				fmt.Println(err)
				return false
			} else {
				fmt.Println("trade 2 avvenuto")
				//time.Sleep(1000 * time.Millisecond)
				fmt.Println(order1)
			}



			fmt.Println("funzione arbitraggio - step 3 valore da scambiare: ", arbitrage.AmountEnd * ratio * 0.998875)
			//execute second trade
			order2, err := client.Orders.Create(arbitrage.SymbolEnd, arbitrage.AmountEnd * ratio * 0.998875, 3, bitfinex.OrderTypeExchangeMarket)
			if err != nil {
				fmt.Println("errore durante trade 3")
				fmt.Println(err)
				return false
			} else {
				fmt.Println("trade 3 avvenuto")
				time.Sleep(500 * time.Millisecond)
				fmt.Println(order2)
				balances, err := client.Balances.All()
				if err != nil {
					fmt.Println("funzione arbitraggio - errore reperimento quantità")
					fmt.Println(err)
				} else {
					fmt.Println("situazione finale arbitraggi - portafoglio valute")
					fmt.Println(balances)
				}
			}



		}else {
			//return initial available quantity for the first cross side a
			initalQuantity := GetAvailableQuantity(arbitrage.SymbolStart, client, false)

			//return the max available quantity within the wallet for the second cross side b
			q1 := GetAvailableQuantity(arbitrage.SymbolTransitory, client, true)
			initalQuantity2 := GetAvailableQuantity(arbitrage.SymbolTransitory, client, false)
			fmt.Println("funzione arbitraggio - q1", q1)
			fmt.Println("funzione arbitraggio - initalQuantity2", initalQuantity2)

			tick, err := client.Ticker.Get(arbitrage.SymbolTransitory)

			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("ticker %s  %s", arbitrage.SymbolTransitory, tick.Ask)
				fmt.Println(tick.Ask)
			}
			divisore, _ := strconv.ParseFloat(tick.Ask, 64)
			amount := (q1 - initalQuantity) / divisore
			fmt.Println("funzione arbitraggio - step 1 da scambiare: ", amount)
			//execute second trade
			order1, err := client.Orders.Create(arbitrage.SymbolTransitory, amount, 3, bitfinex.OrderTypeExchangeMarket)
			if err != nil {
				fmt.Println("errore durante acquisto 1")
				fmt.Println(err)
				return false
			} else {
				fmt.Println("acquisto 1 avvenuto")
				time.Sleep(1000 * time.Millisecond)
				fmt.Println(order1)
			}

			//get value for third trade
			q2 := GetAvailableQuantity(arbitrage.SymbolTransitory, client, false)
			//inverto q2 e initial quantity perchè devo vendere
			amount2 := initalQuantity2 - q2
			fmt.Println("funzione arbitraggio - step 2 valore da scambiare: ", amount2)
			//execute second trade
			order2, err := client.Orders.Create(arbitrage.SymbolEnd, amount2, 3, bitfinex.OrderTypeExchangeMarket)
			if err != nil {
				fmt.Println("errore durante acquisto 2")
				fmt.Println(err)
				return false
			} else {
				fmt.Println("acquisto 2 avvenuto")
				time.Sleep(1000 * time.Millisecond)
				fmt.Println(order2)
				balances, err := client.Balances.All()
				if err != nil {
					fmt.Println("funzione arbitraggio - errore reperimento quantità")
					fmt.Println(err)
				} else {
					fmt.Println("situazione finale arbitraggi - portafoglio valute")
					fmt.Println(balances)
				}
			}
		}

	}

	return true
}



func GetAvailableQuantity(symbol string, client *bitfinex.Client, buy bool) float64  {
	// buy have to be true if you need to buy
	currency1 := strings.ToUpper(symbol[0:3])
	currency2 := strings.ToUpper(symbol[3:6])

	/*fmt.Println(currency1)
	fmt.Println(currency2)*/
	//info, err := client.Account.Info()
	balances, err := client.Balances.All()
	if err != nil {
		fmt.Println("funzione arbitraggio - errore reperimento quantità")
		fmt.Println(err)
	} else {
		fmt.Println("funzione arbitraggio - portafoglio valute")
		fmt.Println(balances)
	}
	for _, balance := range balances {
		if strings.Compare("exchange",balance.Type) == 0 {
			currency := strings.ToUpper(balance.Currency)
			if strings.Compare(currency, currency2) == 0 && buy {
				q, _ := strconv.ParseFloat(balance.Available, 64)
				fmt.Printf("funzione arbitraggio - quantità disponibile %s: %f\n",balance.Currency, q)
				return q
			}
			if strings.Compare(currency, currency1) == 0 && !buy {
				q, _ := strconv.ParseFloat(balance.Available, 64)
				fmt.Printf("funzione arbitraggio - quantità disponibile %s: %f\n",balance.Currency, q)
				return q
			}
		}
	}
	return 0
}

/*
func closeChannels(reset chan int, chanchannel chan []float64, synch chan int){
	<- reset
	time.Sleep(1000 * time.Millisecond)
	close(chanchannel)

	close(synch)
}
*/
/*
url = "https://api.bitfinex.com/v2/trades/t" + strings.ToUpper(symbol) + "/hist?sort=1"
		if ba.FetchSize > 0 {
			url = url + "&limit=" + strconv.Itoa(ba.FetchSize)
		}

		if ts_bitfinex_transactions[symbol] == 0 {
			ts_bitfinex_transactions[symbol] = time.Now().Unix() * 1000
		}

		url = url + "&start=" + strconv.FormatInt(ts_bitfinex_transactions[symbol], 10)

		httpClient := http.Client{
			Timeout: time.Second * 10, // Maximum of 2 secs
		}

		req, err := http.NewRequest(http.MethodGet, url, nil)

		if err != nil {
			log.Fatal(err)
		}

		req.Header.Set("User-Agent", "bitfinex-extractor")
		res, getErr := httpClient.Do(req)

		if getErr != nil {
			log.Fatal(getErr)
		}

		body, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			log.Fatal(readErr)
		}

		var rawtrades [][4]float64
		jsonErr := json.Unmarshal(body, &rawtrades)
		if jsonErr != nil {
			log.Fatal(jsonErr)
		}

 */