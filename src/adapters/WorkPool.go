package adapters

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"time"
	"github.com/goinggo/workpool"
	"../queuemanager"
	//"../datastorage"
	"../models"
	"../arbitrage"
)

var shutdown bool
var startArbitrage chan string
var waitArbitrage chan string

type AdapterWork struct {
	Adapter AdapterInterface
	WP *workpool.WorkPool
}

func (mw *AdapterWork) DoWork(workRoutine int) {
	fmt.Printf("*******> WR: %d \n", workRoutine)

	chantrades := mw.Adapter.getTrade()

	for i := 0; i < len(chantrades); i++ {
		go enqueueTrade(chantrades[i])
	}

	for {
		time.Sleep(1 * time.Second)
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

func enqueueTrade(chantrade chan []models.Trade){
	for {
		data := <-chantrade
		queuemanager.Enqueue(data)
	}
}

func enqueueBook(chanbook chan []models.AggregateBooks){
	for {
		data := <-chanbook
		if data != nil && len(data) > 0{
			queuemanager.BooksEnqueue(data)
		} else {
			panic("Error empty data")
		}

	}
}

func initArbitrageSynch(){
	if startArbitrage == nil {
		startArbitrage = make(chan string)
	}

	if waitArbitrage == nil {
		waitArbitrage = make(chan string)
	}
}

func (mw *AdapterBookWork) DoWork(workRoutine int) {
	fmt.Printf("*******> WR: %d \n", workRoutine)
	//for {
	initArbitrageSynch()
		chanbook, _ := mw.Adapter.getAggregateBooks()

		//for i := 0; i < len(chanbooks); i++ {
			go enqueueBook(chanbook)
		//}

		/*
		for {
			time.Sleep(100 * time.Millisecond)
			if shutdown == true {
				return
			}

			if doReset(reset) {
				//closeChanbooks(chanbooks)
				for{
					time.Sleep(500 * time.Millisecond)
					print("*")
				}
				break
			}
		}
		*/
	//}
}


type DataExtractorBooksWork struct {
	WP *workpool.WorkPool
}

func (mw *DataExtractorBooksWork) DoWork(workRoutine int) {
	queuemanager.BooksDequeue(startArbitrage, waitArbitrage)
}

func ProvaArbitrage(arbitrage models.Arbitrage) {
	var a AdapterInterface
	a = NewBitfinexAdapter().instantiateDefault("BTCUSD")

	a.executeArbitrage(arbitrage)
}

func Instantiate() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	workPool := workpool.New(8, 800)

	shutdown = false // Race Condition, Sorry

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
			var ok AdapterInterface
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



			if shutdown == true {
				return
			}

		}()

	/*book section
	  1 - get the list of the symbol for each exchange
	  2 - generete a go routine for each symbol
	*/


	//subroutine to get books

	go func() {

		//adapter istance
		var a AdapterInterface
		a = NewBitfinexAdapter().instantiateDefault("BTCUSD")

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

	}()


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


	for{
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println("Hit any key to exit")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	shutdown = true

	fmt.Println("Shutting Down")

	workPool.Shutdown("adapterWork")

}

func doReset(reset chan int) bool{
	select {
	case <- reset:
		return true
	default:
		return false
	}
}

type AdapterArbitrageWork struct {
	Adapter AdapterInterface
	WP *workpool.WorkPool
}

func (mw *AdapterArbitrageWork) DoWork(workRoutine int) {
	fmt.Printf("*******> WR: %d \n", workRoutine)
	initArbitrageSynch()
	for {
		//pass the arbitrage selected

		exchangeId := <- startArbitrage

		arbitrages := arbitrage.ExtractArbitrage(exchangeId)


		for _, arbitrage := range arbitrages {
			mw.Adapter.executeArbitrage(arbitrage)
		}


		if len(arbitrages) > 0 {
			print("Pippo")
		}

		if shutdown == true {
			return
		}

		waitArbitrage <- exchangeId
	}

}

/*
func closeChanbooks(chanbooks [] chan []models.AggregateBook){
	for i := 0; i < len(chanbooks); i++ {
		close(chanbooks[i])
	}
}
*/