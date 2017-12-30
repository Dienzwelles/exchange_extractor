package datastorage

import (
	"../models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

func StoreTrades(trades []models.Trade){
	db, err := gorm.Open("mysql", "mysqlusr:Quid2017!@tcp(localhost)/extractor?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("failed to connect database")
	}
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
