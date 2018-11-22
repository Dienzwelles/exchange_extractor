package adapters

import (
	"../utils"
	"bytes"
	"fmt"
	"../models"
	"time"
	"log"
	"../datastorage"
	_ "fmt"
	"strings"
	_"github.com/shopspring/decimal"
	_"github.com/thrasher-/gocryptotrader/exchanges/bitfinex"
	"github.com/bitfinexcom/bitfinex-api-go/v2"
	"github.com/bitfinexcom/bitfinex-api-go/v2/websocket"
	"strconv"
	"sync"
	"context"
	"golang.org/x/sync/syncmap"
	"github.com/bitfinexcom/bitfinex-api-go/v2/rest"
	"path"
	"net/url"
)


//contiene timestamp ultima richiesta fatta per symbol
//sposto i ragionamenti prima fatti su StartMs su questa
//struttura
//var ts_bitfinex_transactions = map[string]int64{}
var lastLotBTFV2 int64
var lastLotBTFV2Symbol string
//var bookmapBTFV2 map[string]models.AggregateBooks
var bookmapBTFV2 *syncmap.Map
var waitSendBookBTFV2 bool
var lastTradeAlignTSBTFV2 string

const POSITIVE_AMOUNT  = "1"
const NEGATIVE_AMOUNT = "0"

type BitfinexAdapter2 struct{
	AbstractAdapter
	StartMs int64
	initBook bool
	chanbook chan []models.AggregateBooks
}

func NewBitfinexAdapter2() AdapterInterface {
	return BitfinexAdapter2{initBook:false}
}

func subscribeBooksBTFV2(chanchannels [] chan []float64, synchs [] chan int, symbols []string, channel string, wssclient *websocket.Client) *websocket.Client{
	wg := sync.WaitGroup{}
	wg.Add(3) // 1. Info with version, 2. Subscription event, 3. 3 x data message

	ctx, cxl := context.WithTimeout(context.Background(), time.Second*500)
	defer cxl()

	for i := 0; i < len(symbols); i++ {
		id, err := wssclient.SubscribeBook(ctx, bitfinex.TradingPrefix+strings.ToUpper(symbols[i]), "P0", "F0", 25)
		if err != nil {
			log.Fatal(err)
		}

		log.Print(id)
	}

	return wssclient

}

