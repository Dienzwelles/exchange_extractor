package src

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"time"
	"./adapters"
/*	"./models" */
	"github.com/goinggo/workpool"
)

type MyWork struct {
	/*Adapter adapters.AbstractAdapterInterface*/
	WP *workpool.WorkPool
}

func (mw *MyWork) DoWork(workRoutine int) {
	//mw.Adapter.getData()
}

func measure2(a (adapters.AbstractAdapterInterface)) {
	a.
}
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	workPool := workpool.New(runtime.NumCPU(), 800)

	shutdown := false // Race Condition, Sorry

	go func() {
		for i := 0; i < 1000; i++ {
			work := MyWork {

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