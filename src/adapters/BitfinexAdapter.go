package adapters

import (
	"../models"
	"strconv"
	"time"
	"net/http"
	"log"
	"io/ioutil"
	"encoding/json"
	"../datastorage"
	_ "fmt"
	"strings"
	_"github.com/shopspring/decimal"
	_"github.com/thrasher-/gocryptotrader/exchanges/bitfinex"
	"github.com/bitfinexcom/bitfinex-api-go/v1"
	"fmt"
	"../properties"
)


const BITFINEX  = "Bitfinex"


//contiene timestamp ultima richiesta fatta per symbol
//sposto i ragionamenti prima fatti su StartMs su questa
//struttura
var ts_bitfinex_transactions = map[string]int64{}
var lastLot int64


type BitfinexAdapter struct{
	AbstractAdapter
	StartMs int64
}

func NewBitfinexAdapter() AdapterInterface {
	return BitfinexAdapter{}
}

func (ba BitfinexAdapter) getTrade() []models.Trade {

	//movimenti che dovranno essere ritornati
	trd := []models.Trade{}

	//ritorna la lista dei mercati attualmente attivi
	symbols := datastorage.GetMarkets(BITFINEX)

	var url string

	// Per ogni mercato attivo recupera ultimi movimenti
	for _, symbol := range symbols {
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

		var trades= make([]models.Trade, len(rawtrades))

		for i := 0; i < len(rawtrades); i++ {
			rawtrade := rawtrades[i]

			ts_bitfinex_transactions[symbol] = int64(rawtrade[1])
			trades[i] = models.Trade{Exchange_id: ba.ExchangeId, Symbol: strings.ToUpper(symbol), Trade_ts: time.Unix(int64(rawtrade[1]/1000), 0), Amount: rawtrade[2], Price: rawtrade[3]}
		}

		for _, trade := range trades{

			//se il record è nuovo allora lo inserisco
			if CheckBitfinexRecord(trade.Symbol, trade.Trade_ts , trade.Amount) {
				trd = AddItem(trd, trade)
			}
		}

		log.Print("Got ", len(trades), " trades")

	}

	return trd
}

func (ba BitfinexAdapter) getAggregateBooks() []models.AggregateBook {
	lastLot = datastorage.GetLastLot(ba.ExchangeId, ba.Symbol)
	log.Println(ba.Symbol, lastLot)
	lastLot++
	log.Println(ba.Symbol , lastLot)


	var url string
	url = "https://api.bitfinex.com/v2/book/t" + strings.ToUpper(ba.Symbol) + "/P0"

	log.Println(url)


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

	var rawbooks [][3]float64
	jsonErr := json.Unmarshal(body,  &rawbooks)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	log.Println(rawbooks)

	books := []models.AggregateBook{}
	for _, rawbook := range rawbooks{
		book := models.AggregateBook{Exchange_id: ba.ExchangeId, Symbol: strings.ToUpper(ba.Symbol), Price: rawbook[0], Count_number: rawbook[1], Amount: rawbook[2],Lot: lastLot, Obsolete: false}
		fmt.Println(book)
		books = append(books, book)
	}

	return books
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