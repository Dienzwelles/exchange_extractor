package models

import "time"

type Trade struct {
	Id int `gorm:"AUTO_INCREMENT"`
	Exchange_id string
	Symbol string
	Trade_ts time.Time
	Amount float64
	Price float64
	Rate float64
	Period int
	Tid string
}