func subscribeTradesBTFV2(chanchannels [] chan []float64, synchs [] chan int, symbols []string, channel string, wssclient *websocket.Client) *websocket.Client{

	wg := sync.WaitGroup{}
	wg.Add(3) // 1. Info with version, 2. Subscription event, 3. 3 x data message

	ctx, cxl := context.WithTimeout(context.Background(), time.Second*500)
	defer cxl()

	/*
	for i := 0; i < len(synchs); i++ {
		<- synchs[i]
	}
	*/

	for i := 0; i < len(symbols); i++ {
		/*if channel == "book" {
			s := bitfinexcustom.SubscribeToChannelData{
				Channel:   channel,
				Pair:      strings.ToUpper(symbols[i]),
				Chan:      chanchannels[i],
				Precision: "R0",
				Length:    "25",
			}
			wss.AddSubscribeFull(s)
		} else {
			wss.AddSubscribe(channel, strings.ToUpper(symbols[i]), chanchannels[i])
		}*/

		id, err := wssclient.SubscribeTrades(ctx, bitfinex.TradingPrefix+strings.ToUpper(symbols[i]))
		if err != nil {
			log.Fatal(err)
		}

		log.Print(id)
	}

	//wss.Subscribe()

	return wssclient
}
/*
func asyncExtractBTFV2(chanchannel chan []float64, synch chan int, symbol string, channel string, wss *bitfinex.WebSocketService){
	wss.AddSubscribe(channel, strings.ToUpper(symbol), chanchannel)

	<- synch

	wss.Subscribe()
}
*/
/*
func extractBTFV2(chanchannel chan []float64, synch chan int, symbol string, channel string) *bitfinex.WebSocketService{
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
*/
func waitTradesBTFV2(chantrade chan []models.Trade, chanchannel chan []float64, synch chan int, exchangeId string, symbol string, wssclient *websocket.Client){

	log.Print("Start waiting trades")
	//synch <- 1
	go func() {
		for {
			select {
			case msg := <-wssclient.Listen():
				if msg == nil {
					return
				}
				log.Printf("recv msg: %#v", msg)
				switch m := msg.(type) {
				case error:
					//errch <- msg.(error)
					log.Fatal(msg.(error))
				case *websocket.UnsubscribeEvent:
					//unsubs <- m
					log.Print(m)
				case *websocket.SubscribeEvent:
					//subs <- m
					log.Print(m)
				case *websocket.InfoEvent:
					//infos <- m
					log.Print(m)
				case *bitfinex.TradeExecutionUpdateSnapshot:
					//trades <- m
					log.Print(m)
				case *bitfinex.Trade:
					log.Print("Trade msg: ")
					log.Println(m)
					tid := strconv.FormatInt(m.ID, 10)
					model := []models.Trade{models.Trade{Exchange_id: exchangeId, Symbol: m.Pair[1: len(m.Pair)],
						Trade_ts: time.Unix(0, m.MTS * int64(time.Millisecond)), Amount: m.Amount, Price: m.Price,
						Tid: tid}}
					log.Print("model")
					log.Print(model)
					chantrade <- model
					log.Print("End trade")
				case *bitfinex.TradeExecutionUpdate:
					//trades <- m
					log.Print(m)
				case *bitfinex.TradeExecution:
					//trades <- m
					log.Print(m)
				case *bitfinex.TradeSnapshot:
					//trades <- m
					log.Print(m)
				default:
					log.Print("test recv: %#v", msg)
				}
			}
		}
	}()
	/*
	for {
		rawtrade := <-chanchannel

		//diff := 1
		tid := "0"
		if len(rawtrade) == 4 {
			//diff = 0
			tid = strconv.FormatFloat(rawtrade[0], 'G', 40, 64)
			chantrade <- []models.Trade{models.Trade{Exchange_id: exchangeId, Symbol: strings.ToUpper(symbol),
				Trade_ts: time.Unix(int64(rawtrade[1 /*- diff*//*]), 0), Amount: rawtrade[3 /*- diff*//*], Price: rawtrade[2 /*- diff*//*],
				Tid: tid}}
		}

	}*/
}

func getMapKey(book models.AggregateBooks) string{
	var str bytes.Buffer
	str.WriteString(book.Symbol)
	str.WriteString(fmt.Sprint(book.Price))
	str.WriteString(utils.Ternary(book.Amount > 0, POSITIVE_AMOUNT, NEGATIVE_AMOUNT))

	return str.String()
}

func convertToAggregateBooks(exchangeId string, bookUpdate *bitfinex.BookUpdate) models.AggregateBooks{
	obsolete := bookUpdate.Action == bitfinex.BookRemoveEntry
	amount := utils.TernaryFloat64(bookUpdate.Side == bitfinex.Ask, bookUpdate.Amount, -bookUpdate.Amount)
	symbol := strings.ToUpper(bookUpdate.Symbol[1:len(bookUpdate.Symbol)])
	ret := models.AggregateBooks{Exchange_id: exchangeId, Symbol: symbol, Price: bookUpdate.Price, Count_number: float64(bookUpdate.Count), Amount: amount, Lot: lastLotBTFV2, Obsolete: obsolete}

	return ret
}

