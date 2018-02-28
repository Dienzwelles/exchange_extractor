package adapters

import (
	_ "net/http"
	_ "log"
	_ "io/ioutil"
	_ "encoding/json"
	"fmt"
	_ "github.com/shopspring/decimal"
	"github.com/toorop/go-bittrex"
	_ "time"
	"strconv"
	"../models"
	"../datastorage"
	"time"
	"log"
)

const BITTREX  = "Bittrex"

//contiene l'ultimo timestamp archiviato
var ts_last_transaction = map[string]time.Time{}

//parametri base per chiamare api bitrex
const (
	API_KEY    = ""
	API_SECRET = ""
)

type BittrexAdapter struct{
	AbstractAdapter
	StartMs int64
}

func NewBittrexAdapter() AdapterInterface {
	return BittrexAdapter{}
}



func AddItem(trades []models.Trade, item models.Trade) []models.Trade {
	trades = append(trades, item)
	return trades
}

func (ba BittrexAdapter) getTrade() []models.Trade {

	// esempio chiamata base
	// url = "https://bittrex.com/api/v1.1/public/getmarkethistory?market=BTC-DOGE"
	trd := []models.Trade{}

	//ritorna la lista dei mercati attualmente attivi
	symbols := datastorage.GetMarkets(BITTREX)

	// Bittrex client
	bittrex := bittrex.New(API_KEY, API_SECRET)

	// Per ogni mercato attivo recupera ultimi movimenti
	for _, symbol := range symbols {

		// Market history
		/*if value, ok := ts_last_transaction[symbol]; ok {
			fmt.Println("value: ", value)
		} else {
			fmt.Println("key not found")
		}*/

		// recupero ultimo time stamp inserito per il mercato in esame
		ts_last_transaction[symbol] = GetLastID(symbol)

		//recupero ultimi scambi
		marketHistory, err := bittrex.GetMarketHistory(symbol)
		max_ts := ts_last_transaction[symbol]

		//leggo i singoli scambi e li storicizzo in un apposita struttura
		for _, trade := range marketHistory {
			if trade.Timestamp.Time.Before(max_ts) || trade.Timestamp.Time.Equal(max_ts) {
				continue
			}
			if trade.Timestamp.Time.After(ts_last_transaction[symbol]) {
				ts_last_transaction[symbol] = trade.Timestamp.Time
			}
			q, _ := strconv.ParseFloat(trade.Quantity.String(), 64)
			if trade.OrderType == "SELL"{
				q = -q
				fmt.Println(q)
			}
			p, _ := strconv.ParseFloat(trade.Price.String(), 64)

			a := models.Trade{Exchange_id: ba.ExchangeId, Symbol: symbol, Trade_ts: trade.Timestamp.Time, Amount: q, Price: p, Rate: 0, Period: 0}

			//se il record è nuovo allora lo inserisco
			if CheckRecord(a.Symbol, max_ts , a.Amount) {
				fmt.Println(err, trade.Timestamp.String(), trade.Quantity, trade.Price, trade.Total, trade.OrderType)
				trd = AddItem(trd, a)
			}
		}
	}
	return trd
}

func (ba BittrexAdapter) getAggregateBooks() []models.AggregateBook {

	//movimenti che dovranno essere ritornati
	bks := []models.AggregateBook{}
	return bks
}


func (ba BittrexAdapter) instantiateDefault(symbol string) AdapterInterface {
	ba.ExchangeId = BITTREX
	aa := ba.abstractInstantiateDefault(symbol)
	ba.AbstractAdapter = aa
	return ba
}

func (ba BittrexAdapter) instantiate(Symbol string, FetchSize int, ReloadInterval int) AdapterInterface {
	ba.ExchangeId = BITTREX
	aa := ba.abstractInstantiate(Symbol, FetchSize, ReloadInterval)
	ba.AbstractAdapter = aa
	return ba
}

func GetLastID(symbol string) time.Time {
	conn := datastorage.NewConnection()
	db := datastorage.GetConnection(conn)
	defer db.Close()

	var tm_st = time.Time{}
	var hasRow bool
	err := db.QueryRow("select IF(COUNT(*),'true','false') from trades where exchange_id = '" + BITTREX + "' and symbol = ?", symbol).Scan(&hasRow)
	if err != nil {
		log.Fatal(err)
	}
	if !hasRow{
		return tm_st
	}

	rows, err := db.Query("select max(trade_ts) as max_ts from trades where exchange_id = '" + BITTREX + "' and symbol = ?", symbol)
	if err!=nil{
		log.Fatal(err)
	}

	defer rows.Close()
	for rows.Next(){
		err:=rows.Scan(&tm_st)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(tm_st)
	}

	return tm_st
}

func CheckRecord (symbol string, time_last time.Time, quantity float64) bool {
	conn := datastorage.NewConnection()
	db := datastorage.GetConnection(conn)
	defer db.Close()

	type el struct {
		sy string
		ts time.Time
		am float64
	}

	var record el

	rows, err := db.Query("select symbol, trade_ts, amount from trades where exchange_id = '" + BITTREX + "' and symbol = ? and trade_ts = ? group by symbol, trade_ts, amount ", symbol, time_last)
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