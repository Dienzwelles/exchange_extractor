package exchanges

import "../datastorage"

func Instantiate() {

	//bittrex market
	var a MarketInterface
	a = NewBittrexMarket().instantiateDefault()
	//get market list
	markets := a.getMarkets()

	//bitfinex market
	var b MarketInterface
	b = NewBitfinexMarket().instantiateDefault()

	markets = append(markets, b.getMarkets()...)
	//markets := b.getMarkets()

	//store market code
	datastorage.StoreMarkets(markets)

}