func waitBooksBTFV2(chanbook chan []models.AggregateBooks, chanchannel chan []float64, synch chan int, reset [] chan int, exchangeId string, symbol string, wssclient *websocket.Client){
	//synch <- 1

	errch := make(chan error)
	go func() {
		for {
			select {
			case msg := <-wssclient.Listen():
				if msg == nil {
					return
				}
				log.Printf("recv msg: %#v", msg)
				retChanbooks := []models.AggregateBooks{}
				switch m := msg.(type) {
				case error:
					errch <- msg.(error)
				case *websocket.UnsubscribeEvent:
					//unsubs <- m
				case *websocket.SubscribeEvent:
					//subs <- m
				case *websocket.InfoEvent:
					//infos <- m
				case *bitfinex.BookUpdateSnapshot:
					//books <- m
					for i := 0; i < len(m.Snapshot); i++{
						log.Print("---------------------", m.Snapshot[i])
						data := m.Snapshot[i]

						aggregateBook := convertToAggregateBooks(exchangeId, data)
						bookmapBTFV2.Store(getMapKey(aggregateBook), aggregateBook)
						retChanbooks = append(retChanbooks, aggregateBook)
					}

					chanbook <- retChanbooks
					lastLotBTFV2++
				case *bitfinex.BookUpdate:
					//books <- m
					log.Print(m)

					aggregateBook := convertToAggregateBooks(exchangeId, m)


					aggregateBook.Lot = lastLotBTFV2
					retChanbooks = append(retChanbooks, aggregateBook)
					chanbook <- retChanbooks
					lastLotBTFV2++
				default:
					log.Print("test recv: %#v", msg)
				}
			}
		}
	}()
/*/*
	for {
		rawbook := <-chanchannel

		if (len(rawbook) != 3) {
			//println("pluto: " + symbol)
			lastLotBTFV2++

			retChanbooks := []models.AggregateBooks{}

			bookmapBTFV2.Range(func(ki, vi interface{}) bool{
			/*for k,_ := range bookmapBTFV2{*/
			/*/*
				_, v := ki.(float64), vi.(models.AggregateBooks)
				//v := bookmapBTFV2[k]
				v.Lot = lastLotBTFV2
				retChanbooks = append(retChanbooks, v)

				return true
			})

			//print("lunghezza array: ")
			//println(len(retChanbooks))

			if retChanbooks != nil && len(retChanbooks) > 0{
				chanbook <- retChanbooks
			}
			//chanbooks <- []models.AggregateBook{models.AggregateBook{Exchange_id: exchangeId, Symbol: strings.ToUpper(symbol), Price: 0, Count_number: 0, Amount: rawbook[0], Lot: lastLotBTFV2, Obsolete: true}}
		} else {
			aggregateBook := models.AggregateBooks{Exchange_id: exchangeId, Symbol: strings.ToUpper(symbol), Price: rawbook[1], Count_number: 1, Amount: rawbook[2], Lot: lastLotBTFV2, Obsolete: false}
			//key := strconv.FormatFloat(rawbook[0], 'G', 40, 64)

			if aggregateBook.Price == 0 {
				if !waitSendBookBTFV2 {
					bookmapBTFV2.Delete(rawbook[0])

					//delete(bookmapBTFV2, key)
				}
			} else {
				if !waitSendBookBTFV2 {
					bookmapBTFV2.Store(rawbook[0], aggregateBook)
					//bookmapBTFV2[key] = aggregateBook
				}
			}
		}

	}
			*/
}

/*
func closeWss(wss *bitfinex.WebSocketService, closed bool){
	if !closed{
		wss.ClearSubscriptions()
		wss.Close()
	}
}
*/
func (ba BitfinexAdapter2) instantiateBooks(symbols []string, chanbook chan []models.AggregateBooks, chanchannels []chan []float64, reset [] chan int){

	//for {
		waitSendBookBTFV2= false
		var synchs [] chan int
		print("1:extract-open lot:")
		println(lastLotBTFV2)

		for i := 0; i < len(symbols); i++ {
			synchs = append(synchs, make(chan int))
		}

		wssclient := websocket.New()
		err := wssclient.Connect()
		if err != nil {
			log.Fatal("Error connecting to web socket : ", err)
		}

		//for i := 0; i < len(symbols); i++ {
			ba.waitBooksBTFV2(symbols[0], chanbook, chanchannels[0], synchs[0], reset, 0, wssclient)

		//}

		wssclient = subscribeBooksBTFV2(chanchannels, synchs, symbols, "book", wssclient)
		/*
		for i := 0; i < len(reset); i++ {
			<-reset[i]
		}
		*/
		/*time.Sleep(10000 * time.Millisecond)
		println("2:extract-closing ")
		//wss.ClearSubscriptions()
		//wss.Close()
		print("3:extractclosed ")
		println(lastLotBTFV2)

		waitSendBookBTFV2 = true
		//bookmapBTFV2 = make(map[string]models.AggregateBooks)
		//bookmapBTFV2 = new(syncmap.Map)

		print("Size of Map")
		size := 0
		bookmapBTFV2.Range(func(key interface{}, value interface{}) bool {
			size++
			return true
		})
		println(size)

		bookmapBTFV2.Range(func(key interface{}, value interface{}) bool {
			bookmapBTFV2.Delete(key)
			return true
		})

		size = 0
		bookmapBTFV2.Range(func(key interface{}, value interface{}) bool {
			size++
			return true
		})
		print("Size after delete")
		println(size)
	}*/
}

