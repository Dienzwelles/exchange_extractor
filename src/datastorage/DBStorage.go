package datastorage

import (
	"../models"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	_ "github.com/jinzhu/gorm"
	_ "fmt"
	"strings"
	"github.com/jinzhu/gorm"
	"../utils/sqlcustom"
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

		defer dbe.Close()

		if res2{
			//log.Print("insert new trade")
		}

		if dbe.Error != nil{
			panic(dbe.Error)
		}
	}
}


func StoreBooks(books []models.AggregateBooks){

	conn := NewConnection()
	db := GetConnectionORM(conn)
	defer db.Close()
	//get the last lot
	lastLot := books[0].Lot -1

	//set last lot as old
	if lastLot>=1 {
		setLotAsOld(db, books[0].Exchange_id, books[0].Symbol)
	}
	sqlcustom.BatchInsert(db.DB(), books)
/*
	for i := 0; i < len(books); i++ {

		//book := books[i]

		res2 := db.NewRecord(books)
		dbe := db.Create(&books)
		defer dbe.Close()

		if res2{
			log.Print("insert book")
		}

		if db.Error != nil{
			panic(db.Error)
		}
	}
*/
}




func StoreMarkets(markets []models.Market) {

	conn := NewConnection()
	db := GetConnectionORM(conn)

	defer db.Close()
	for _, market := range markets {
		//fmt.Println(market)
		res2 := db.NewRecord(market)
		rows, _ := db.Table("markets").Select("exchange_id, symbol").Where("exchange_id = ? and symbol = ? ", market.Exchange_id, market.Symbol).Rows()

		defer rows.Close()
		count := 0
		for rows.Next() {
			count++
		}

		if  count  == 0  {
			db.Create(&market)

			if res2{
				log.Print("insert new exchange market")
			}
			if db.Error != nil{
				panic(db.Error)
			}
		}

	}

}




func GetMarkets(exchange_id string) []string{

	conn := NewConnection()
	db := GetConnectionORM(conn)

	defer db.Close()


	rows, _ := db.Table("markets").Select("symbol").Where("exchange_id = ? and evaluated = 1 ", exchange_id).Rows()
	defer rows.Close()

	var markets []string
	for rows.Next() {
		var symbol string
		rows.Scan(&symbol)
		markets = append(markets, symbol)
	}

	return markets
}



func GetLastLot(exchange string, symbol string ) int64 {

	conn := NewConnection()
	db := GetConnectionORM(conn)
	//db.LogMode(true)

	defer db.Close()

	var record models.AggregateBooks

	db.Model(&record).Where("exchange_id = ? and symbol = ? ", exchange, strings.ToUpper(symbol)).Order("lot desc").Last(&record)

	log.Println(record.Lot, exchange, strings.ToUpper(symbol))
	return record.Lot
}


func setLotAsOld(db *gorm.DB, exchange string, symbol string) {

	//db.Table("aggregate_books").Debug().Where("exchange_id = ? and symbol = ? and lot = ?", exchange, strings.ToUpper(symbol), lot).UpdateColumn("obsolete", 1)
	db.Where("exchange_id = ?", exchange).Delete(models.AggregateBook{})

}