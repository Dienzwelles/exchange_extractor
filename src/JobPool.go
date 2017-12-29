package src

import (
	"fmt"
	"time"

	"github.com/goinggo/jobpool"
)

type WorkProvider1 struct {
	Name string
}

func (wp *WorkProvider1) RunJob(jobRoutine int) {
	fmt.Printf("Perform Job : Provider 1 : Started: %s\n", wp.Name)
	time.Sleep(2 * time.Second)
	fmt.Printf("Perform Job : Provider 1 : DONE: %s\n", wp.Name)
}

type WorkProvider2 struct {
	Name string
}

func (wp *WorkProvider2) RunJob(jobRoutine int) {
	fmt.Printf("Perform Job : Provider 2 : Started: %s\n", wp.Name)
	time.Sleep(5 * time.Second)
	fmt.Printf("Perform Job : Provider 2 : DONE: %s\n", wp.Name)
}

func main() {
	jobPool := jobpool.New(2, 1000)

	jobPool.QueueJob("main", &WorkProvider1{"Normal Priority : 1"}, false)

	jobPool.
	fmt.Printf("*******> QW: %d AR: %d\n",
		jobPool.QueuedJobs(),
		jobPool.ActiveRoutines())

	time.Sleep(1 * time.Second)

	jobPool.QueueJob("main", &WorkProvider1{"Normal Priority : 2"}, false)
	jobPool.QueueJob("main", &WorkProvider1{"Normal Priority : 3"}, false)

	jobPool.QueueJob("main", &WorkProvider2{"High Priority : 4"}, true)
	fmt.Printf("*******> QW: %d AR: %d\n",
		jobPool.QueuedJobs(),
		jobPool.ActiveRoutines())

	time.Sleep(15 * time.Second)

	jobPool.Shutdown("main")
}