package exchanges

import "../models"

type MarketInterface interface {
	instantiateDefault() MarketInterface
	getMarkets() []models.Market
}

type AbstractMarket struct{
	ExchangeId string
}
