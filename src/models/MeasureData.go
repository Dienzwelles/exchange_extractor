package models

type MeasureData struct {
	Slow float64
	Medium float64
	High float64
}

type TickData struct {
	Momentum float64
	Ticks MeasureData
}

type MeasuresData struct {
	ExchangeId string
	Symbol string
	Momentum float64
	Ticks MeasureData
	//Price MeasureData
	Trades MeasureData
	NegativeTrades MeasureData
	PositiveTrades MeasureData
	AbsAmount MeasureData
	AmountOnAbs MeasureData
	AmountOnTrade MeasureData
}
