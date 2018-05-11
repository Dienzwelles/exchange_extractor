package main

import (
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/go-sql-driver/mysql"
	"time"
	"./adapters"
	"./exchanges"
	"./services"

	"fmt"
)


type BitfinexTrade struct {
	MTS time.Time
	AMOUNT float64
	PRICE float64
	RATE float64
	PERIOD int
}


func main() {
	exchanges.Instantiate()
	//adapters.Instantiate()
	ch := make(chan bool)
	exitMain := make(chan bool)
	go func() {
		adapters.Instantiate(ch, exitMain)
	}()


	go services.StopWorkpool(ch)
	fmt.Println("attesa comando di uscita")
	uscita := <-exitMain
	_ = uscita
	fmt.Println("****************     exit main              *********")

}