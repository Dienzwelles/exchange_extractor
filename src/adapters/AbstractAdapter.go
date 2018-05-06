package adapters

import models "../models"

type AdapterInterface interface {
	instantiate(symbol string, fetchSize int, reloadInterval int) AdapterInterface
	instantiateDefault(symbol string) AdapterInterface
	getTrade() [] chan []models.Trade
	getAggregateBooks() (chan []models.AggregateBooks, chan int)
	executeArbitrage(arbitrage models.Arbitrage) bool
}

type AbstractAdapter struct{
	ExchangeId string
	Symbol string
	FetchSize int
	ReloadInterval int
}


func (aa AbstractAdapter) abstractInstantiateDefault(symbol string) AbstractAdapter {
	return aa.abstractInstantiate(symbol, -1, 0)
}

func (aa AbstractAdapter) abstractInstantiate(Symbol string, FetchSize int, ReloadInterval int) AbstractAdapter {
	aa.Symbol = Symbol
	aa.FetchSize = FetchSize
	aa.ReloadInterval = ReloadInterval

	return aa;
}
