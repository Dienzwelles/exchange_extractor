package models

import "time"

type Book struct {
	Id int `gorm:"AUTO_INCREMENT"`
	Exchange_id string
	Symbol string
	Type string
	Book_ts time.Time
	Amount float64
	Price float64

}

type AggregateBook struct {
	Id int `gorm:"AUTO_INCREMENT"`
	Exchange_id string
	Symbol string
	Price float64
	Count float64
	Amount float64
}