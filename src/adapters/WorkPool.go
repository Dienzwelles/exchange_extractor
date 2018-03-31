package adapters

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"time"
	"github.com/goinggo/workpool"
	"../queuemanager"
	_"../datastorage"
	"../models"
)

var shutdown bool

type AdapterWork struct {
	Adapter AdapterInterface
	WP *workpool.WorkPool
}

func (mw *AdapterWork) DoWork(workRoutine int) {
	fmt.Printf("*******> WR: %d \n", workRoutine)
	for {
		trades := mw.Adapter.getTrade()
		queuemanager.Enqueue(trades)
		time.Sleep(10 * time.Second)
		if shutdown == true {
			return
		}
	}

}

type DataExtractorWork struct {
	WP *workpool.WorkPool
}

func (mw *DataExtractorWork) DoWork(workRoutine int) {
	queuemanager.Dequeue()
}


type AdapterBookWork struct {
	Adapter AdapterInterface
	WP *workpool.WorkPool
}

func (mw *AdapterBookWork) DoWork(workRoutine int) {
	fmt.Printf("*******> WR: %d \n", workRoutine)
	for {
		books := mw.Adapter.getAggregateBooks()
		queuemanager.BooksEnqueue(books)
		time.Sleep(10 * time.Second)
		if shutdown == true {
			return
		}
	}

}


type DataExtractorBooksWork struct {
	WP *workpool.WorkPool
}

func (mw *DataExtractorBooksWork) DoWork(workRoutine int) {
	queuemanager.BooksDequeue()
}



func Instantiate() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	workPool := workpool.New(8, 800)

	shutdown = false // Race Condition, Sorry
/*
	go func() {

		var a AdapterInterface
		a = NewBitfinexAdapter().instantiateDefault("BTCUSD")

		adapterWork := AdapterWork{
			Adapter: a,
			WP:      workPool,
		}

		dataExtractorWork := DataExtractorWork{}

		if err := workPool.PostWork("adapterWork", &adapterWork); err != nil {
			fmt.Printf("ERROR: %s\n", err)
			time.Sleep(100 * time.Millisecond)
		}

		//		dataExtractorWork := DataExtractorWork{}
		if err := workPool.PostWork("dataExtractorWork", &dataExtractorWork); err != nil {
			fmt.Printf("ERROR: %s\n", err)
			time.Sleep(100 * time.Millisecond)
		}

		//bittrex istance
		var br AdapterInterface
		br = NewBittrexAdapter().instantiateDefault("BTC-DOGE")

		//adapter bittrex
		brAdapterWork := AdapterWork{
			Adapter: br,
			WP:      workPool,
		}

		if err := workPool.PostWork("brAdapterWork", &brAdapterWork); err != nil {
			fmt.Printf("ERROR: %s\n", err)
			time.Sleep(100 * time.Millisecond)
		}

		//okex istance
		/*		var ok AdapterInterface
		ok = NewOkexAdapter().instantiateDefault("ltc_btc")


		//adapter bittrex
		okAdapterWork := AdapterWork {
			Adapter: ok,
			WP: workPool,
		}



		if err := workPool.PostWork("okAdapterWork", &okAdapterWork); err != nil {
			fmt.Printf("ERROR: %s\n", err)
			time.Sleep(100 * time.Millisecond)
		}

*/
/*
		if shutdown == true {
			return
		}

	}()

	/*book section
	  1 - get the list of the symbol for each exchange
	  2 - generete a go routine for each symbol
	*/
/*
	symbols := datastorage.GetMarkets(BITFINEX)
	//subroutine to get books
	for _, symbol := range symbols {
		go func(symbol string) {

			//adapter istance
			var a AdapterInterface
			a = NewBitfinexAdapter().instantiateDefault(symbol)

			adapterBookWork := AdapterBookWork{
				Adapter: a,
				WP:      workPool,
			}

			if err := workPool.PostWork("adapterBookWork", &adapterBookWork); err != nil {
				fmt.Printf("ERROR: %s\n", err)
				time.Sleep(100 * time.Millisecond)
			}

			if shutdown == true {
				return
			}

		}(symbol)
	}

	//subroutine to extract book and store it
	go func() {
		dataExtractorBooksWork := DataExtractorBooksWork{}
		if err := workPool.PostWork("dataExtractorBooksWork", &dataExtractorBooksWork); err != nil {
			fmt.Printf("ERROR: %s\n", err)
			time.Sleep(100 * time.Millisecond)
		}

		if shutdown == true {
			return
		}
	}()
*/
	//subroutine to execute arbitrage
	go func() {

		//adapter istance
		var a AdapterInterface
		a = NewBitfinexAdapter().instantiateDefault("")

		adapterArbitrageWork := AdapterArbitrageWork{
			Adapter: a,
			WP:      workPool,
		}

		if err := workPool.PostWork("adapterArbitrageWork", &adapterArbitrageWork); err != nil {
			fmt.Printf("ERROR: %s\n", err)
			time.Sleep(100 * time.Millisecond)
		}

		if shutdown == true {
			return
		}

	}()


	fmt.Println("Hit any key to exit")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	shutdown = true

	fmt.Println("Shutting Down")

	workPool.Shutdown("adapterWork")

}









type AdapterArbitrageWork struct {
	Adapter AdapterInterface
	WP *workpool.WorkPool
}

func (mw *AdapterArbitrageWork) DoWork(workRoutine int) {
	fmt.Printf("*******> WR: %d \n", workRoutine)
	for {
		//pass the arbitrage selected
		arb:= models.Arbitrage{SymbolStart: "btcusd",SymbolTransitory:"bchbtc", SymbolEnd:"bchusd", AmountStart: 0.003, Price: 30}
		mw.Adapter.executeArbitrage(arb)

		time.Sleep(10 * time.Second)
		if shutdown == true {
			return
		}
	}

}