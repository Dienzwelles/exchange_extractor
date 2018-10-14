package adapters


import (
	_ "log"
	_ "io/ioutil"
	_ "encoding/json"
	_ "fmt"
	_ "github.com/shopspring/decimal"
	//contiene esempi di connessione al server
	_ "github.com/nntaoli-project/GoEx"
	_ "time"
	_ "strconv"
	"../models"
	"time"
	"log"
	_ "errors"
	_ "net/url"
	_ "strings"
	_ "net/http"
	_ "io/ioutil"
	"../datastorage"
	"fmt"
	_ "strings"
	_ "github.com/thrasher-/gocryptotrader/config"
	"github.com/thrasher-/gocryptotrader/exchanges/okex"
	_ "strconv"
	_ "strconv"
)

const (
	OKEX 	= "Okex"
	apiKey    = ""
	apiSecret = ""
)

var o okex.OKEX


/*func (okFuture *OKEx) GetTrades(currencyPair CurrencyPair, since int64) ([]Trade, error) {

}
*/

//contiene l'ultimo timestamp archiviato
var okex_ts_last_transaction = map[string]time.Time{}
//var okex_ts_last_transaction  = time.Time{}

type OkexAdapter struct{
	AbstractAdapter
	StartMs int64
}

func NewOkexAdapter() AdapterInterface {
	return OkexAdapter{}
}



func OkexAddItem(trades []models.Trade, item models.Trade) []models.Trade {
	trades = append(trades, item)
	return trades
}

func (ba OkexAdapter) getTrade() [] chan []models.Trade {

	//trd := []models.Trade{}

	//ritorna la lista dei mercati attualmente attivi
	symbols := datastorage.GetMarkets(OKEX)


	// Per ogni mercato attivo recupera ultimi movimenti
	for _, symbol := range symbols {

		// recupero ultimo time stamp inserito per il mercato in esame
		okex_ts_last_transaction[symbol] = GetLastIDOkex(symbol)

		//marketHistory, err := o.GetSpotRecentTrades("ltc_btc", "0")
		/*marketHistory, err := o.GetSpotRecentTrades(nil)
		if err != nil {
			log.Fatal(err)
		}
		max_ts := okex_ts_last_transaction[symbol]
		_ = max_ts

		for _, trade := range marketHistory {

			a := time.Unix(int64(trade.Date), int64(trade.DateInMS))
			//fmt.Println(trade)
			//fmt.Println(a)
			if a.Before(max_ts) || a.Equal(max_ts) {
				continue
			}
			if a.After(okex_ts_last_transaction[symbol]) {
				okex_ts_last_transaction[symbol] = a
			}
			q := trade.Amount
			if trade.Type == "SELL" {
				q = -q
				fmt.Println(q)
			}
			b := models.Trade{Exchange_id: ba.ExchangeId, Symbol: "ltc_btc", Trade_ts: a, Amount: q, Price: trade.Price, Rate: 0, Period: 0}
			trd = AddItem(trd, b)

		}
		*/
	}

	return nil//trd







	// Per ogni mercato attivo recupera ultimi movimenti
	/*for _, symbol := range symbols {


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
	}*/
}


func (ba OkexAdapter) getAggregateBooks() (chan []models.AggregateBooks, chan int) {

	//movimenti che dovranno essere ritornati
	//bks := []models.AggregateBook{}
	return nil, nil//bks
}

func (ba OkexAdapter) instantiateDefault(symbol string) AdapterInterface {
	ba.ExchangeId = OKEX
	aa := ba.abstractInstantiateDefault(symbol)
	ba.AbstractAdapter = aa
	return ba
}

func (ba OkexAdapter) instantiate(Symbol string, FetchSize int, ReloadInterval int) AdapterInterface {
	ba.ExchangeId = OKEX
	aa := ba.abstractInstantiate(Symbol, FetchSize, ReloadInterval)
	ba.AbstractAdapter = aa
	return ba
}






func GetLastIDOkex(symbol string) time.Time {
	conn := datastorage.NewConnection()
	db := datastorage.GetConnection(conn)
	defer db.Close()

	var tm_st = time.Time{}
	var hasRow bool
	err := db.QueryRow("select IF(COUNT(*),'true','false') from trades where exchange_id = '" + OKEX+ "' and symbol = ?", symbol).Scan(&hasRow)
	if err != nil {
		log.Fatal(err)
	}
	if !hasRow{
		return tm_st
	}

	rows, err := db.Query("select max(trade_ts) as max_ts from trades where exchange_id = '" + OKEX + "' and symbol = ?", symbol)
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

func (ba OkexAdapter) executeArbitrage(arbitrage models.Arbitrage) bool  {
	return false
}