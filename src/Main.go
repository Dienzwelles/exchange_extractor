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

type subscribeMsg struct {
	Event     string  `json:"event"`
	Channel   string  `json:"channel"`
	Pair      string  `json:"pair"`
	ChanID    float64 `json:"chanId,omitempty"`
	Frequency string `json:"freq,omitempty"`
	Precision string `json:"prec,omitempty"`
	Length int `json:"len,omitempty"`
	Key float64 `json:"key,omitempty"`
}

func main() {
	exchanges.Instantiate()
	adapters.Instantiate()

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