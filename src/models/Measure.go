package models

import "time"

type Ticks struct {
	Id int `gorm:"AUTO_INCREMENT"`
	Exchange_id string
	Symbol string
	Tick_ts time.Time
	Momentum float64
	Ticks_slow float64
	Ticks_medium float64
	Ticks_high float64
}

type Measures struct {
	Id int `gorm:"AUTO_INCREMENT"`
	Exchange_id string
	Symbol string
	Measure_ts time.Time
	Trades_slow float64
	Trades_medium float64
	Trades_high float64
	NegativeTrades_slow float64
	NegativeTrades_medium float64
	NegativeTrades_high float64
	PositiveTrades_slow float64
	PositiveTrades_medium float64
	PostiveTrades_high float64
	AbsAmount_slow float64
	AbsAmount_medium float64
	AbsAmount_high float64
	AmountOnAbs_slow float64
	AmountOnAbs_medium float64
	AmountOnAbs_high float64
	AmountOnTrade_slow float64
	AmountOnTrade_medium float64
	AmountOnTrade_high float64
}
