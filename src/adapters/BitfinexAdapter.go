package adapters

import (
	"../models"
	"strconv"
	"time"
	"net/http"
	"log"
	"io/ioutil"
	"encoding/json"
)

const BITFINEX  = "Bitfinex"

type BitfinexAdapter struct{
	*AbstractAdapter
	StartMs int64
}

func (ba *BitfinexAdapter) getData() []models.Trade {
	url := "https://api.bitfinex.com/v2/trades/t" + ba.Symbol + "/hist?sort=1"
	if ba.FetchSize > 0{
		url := url + "&limit=" + strconv.Itoa(ba.FetchSize)
	}

	if ba.StartMs != 0 {
		url := url + "&start=" + strconv.FormatInt(ba.StartMs, 0)
	}

	httpClient := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
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

	var trades []models.Trade

	for i := 0; i < len(rawtrades); i++ {
		rawtrade := rawtrades[i]

		ba.StartMs = int64(rawtrade[1])
		trades[i] = models.Trade{Exchange_id: ba.ExchangeId, Symbol: ba.Symbol, Trade_ts: time.Unix(int64(rawtrade[1]/1000), 0), Amount: rawtrade[2], Price: rawtrade[3]}
	}

	log.Print("Got %d trades", len(trades))
	return trades;
}

func (ba *BitfinexAdapter) abstractInstantiateDefault(symbol string) *AbstractAdapter {
	ba.ExchangeId = BITFINEX
	return ba.abstractInstantiateDefault(symbol)
}

func (ba *BitfinexAdapter) abstractInstantiate(Symbol string, FetchSize int, ReloadInterval int) *AbstractAdapter {
	ba.ExchangeId = BITFINEX
	return ba.abstractInstantiate(Symbol, FetchSize, ReloadInterval)
}
