package models

import "time"

type Arbitrage struct {
	SymbolStart string
	SymbolTransitory string
	SymbolEnd string
	AmountStart float64
	AmountTransitory float64
	AmountEnd float64
	PriceStart float64
	PriceTransitory float64
	PriceEnd float64
}

type HistoricalArbitrage struct {
	Id int `gorm:"AUTO_INCREMENT"`
	Exchange_id string
	SymbolStart string
	TypeStart string
	TimeStart time.Time
	TidStart string
	SymbolTransitory string
	TypeTransitory string
	TimeTransitory time.Time
	TidTransitory string
	SymbolEnd string
	TypeEnd string
	TimeEnd time.Time
	TidEnd string
	AmountStart float64
	AmountTransitory float64
	AmountEnd float64
	PriceStart float64
	PriceTransitory float64
	PriceEnd float64
}