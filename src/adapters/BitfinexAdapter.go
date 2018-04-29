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
	chanbook chan []models.AggregateBook
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
			Length: "100",
		}
		wss.AddSubscribeFull(s)
		//wss.AddSubscribeFull(channel, strings.ToUpper(symbols[i]), "R0", "100", chanchannels[i])
	}

	for i := 0; i < len(synchs); i++ {
		<- synchs[i]
	}

	wss.Subscribe()
}

func extractAll(chanchannels [] chan []float64, synchs [] chan int, symbols []string, channel string) *bitfinexcustom.WebSocketService{
	ac := properties.GetInstance()
	client := bitfinex.NewClient().Auth(ac.Bitfinex.Key, ac.Bitfinex.Secret)

	trades, err:= client.Trades.All("BTCUSD", time.Now(), 50)

	if err != nil{
		panic(err)
	}

	print(trades)

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

func waitBooks(chanbook chan []models.AggregateBook, chanchannel chan []float64, synch chan int, reset [] chan int, exchangeId string, symbol string){
	synch <- 1

	inDiff := false

	for {
		rawbook := <-chanchannel

		if(len(rawbook) != 3){
			//println("pluto: " + symbol)
			lastLot++

			inDiff = true

			retChanbooks := []models.AggregateBook{}

			bookmap.Range(func(ki, vi interface{}) bool {
				_, v := ki.(float64), vi.(models.AggregateBook)

				v.Lot = lastLot
				retChanbooks = append(retChanbooks, v)

				return true
			})

			//print("lunghezza array: ")
			println(len(retChanbooks))
			chanbook <- retChanbooks
			//chanbooks <- []models.AggregateBook{models.AggregateBook{Exchange_id: exchangeId, Symbol: strings.ToUpper(symbol), Price: 0, Count_number: 0, Amount: rawbook[0], Lot: lastLot, Obsolete: true}}
		} else {
			aggregateBook := models.AggregateBook{Exchange_id: exchangeId, Symbol: strings.ToUpper(symbol), Price: rawbook[1], Count_number: 1, Amount: rawbook[2], Lot: lastLot, Obsolete: false}

			if aggregateBook.Price == 0 && inDiff{
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
func (ba BitfinexAdapter) instantiateExtracts(symbols []string, chanbook chan []models.AggregateBook, chanchannels []chan []float64, synchs []chan int, reset [] chan int){
	for {
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
	}
}

func (ba BitfinexAdapter) instantiateExtract(symbol string, chanbook chan []models.AggregateBook, chanchannel chan []float64, synch chan int, reset [] chan int, i int){
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

func (ba BitfinexAdapter) getAggregateBooks() (chan []models.AggregateBook, chan int) {

	lastLot = datastorage.GetLastLot(ba.ExchangeId, ba.Symbol)
	log.Println(ba.Symbol, lastLot)
	lastLot++
	log.Println(ba.Symbol , lastLot)

	symbols := datastorage.GetMarkets(BITFINEX)
	var chanchannels [] chan []float64
	var synchs [] chan int
	var reset [] chan int
	outReset := make(chan int)

	if !ba.initBook {
		ba.chanbook = make(chan []models.AggregateBook)
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
	//ottengo application context
	ac := properties.GetInstance()
	fmt.Println(ac.Bitfinex.Key)
	fmt.Println(ac.Bitfinex.Secret)
	fmt.Println(ac.Bitfinex.ExecArbitrage)
	client := bitfinex.NewClient().Auth(ac.Bitfinex.Key, ac.Bitfinex.Secret)
	if (arbitrage.AmountStart >= 0) {
		// case sell
		if ac.Bitfinex.ExecArbitrage == "S" {
			fmt.Println("funzione arbitraggio - esecuzione trade")
			info, err := client.Account.Info()
			if err != nil {
				fmt.Println(err)
				fmt.Println("funzione arbitraggio - errore esecuzione trade")
			} else {
				fmt.Println(info)
			}
		}

		conn := datastorage.NewConnection()
		db := datastorage.GetConnectionORM(conn)
		//db.LogMode(true)
		defer db.Close()
		historicalArbitrage := models.HistoricalArbitrage{Exchange_id:BITFINEX, SymbolStart: arbitrage.SymbolStart, SymbolTransitory: arbitrage.SymbolTransitory, SymbolEnd: arbitrage.SymbolEnd}

		res2 := db.NewRecord(historicalArbitrage)
		dbe := db.Create(&historicalArbitrage)

		if res2{
			log.Print("insert new historical arbitrage")
		}

		if dbe.Error != nil{
			panic(dbe.Error)
		}

	} else{
		//caso di vendita
	}



	return false
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