func (ba BitfinexAdapter2) waitBooksBTFV2(symbol string, chanbook chan []models.AggregateBooks, chanchannel chan []float64, synch chan int, reset [] chan int, i int, wssclient *websocket.Client){
	waitBooksBTFV2(chanbook, chanchannel, synch, reset, ba.ExchangeId, symbol, wssclient)
}

func (ba BitfinexAdapter2) getTrade() [] chan []models.Trade {

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

	wssclient := websocket.New()
	err := wssclient.Connect()
	if err != nil {
		log.Fatal("Error connecting to web socket : ", err)
	}
	//defer wssclient.Close()

	//for i := 0; i < len(synchs); i++ {
		//extract(chanchannels[i], synchs[i], symbols[i], "trades")
		waitTradesBTFV2(chantrades[0], chanchannels[0], synchs[0], ba.ExchangeId, symbols[0], wssclient)
	//}

	subscribeTradesBTFV2(chanchannels, synchs, symbols, "trades", wssclient)

	return chantrades
}

func (ba BitfinexAdapter2) getAggregateBooks() (chan []models.AggregateBooks, chan int) {

	lastLotBTFV2 = 1//datastorage.GetLastLot(ba.ExchangeId, ba.Symbol)
	//log.Println(ba.Symbol, lastLotBTFV2)
	lastLotBTFV2++
	//log.Println(ba.Symbol , lastLotBTFV2)

	symbols := datastorage.GetMarkets(BITFINEX)
	var chanchannels [] chan []float64
	var reset [] chan int
	outReset := make(chan int)

	if !ba.initBook {
		ba.chanbook = make(chan []models.AggregateBooks)
	}

	for i := 0; i < len(symbols); i++ {
		chanchannels = append(chanchannels, make(chan []float64))
		reset = append(reset, make(chan int))
	}

	//bookmapBTFV2 = make(map[string]models.AggregateBooks)
	bookmapBTFV2 = new(syncmap.Map)

	ba.initBook = true

	//for i := 0; i < len(synchs); i++ {
	go ba.instantiateBooks(symbols, ba.chanbook, chanchannels, reset)
	//}


	return ba.chanbook, outReset
}

func (ba BitfinexAdapter2) instantiateDefault(symbol string) AdapterInterface {
	ba.ExchangeId = BITFINEX
	aa := ba.abstractInstantiateDefault(symbol)
	ba.AbstractAdapter = aa
	return ba
}

