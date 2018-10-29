package datastorage

import (
	"math"
	"../models"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	_ "github.com/jinzhu/gorm"
	_ "fmt"
	"strings"
	"github.com/jinzhu/gorm"
	"../utils/sqlcustom"
	"../utils"
)

type AggregateBooks struct {
	//Id int `gorm:"AUTO_INCREMENT"`
	Exchange_id string
	Symbol string
	Price float64
	Amount float64
}


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
	if lastLot == 1 {
		clearBooks(db, books[0].Exchange_id)
	}
	//sqlcustom.

	obsoleteBooks, newBooks, newBooksDelete  := spliceBooks(books)

	if len(newBooks) > 0 {
		rs, err := sqlcustom.BatchDelete(db.DB(), newBooksDelete)

		if err != nil{
			panic(err)
		} else {
			_, err2 := rs.RowsAffected()

			if err2 != nil{
				panic(err2)
			}

		}

		for _, v := range newBooks {
			println("log: ", v.Symbol, ", ", v.Price, ", ",  v.Amount)
		}

		rs, err = sqlcustom.BatchInsert(db.DB(), newBooks)
		if err != nil{
			panic(err)
		} else {
			rows, err2 := rs.RowsAffected()

			if err2 != nil{
				panic(err2)
			}
			println("inseriti", rows, "da inserire ", len(books))
		}
	}

	if len(obsoleteBooks) > 0 {
		rs, err := sqlcustom.BatchDelete(db.DB(), obsoleteBooks)

		if err != nil{
			panic(err)
		} else {
			rows, err2 := rs.RowsAffected()

			if err2 != nil{
				panic(err2)
			}
			println("cancellati", rows, "da cancellare ", len(obsoleteBooks))
			if int(rows) != len(obsoleteBooks){
				print("anomalia in fase di cancellazione")
			}
		}
	}


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

const float64EqualityThreshold = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a - b) <= float64EqualityThreshold
}

func spliceBooks(books []models.AggregateBooks) ([]AggregateBooks, []models.AggregateBooks, []AggregateBooks){
	obsoleteBooks := []AggregateBooks{}
	newBooks := []models.AggregateBooks{}
	newBooksDelete := []AggregateBooks{}

	for _, s := range books {
		if (s.Obsolete) {
			v := AggregateBooks{Exchange_id: s.Exchange_id, Symbol: s.Symbol, Price: s.Price, Amount: utils.Sgn(s.Amount)}
			obsoleteBooks = append(obsoleteBooks, v)
		} else{
			v := AggregateBooks{Exchange_id: s.Exchange_id, Symbol: s.Symbol, Price: s.Price, Amount: utils.Sgn(s.Amount)}
			newBooksDelete = append(newBooksDelete, v)
			newBooks = append(newBooks, s)
		}
	}
	return obsoleteBooks, newBooks, newBooksDelete
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


func clearBooks(db *gorm.DB, exchange string) {

	//db.Table("aggregate_books").Debug().Where("exchange_id = ? and symbol = ? and lot = ?", exchange, strings.ToUpper(symbol), lot).UpdateColumn("obsolete", 1)
	db.Where("exchange_id = ?", exchange).Delete(models.AggregateBook{})

}