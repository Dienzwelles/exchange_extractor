package properties

import (
	"fmt"
	"encoding/json"
	"os"
	"sync"
)

type Context struct {
	Database struct {
		Host     string `json:"host"`
		Port     string `json:"port"`

		User     string `json:"user"`
		Password string `json:"password"`
		DbName string `json:"dbName"`
	} `json:"database"`
	Bitfinex struct {
		Key     string `json:"key"`
		Secret  string `json:"secret"`
		ExecArbitrage  string `json:"exec_arbitrage"`
	} `json:"bitfinex"`
}

var instance *Context
var instanceVal Context
var once sync.Once

func setup()  {
	instanceVal = LoadConfiguration("./config.json")
	instance = &instanceVal

	//log configurazioni caricate
	fmt.Println(instance.Database.Host + instance.Database.Port)
	fmt.Println(instance.Database.User)
	fmt.Println(instance.Database.Password)
	fmt.Println(instance.Database.DbName)
}


func GetInstance() *Context {
	once.Do(setup)
	return instance
}


func LoadConfiguration(file string) Context {
	var config Context
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}
