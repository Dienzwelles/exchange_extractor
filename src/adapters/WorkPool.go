package adapters

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"time"
	"github.com/goinggo/workpool"
	"../queuemanager"
)

type AdapterWork struct {
	Adapter AdapterInterface
	WP *workpool.WorkPool
}

func (mw *AdapterWork) DoWork(workRoutine int) {
	for {
		trades := mw.Adapter.getTrade()
		queuemanager.Enqueue(trades)
		time.Sleep(4 * time.Second)
	}
}

type DataExtractorWork struct {
	WP *workpool.WorkPool
}

func (mw *DataExtractorWork) DoWork(workRoutine int) {
	queuemanager.Dequeue()
}

func Instantiate() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	workPool := workpool.New(runtime.NumCPU(), 800)

	shutdown := false // Race Condition, Sorry

	go func() {

		var a AdapterInterface
		a = NewBitfinexAdapter().instantiateDefault("BTCUSD")

		//bittrex istance
		var br AdapterInterface
		br = NewBittrexAdapter().instantiateDefault("BTC-DOGE")


		adapterWork := AdapterWork {
			Adapter: a,
			WP: workPool,
		}

		dataExtractorWork := DataExtractorWork{}

		if err := workPool.PostWork("adapterWork", &adapterWork); err != nil {
			fmt.Printf("ERROR: %s\n", err)
			time.Sleep(100 * time.Millisecond)
		}

		if err := workPool.PostWork("dataExtractorWork", &dataExtractorWork); err != nil {
			fmt.Printf("ERROR: %s\n", err)
			time.Sleep(100 * time.Millisecond)
		}


		//adapter bittrex
		brAdapterWork := AdapterWork {
			Adapter: br,
			WP: workPool,
		}

		brDataExtractorWork := DataExtractorWork{}

		if err := workPool.PostWork("adapterWork", &brAdapterWork); err != nil {
			fmt.Printf("ERROR: %s\n", err)
			time.Sleep(100 * time.Millisecond)
		}

		if err := workPool.PostWork("dataExtractorWork", &brDataExtractorWork); err != nil {
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
	workPool.Shutdown("routine")
}