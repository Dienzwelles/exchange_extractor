package exchanges


import (
	_ "net/http"
	_ "log"
	_ "io/ioutil"
	_ "encoding/json"
	_ "fmt"
	_ "github.com/shopspring/decimal"
	_ "time"
	_ "github.com/jinzhu/gorm"

	"../models"
	"../datastorage"

	"log"
	_ "github.com/thrasher-/gocryptotrader/exchanges/okex"
	_ "github.com/thrasher-/gocryptotrader/config"

)



const (
	OKEX 	= "Okex"
	apiKey    = ""
	apiSecret = ""
)


type OkexMarket struct{
	AbstractMarket

}

func NewOkexMarket() MarketInterface {
	return BittrexMarket{}
}



func (ba OkexMarket) getMarkets() []models.Market {

	//url = "https://bittrex.com/api/v1.1/public/getmarkethistory?market=BTC-DOGE"
	exc := []models.Market{}


	return exc
}

func (ba OkexMarket) instantiateDefault() MarketInterface {
	ba.ExchangeId = "Okex"
	return ba
}


func storeOkexExchanges(markets []models.Market) {



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
