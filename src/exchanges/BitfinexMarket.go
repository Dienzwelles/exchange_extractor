package exchanges

import (
	_ "net/http"
	_ "log"
	_ "io/ioutil"
	_ "encoding/json"
	"fmt"
	_ "github.com/shopspring/decimal"
	_ "time"
	_ "github.com/jinzhu/gorm"

	"../models"

	"log"
	"net/http"
	"io/ioutil"
	"time"
	"encoding/json"
)




type BitfinexMarket struct{
	AbstractMarket

}

func NewBitfinexMarket() MarketInterface {
	return BitfinexMarket{}
}



func (ba BitfinexMarket) getMarkets() []models.Market {


	var url string
	url = "https://api.bitfinex.com/v1/symbols"

	req, err := http.NewRequest(http.MethodGet, url, nil)



	exc := []models.Market{}

	if err != nil {
		log.Fatal(err)
	}
	httpClient := http.Client{
		Timeout: time.Second * 10, // Maximum of 2 secs
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

	var markets []string
	jsonErr := json.Unmarshal(body, &markets)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}


	//get trades
	//markets, err := bittrex.GetMarkets()

	for _, market := range markets {
		fmt.Println(err, market)



		a :=  models.Market{Exchange_id: ba.ExchangeId , Symbol: market, Evaluated: false}

		exc = append(exc,a)
	}

	return exc
}

func (ba BitfinexMarket) instantiateDefault() MarketInterface {
	ba.ExchangeId = "Bitfinex"
	return ba
}

