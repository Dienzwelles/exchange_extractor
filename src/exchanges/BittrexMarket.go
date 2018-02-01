package exchanges

import (
	_ "net/http"
	_ "log"
	_ "io/ioutil"
	_ "encoding/json"
	_ "fmt"
	_ "github.com/shopspring/decimal"
	"github.com/toorop/go-bittrex"
	_ "time"
	_ "github.com/jinzhu/gorm"

	"../models"
	"../datastorage"

	"log"
)




const (
	API_KEY    = ""
	API_SECRET = ""
)

type BittrexMarket struct{
	AbstractMarket

}

func NewBittrexMarket() MarketInterface {
	return BittrexMarket{}
}



func (ba BittrexMarket) getMarkets() []models.Market {

	//url = "https://bittrex.com/api/v1.1/public/getmarkethistory?market=BTC-DOGE"
	exc := []models.Market{}

	// Bittrex client
	bittrex := bittrex.New(API_KEY, API_SECRET)


	//get trades
	markets, _ := bittrex.GetMarkets()

	for _, market := range markets {
		//fmt.Println(err, market)

		a :=  models.Market{Exchange_id: ba.ExchangeId , Symbol: market.MarketName, Evaluated: false}

		exc = append(exc,a)
	}

	return exc
}

func (ba BittrexMarket) instantiateDefault() MarketInterface {
	ba.ExchangeId = "Bittrex"
	return ba
}


func storeExchanges(markets []models.Market) {



	conn := datastorage.NewConnection()
	db := datastorage.GetConnectionORM(conn)

	defer db.Close()

	for _, market := range markets {
		//fmt.Println(market)
		res2 := db.NewRecord(market)
		rows, _ := db.Table("markets").Select("exchange_id, symbol").Where("exchange_id = ? and symbol = ? ", market.Exchange_id, market.Symbol).Rows()
		count := 0
		for rows.Next() {
			count++
		}

		if  count  == 0  {
			dbe := db.Create(&market)
			if res2{
				log.Print("insert new exchange market")
			}
			if dbe.Error != nil{
				panic(dbe.Error)
			}
		}

	}

}