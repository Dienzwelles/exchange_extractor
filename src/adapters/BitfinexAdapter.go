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
	historicalArbitrage := models.HistoricalArbitrage{Exchange_id:BITFINEX, SymbolStart: arbitrage.SymbolStart, SymbolTransitory: arbitrage.SymbolTransitory, SymbolEnd: arbitrage.SymbolEnd  }

	//ottengo application context
	ac := properties.GetInstance()
	client := bitfinex.NewClient().Auth(ac.Bitfinex.Key, ac.Bitfinex.Secret)
	if (arbitrage.AmountStart >= 0) {
		// case sell
		if ac.Bitfinex.ExecArbitrage == "S" {
			fmt.Println("funzione arbitraggio - esecuzione trade")
			//return the max available quantity within the wallet for the first cross side b
			q0 := GetAvailableQuantity(arbitrage.SymbolStart, client, true)
			//return initial available quantity for the first cross side a
			initalQuantity := GetAvailableQuantity(arbitrage.SymbolStart, client, false)

			//check if quantity is ok
			if arbitrage.AmountStart > q0 {
				fmt.Println("funzione arbitraggio - valore da scambiare eccede la dispomibilità massima" )
				return false
			}
			fmt.Println("funzione arbitraggio - step 0 valore da scambiare: ", arbitrage.AmountStart )
			//execute first trade
			order, err := client.Orders.Create(arbitrage.SymbolStart, arbitrage.AmountStart, 3, bitfinex.OrderTypeExchangeMarket)
			if err != nil {
				fmt.Println("errore durante acquisto 0")
				fmt.Println(err)
				return false
			} else {
				fmt.Println("acquisto 0 avvenuto")
				time.Sleep(1000 * time.Millisecond)
				fmt.Println(order)
			}
			//historicalArbitrage.TidStart = order.ID

			//return the max available quantity within the wallet for the second cross side b
			q1:= GetAvailableQuantity(arbitrage.SymbolTransitory, client, true)
			initalQuantity2 := GetAvailableQuantity(arbitrage.SymbolTransitory, client, false)
			fmt.Println("funzione arbitraggio - q1", q1)
			fmt.Println("funzione arbitraggio - initalQuantity2", initalQuantity2)

			tick, err := client.Ticker.Get(arbitrage.SymbolTransitory)

			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("ticker %s  %s", arbitrage.SymbolTransitory, tick.Ask )
				fmt.Println(tick.Ask)
			}
			divisore, _ := strconv.ParseFloat(tick.Ask, 64)
			amount := (q1 - initalQuantity)/divisore
			fmt.Println("funzione arbitraggio - step 1 da scambiare: ", amount )
			//execute second trade
			order1, err := client.Orders.Create(arbitrage.SymbolTransitory, amount , 3, bitfinex.OrderTypeExchangeMarket)
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
			q2:= GetAvailableQuantity(arbitrage.SymbolTransitory, client, false)
			//inverto q2 e initial quantity perchè devo vendere
			amount2 := initalQuantity2 - q2
			fmt.Println("funzione arbitraggio - step 2 valore da scambiare: ", amount2 )
			//execute second trade
			order2, err := client.Orders.Create(arbitrage.SymbolEnd, amount2 , 3, bitfinex.OrderTypeExchangeMarket)
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


		conn := datastorage.NewConnection()
		db := datastorage.GetConnectionORM(conn)
		//db.LogMode(true)
		defer db.Close()

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



	return true
}

func GetAvailableQuantity(symbol string, client *bitfinex.Client, buy bool) float64  {
	// buy have to be true if you need to buy
	currency1 := symbol[0:3]
	currency2 := symbol[3:6]

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
			if strings.Compare(balance.Currency, currency2) == 0 && buy {
				q, _ := strconv.ParseFloat(balance.Available, 64)
				fmt.Printf("funzione arbitraggio - quantità disponibile %s: %f\n",balance.Currency, q)
				return q
			}
			if strings.Compare(balance.Currency, currency1) == 0 && !buy {
				q, _ := strconv.ParseFloat(balance.Available, 64)
				fmt.Printf("funzione arbitraggio - quantità disponibile %s: %f\n",balance.Currency, q)
				return q
			}
		}
	}
	return 0
}