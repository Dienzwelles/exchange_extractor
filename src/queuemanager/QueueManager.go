package queuemanager

import (
	"log"
	"fmt"
	"encoding/json"
	"../models"
	"github.com/streadway/amqp"
	"../datastorage"
	"../properties"
	"sync"
)

var once sync.Once
var connection * amqp.Connection

type ConnectionRabbitMQ struct {
	Host	 string
	Port	 string
	User	 string
	Pass	 string
	DbName	 string
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func Enqueue(trades []models.Trade){
	conn := getConnection()

	ch := getChannel(conn)
	defer ch.Close()


	q := getQueue(ch, "trades")
	trades2B, _ := json.Marshal(trades)
	body := string(trades2B)
	err := ch.Publish(
		"",     // exchange
		q.Name, // routing key
		true,  // mandatory
		false,  // immediate
		amqp.Publishing {
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message")
}

func setup(){

	//reperisco dalle properties i parametri di connessione
	ac := properties.GetInstance()
	/* only to check the credentials
	fmt.Println(ac.RabbitMQ.User )
	fmt.Println(ac.RabbitMQ.Password )
	fmt.Println("amqp://" + ac.RabbitMQ.User + ":" + ac.RabbitMQ.Password + "@"  + ac.RabbitMQ.Host + ":" + ac.RabbitMQ.Port + "/")
	*/
	conn, err := amqp.Dial("amqp://" + ac.RabbitMQ.User + ":" + ac.RabbitMQ.Password + "@"  + ac.RabbitMQ.Host + ":" + ac.RabbitMQ.Port + "/")
	failOnError(err, "Failed to connect to RabbitMQ")

	connection = conn
}

func getConnection() (* amqp.Connection){
	once.Do(setup)
	return connection
}

func getChannel(conn *amqp.Connection) (*amqp.Channel){
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	return ch
}

func getQueue(channel *amqp.Channel, name string) (amqp.Queue){
	q, err := channel.QueueDeclare(
		name, // name
		true,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	return q;
}

func Dequeue(){
	conn := getConnection()

	ch := getChannel(conn)
	defer ch.Close()

	q := getQueue(ch, "trades")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			//log.Printf("Received a message: %s", d.Body)
			var trades []models.Trade
			jsonErr := json.Unmarshal(d.Body, &trades)
			if jsonErr != nil {
				log.Fatal(jsonErr)
			}

			datastorage.StoreTrades(trades)
		}
	}()
	<-forever
}


func BooksEnqueue(books []models.AggregateBooks){
	if books == nil || len(books) == 0{
		return
	}

	conn := getConnection()

	ch := getChannel(conn)
	defer ch.Close()


	q := getQueue(ch, "books")
	books2B, _ := json.Marshal(books)
	body := string(books2B)
	err := ch.Publish(
		"",     // exchange
		q.Name, // routing key
		true,  // mandatory
		false,  // immediate
		amqp.Publishing {
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message type books")
}


func BooksDequeue(startArbitrage chan string, waitArbitrage chan string){
	conn := getConnection()

	ch := getChannel(conn)
	defer ch.Close()

	q := getQueue(ch, "books")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer type books")

	forever := make(chan bool)

	go func() {
		//wait := "START"
		//var waitChan chan string
		for d := range msgs {
			//log.Printf("Received a message: %s", d.Body)

			var books []models.AggregateBooks
			jsonErr := json.Unmarshal(d.Body, &books)
			if jsonErr != nil {
				log.Fatal(jsonErr)
			}

			print("inizio scrittura")

			if len(books) > 0 {
				//if len(wait) > 0 {
					//waitChan = make(chan string)
					//wait = ""
					storeAndArbitrage(books, nil/*waitChan*/, startArbitrage, waitArbitrage)
				/*} /*else {
					println("scartato")
				}*/
			}

			print("fine scrittura")
			/*
			if len(wait) == 0 {
				wait = waitOperation(waitChan)
			}
			*/
		}
	}()
	<-forever
}

func storeAndArbitrage(books []models.AggregateBooks, waitChan chan string, startArbitrage chan string, waitArbitrage chan string){
	//t := time.Now()
	//fmt.Println(t.Format("20060102150405"))

	datastorage.StoreBooks(books)
	//exchangeId := books[0].Exchange_id
	//startArbitrage <- exchangeId

	//<- waitArbitrage
	//waitChan <- exchangeId
}

func waitOperation(waitChan chan string) string{
	select {
	case msg := <-waitChan:
		return msg
	default:
		return ""
	}
}