func (ba BitfinexAdapter2) instantiate(Symbol string, FetchSize int, ReloadInterval int) AdapterInterface {
	ba.ExchangeId = BITFINEX
	aa := ba.abstractInstantiate(Symbol, FetchSize, ReloadInterval)
	ba.AbstractAdapter = aa
	return ba
}
/*
func GetTradesFromTS() (*bitfinex.TradeSnapshot, error){
	s := rest.NewClient();
	symbol := "tBTCUSD"
	req := rest.NewRequestWithDataMethod(path.Join("trades", symbol, "hist", "?limit="), map[string]interface{}{}, "GET")

	raw, err := s.Request(req)

	if err != nil {
		return nil, err
	}

	dat := make([][]float64, 0)
	for _, r := range raw {
		if f, ok := r.([]float64); ok {
			dat = append(dat, f)
		}
	}

	os, err := bitfinex.NewTradeSnapshotFromRaw(symbol, dat)
	if err != nil {
		return nil, err
	}
	return os, nil

}
*/
func (ba BitfinexAdapter2) getAlignTrades(symbol string, start string, end string, limit int) ([]models.Trade){
	if limit <= 0 {
		limit = 1000
	}
/*
	var startInt interface{};
	var endInt interface{};
	endInt = nil
	if start == ""{
		startInt = nil
	} else {
		var err interface{}
		startInt, err = strconv.ParseInt(start, 10, 64)

		if err != nil {
			startInt = nil
		}
	}
*/

	//start = utils.Ternary(start != "", start, lastTradeAlignTSBTFV2)

	s := rest.NewClient()
	symbol = "t"  + symbol

	path := path.Join("trades", symbol, "hist")
	path += "?start=" + start + "&end=" + end + "&limit=" + strconv.Itoa(limit) + utils.Ternary(start != "", "&sort=1", "")
	print("path=", path)
	req := rest.NewRequestWithDataMethod(path, map[string]interface{}{}, "GET")

	raw, err := s.Request(req)

	if err != nil {
		println("Error on retrieving trades from Bitfinex2")
		return nil
	}

	dat := make([][]float64, 0)
	for _, r := range raw {
		rcast := r.([]interface{})
		row := make([]float64, 0)
		for _, s := range rcast{
			val := s.(float64)
			row = append(row, val)
		}

		dat = append(dat, row)
		// print(r.(type))
	}

	os, err := bitfinex.NewTradeSnapshotFromRaw(symbol, dat)
	if err != nil {
		println("Error on retrieving trades from Bitfinex2")
		return nil
	}

	trades := os.Snapshot
	var retTrades = []models.Trade{}
	for _, el := range trades{
		tid := strconv.FormatInt(el.ID, 10)
		retTrade := models.Trade{Exchange_id: ba.ExchangeId, Symbol: el.Pair[1: len(el.Pair)],
			Trade_ts: time.Unix(0, el.MTS * int64(time.Millisecond)), Amount: el.Amount, Price: el.Price,
			Tid: tid}
		retTrades = append(retTrades, retTrade)
	}
	return retTrades

}


func CheckBitfinexRecordBTFV2 (symbol string, time_last time.Time, quantity float64) bool {
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

type Request struct {
	RefURL  string                 // ref url
	Data    map[string]interface{} // body data
	Method  string                 // http method
	Params  url.Values             // query parameters
	Headers map[string]string
}

func (ba BitfinexAdapter2) executeArbitrage(arbitrage models.Arbitrage) bool  {
	/*ac := properties.GetInstance()
	// case buy
	if ac.Bitfinex.ExecArbitrage == "S" {
		fmt.Println("attivata funzione arbitraggio")
		//historicalArbitrage := models.HistoricalArbitrage{Exchange_id:BITFINEX, SymbolStart: arbitrage.SymbolStart, SymbolTransitory: arbitrage.SymbolTransitory, SymbolEnd: arbitrage.SymbolEnd  }

		//ottengo application context

		client := bitfinex.NewClient().Auth(ac.Bitfinex.Key, ac.Bitfinex.Secret)

		valueTrade := arbitrage.PriceStart * arbitrage.AmountStart

		//condition to test limit valuo for the trading
		if valueTrade < (20 * (1 + 0.006)) {
			fmt.Println("funzione arbitraggio - valore vendita troppo basso")
			return false
		}

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

			fmt.Println("funzione arbitraggio - step 2 da scambiare: ", arbitrage.AmountTransitory * ratio * 0.99785)
			order1, err := client.Orders.Create(arbitrage.SymbolTransitory, arbitrage.AmountTransitory * ratio * 0.99785, 3, bitfinex.OrderTypeExchangeMarket)
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

	*/return true
}


/*
func GetAvailableQuantityBTFV2(symbol string, client *bitfinex.Client, buy bool) float64  {
	// buy have to be true if you need to buy
	currency1 := strings.ToUpper(symbol[0:3])
	currency2 := strings.ToUpper(symbol[3:6])

	/*fmt.Println(currency1)
	fmt.Println(currency2)*/
	//info, err := client.Account.Info()
	/*
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
*/
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