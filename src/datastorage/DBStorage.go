package datastorage

import (
	"../models"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	_ "github.com/jinzhu/gorm"
	_ "fmt"
	_ "strings"
)

func StoreTrades(trades []models.Trade){
	//db, err := gorm.Open("mysql", "mysqlusr:Quid2017!@tcp(localhost)/extractor?charset=utf8&parseTime=True&loc=Local")
	/*db, err := gorm.Open("mysql", "root:root@tcp(localhost)/extractor?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("failed to connect database")
	}*/
	conn := NewConnection()
	db := GetConnectionORM(conn)

	defer db.Close()

	//db.LogMode(true)
	// Migrate the schema

	//db.AutoMigrate(&Trade{})

	// Create

	for i := 0; i < len(trades); i++ {

		trade := trades[i]

		res2 := db.NewRecord(trade)
		dbe := db.Create(&trade)

		if res2{
			log.Print("insert new trade")
		}

		if dbe.Error != nil{
			panic(dbe.Error)
		}
	}
}


func StoreBooks(books []models.AggregateBook){

	conn := NewConnection()
	db := GetConnectionORM(conn)

	defer db.Close()


	for i := 0; i < len(books); i++ {

		book := books[i]

		res2 := db.NewRecord(book)
		dbe := db.Create(&book)

		if res2{
			log.Print("insert book")
		}

		if dbe.Error != nil{
			panic(dbe.Error)
		}
	}
}




func StoreMarkets(markets []models.Market) {

	conn := NewConnection()
	db := GetConnectionORM(conn)

	defer db.Close()

	for _, market := range markets {
		//fmt.Println(market)
		res2 := db.NewRecord(market)
		rows, _ := db.Table("markets").Select("exchange_id, symbol").Where("exchange_id = ? and symbol = ? ", market.Exchange_id, market.Symbol).Rows()
		count := 0
		for rows.Next() {
			count++
		}

		if  count  == 0  {
			dbe := db.Create(&market)
			if res2{
				log.Print("insert new exchange market")
			}
			if dbe.Error != nil{
				panic(dbe.Error)
			}
		}

	}

}




func GetMarkets(exchange_id string) []string{

	conn := NewConnection()
	db := GetConnectionORM(conn)

	defer db.Close()


	rows, _ := db.Table("markets").Select("symbol").Where("exchange_id = ? and evaluated = 1 ", exchange_id).Rows()
	var markets []string
	for rows.Next() {
		var symbol string
		rows.Scan(&symbol)
		markets = append(markets, symbol)
	}

	return markets
}


func GetLastBulk(exchange string, symbol string ) int64 {

	conn := NewConnection()
	db := GetConnectionORM(conn)

	defer db.Close()

	type LastBulk struct{
		exchange_id string
		symbol string
		bulk int64
	}
	var lastBulk LastBulk

	db.Table("accorpate_books").Select("exchange_id, symbol, max(bulk)").Where("exchange_id = ? and symbol = ? ", exchange, symbol).First(&lastBulk)

	return lastBulk.bulk
}


func SetBulkAsOld(exchange string, symbol string, bulk int64 ) {

	conn := NewConnection()
	db := GetConnectionORM(conn)

	defer db.Close()

	db.Table("accorpate_books").Select("exchange_id, symbol, max(bulk)").Where("exchange_id = ? and symbol = ? ", exchange, symbol).First(&lastBulk)

	return lastBulk
}