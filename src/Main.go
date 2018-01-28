package main

import (
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/go-sql-driver/mysql"
	"time"
	"./adapters"
	"./exchanges"
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
	adapters.Instantiate()
	/*
	db, err := gorm.Open("mysql", "mysqlusr:Quid2017!@tcp(localhost)/extractor?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	db.LogMode(true)
	// Migrate the schema

	//db.AutoMigrate(&Trade{})

	// Create

	const limit = 20
	symbol := "BTCUSD"
	url := "https://api.bitfinex.com/v2/trades/t" + symbol + "/hist?limit=" + strconv.Itoa(limit)

	httpClient := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "bitfinex-extractor")
	res, getErr := httpClient.Do(req)

	if getErr != nil {
		log.Fatal(getErr)
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	var trades [][4]float64
	jsonErr := json.Unmarshal(body, &trades)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	for i := 0; i < len(trades); i++ {

		trade := trades[i]

		dbTrade := Trade{Exchange_id: "Bitfinex", Symbol: symbol, Trade_ts: time.Unix(int64(trade[1]/1000), 0), Amount: trade[2], Price: trade[3]}

		res2 := db.NewRecord(dbTrade)
		dbe := db.Create(&dbTrade)

		if res2{

		}

		if dbe.Error != nil{
		panic(dbe.Error)
		}
	}
	/*
	if !res {
		panic("Failed to insert")
	}
	*/
	// Read
/*	var product Product
	db.First(&product, 1) // find product with id 1
	db.First(&product, "code = ?", "L1212") // find product with code l1212

	// Update - update product's price to 2000
	db.Model(&product).Update("Price", 2000)
*/
	/*db.Update(&product)*/
	// Delete - delete product
//	db.Delete(&product)
}
/*
type MyWork struct {
	Name string
	BirthYear int
	WP *workpool.WorkPool
}

func (mw *MyWork) DoWork(workRoutine int) {
	fmt.Printf("%s : %d\n", mw.Name, mw.BirthYear)
	fmt.Printf("Q:%d R:%d\n", mw.WP.QueuedWork(), mw.WP.ActiveRoutines())

	// Simulate some delay
	time.Sleep(100 * time.Millisecond)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	workPool := workpool.New(runtime.NumCPU(), 800)

	shutdown := false // Race Condition, Sorry

	go func() {
		for i := 0; i < 1000; i++ {
			work := MyWork {
				Name: "A" + strconv.Itoa(i),
				BirthYear: i,
				WP: workPool,
			}

			if err := workPool.PostWork("routine", &work); err != nil {
				fmt.Printf("ERROR: %s\n", err)
				time.Sleep(100 * time.Millisecond)
			}

			if shutdown == true {
				return
			}
		}
	}()

	fmt.Println("Hit any key to exit")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	shutdown = true

	fmt.Println("Shutting Down")
	workPool.Shutdown("routine")
}
*/