package main

import (
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/natefinch/lumberjack.v2"
	"./adapters"
	"./exchanges"
	"./services"


	"time"
	"fmt"
	//"./utils/sqlcustom"
	//"./models"
	"log"
)

type BitfinexTrade struct {
	MTS time.Time
	AMOUNT float64
	PRICE float64
	RATE float64
	PERIOD int
}


func main() {

	//adapters.GetTradesFromTS()
/*
	file, err := os.OpenFile("/home/marcob/log/log.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()
*/

	log.SetOutput(&lumberjack.Logger{
		Filename:   "../log/log.log",
		MaxSize:    5,  // megabytes after which new file is created
		MaxBackups: 30,  // number of backups
		MaxAge:     28, //days
	})

	exchanges.Instantiate()
	//adapters.Instantiate()
	ch := make(chan bool)
	exitMain := make(chan bool)
	go func() {
		adapters.Instantiate(ch, exitMain)
	}()


	go services.StopWorkpool(ch)
	fmt.Println("attesa comando di uscita")
	uscita := <-exitMain
	_ = uscita
	fmt.Println("****************     exit main              *********")

}
/*
func Pippo() {
	c := websocket.New()
	wg := sync.WaitGroup{}
	wg.Add(3) // 1. Info with version, 2. Subscription event, 3. 3 x data message

	err := c.Connect()
	if err != nil {
		log.Fatal("Error connecting to web socket : ", err)
	}
	defer c.Close()

	subs := make(chan interface{}, 10)
	unsubs := make(chan interface{}, 10)
	infos := make(chan interface{}, 10)
	trades := make(chan interface{}, 100)

	errch := make(chan error)
	go func() {
		for {
			select {
			case msg := <-c.Listen():
				if msg == nil {
					return
				}
				log.Printf("recv msg: %#v", msg)
				switch m := msg.(type) {
				case error:
					errch <- msg.(error)
				case *websocket.UnsubscribeEvent:
					unsubs <- m
				case *websocket.SubscribeEvent:
					subs <- m
				case *websocket.InfoEvent:
					infos <- m
				case *bitfinex.TradeExecutionUpdateSnapshot:
					trades <- m
				case *bitfinex.Trade:
					trades <- m
				case *bitfinex.TradeExecutionUpdate:
					trades <- m
				case *bitfinex.TradeExecution:
					trades <- m
				case *bitfinex.TradeSnapshot:
					trades <- m
				default:
					log.Print("test recv: %#v", msg)
				}
			}
		}
	}()

	ctx, cxl := context.WithTimeout(context.Background(), time.Second*500)
	defer cxl()
	id, err := c.SubscribeTrades(ctx, bitfinex.TradingPrefix+bitfinex.BTCUSD)
	if err != nil {
		log.Fatal(err)
	}

	if err := wait2(trades, 1, errch, 2*time.Second); err != nil {
		log.Print("failed to receive trade message from websocket: %s", err)
	}

	log.Print(id)

	select {}
	/*err = c.Unsubscribe(ctx, id)
	if err != nil {
		log.Fatal(err)
	}

	if err := wait2(unsubs, 1, errch, 2*time.Second); err != nil {
		log.Print("failed to receive unsubscribe message from websocket: %s", err)
	}*//*
}*/

func wait2(ch <-chan interface{}, count int, bc <-chan error, t time.Duration) error {
	c := make(chan interface{})
	go func() {
		<-ch
		close(c)
	}()
	select {
	case <-bc:
		return fmt.Errorf("transport closed while waiting")
	case <-c:
		return nil // normal
	case <-time.After(t):
		return fmt.Errorf("timed out waiting")
	}
	return nil
}