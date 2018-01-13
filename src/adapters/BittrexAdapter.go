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
)

const BITTREX  = "Bittrex"

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

	//read single trade and stored on local structure
	for _, trade := range marketHistory {

		fmt.Println(err, trade.Timestamp.String(), trade.Quantity, trade.Price, trade.Total, trade.OrderType)
		q, _ := strconv.ParseFloat(trade.Quantity.String(), 64)
		p, _ := strconv.ParseFloat(trade.Price.String(), 64)

		a :=  models.Trade{Exchange_id: ba.ExchangeId , Symbol: trade_symbol, Trade_ts: trade.Timestamp.Time , Amount: q , Price: p , Rate: 0, Period: 0 }
		trd = AddItem(trd, a)
	}

	return trd
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
