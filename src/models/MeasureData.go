package models

type MeasureData struct {
	Slow float64
	Medium float64
	High float64
}

type MeasuresData struct {
	ExchangeId string
	Symbol string
	Ticks MeasureData
	Price MeasureData
	Trades MeasureData
	NegativeTrades MeasureData
	PositiveTrades MeasureData
	AbsAmount MeasureData
	AmountOnAbs MeasureData
	AmountOnTrade MeasureData
}
