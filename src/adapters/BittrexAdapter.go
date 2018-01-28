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
var ts_last_transaction time.Time

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

	//url = "https://bittrex.com/api/v1.1/public/getmarkethistory?market=BTC-DOGE"
	trd := []models.Trade{}

	// Bittrex client
	bittrex := bittrex.New(API_KEY, API_SECRET)

	// Market history
	trade_symbol := ba.Symbol

	//get trades
	marketHistory, err := bittrex.GetMarketHistory(trade_symbol)
	max_ts := ts_last_transaction
	//read single trade and stored on local structure
	for _, trade := range marketHistory {
		fmt.Println(err, trade.Timestamp.String(), trade.Quantity, trade.Price, trade.Total, trade.OrderType)
		if !trade.Timestamp.Time.After(max_ts) {
			continue
		}
		if trade.Timestamp.Time.After(ts_last_transaction) {
			ts_last_transaction=trade.Timestamp.Time
		}
		q, _ := strconv.ParseFloat(trade.Quantity.String(), 64)
		p, _ := strconv.ParseFloat(trade.Price.String(), 64)

		a :=  models.Trade{Exchange_id: ba.ExchangeId , Symbol: trade_symbol, Trade_ts: trade.Timestamp.Time , Amount: q , Price: p , Rate: 0, Period: 0 }

		trd = AddItem(trd, a)
	}

	return trd
}

func (ba BittrexAdapter) instantiateDefault(symbol string) AdapterInterface {
	ba.ExchangeId = BITTREX
	ts_last_transaction = GetLastID(symbol)
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