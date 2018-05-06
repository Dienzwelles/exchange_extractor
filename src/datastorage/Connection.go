package datastorage

import (
	/*
	_ "github.com/go-sql-driver/mysql"*/
	"database/sql"
	"github.com/jinzhu/gorm"
	"../properties"
	"fmt"
)


type ConnectionEl struct {
	Host	 string
	Port	 string
	User	 string
	Pass	 string
	DbName	 string
}

func NewConnection() ConnectionEl{
	//ottengo application context
	ac := properties.GetInstance()
	conn := ConnectionEl{
		Host:   ac.Database.Host,
		Port:   ac.Database.Port,
		User:   ac.Database.User,
		Pass:   ac.Database.Password,
		DbName: ac.Database.DbName,
	}
	fmt.Println(conn.Host)
	fmt.Println(conn.User)
	return conn
}



func GetConnection(conn ConnectionEl) *sql.DB  {
	db, err := sql.Open("mysql", conn.User + ":" + conn.Pass + "@tcp(" + conn.Host + ":" + conn.Port + ")/" + conn.DbName + "?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic(err.Error())  // Just for example purpose. You should use proper error handling instead of panic
	}


	//pinga la risorsa
	err = db.Ping()
	if err != nil {
		panic("errore ping") // proper error handling instead of panic in your app
	}
	return db
}


func GetConnectionORM(conn ConnectionEl) *gorm.DB  {
	db, err := gorm.Open("mysql", conn.User + ":" + conn.Pass + "@tcp(" + conn.Host + ":" + conn.Port + ")/" + conn.DbName + "?charset=utf8&parseTime=True&loc=Local")
	db.LogMode(true)
	if err != nil {
		panic(err.Error())  // Just for example purpose. You should use proper error handling instead of panic
	}
	return db
}
