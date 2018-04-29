package queuemanager

import (
	"log"
	"fmt"
	"encoding/json"
	"../models"
	"github.com/streadway/amqp"
	"../datastorage"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func Enqueue(trades []models.Trade){
	conn := getConnection()
	defer conn.Close()

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

func getConnection() (* amqp.Connection){
	conn, err := amqp.Dial("amqp://guest:guest@linux-a3kt:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")

	return conn
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
	defer conn.Close()

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
			log.Printf("Received a message: %s", d.Body)
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


func BooksEnqueue(books []models.AggregateBook){
	if books == nil || len(books) == 0{
		return
	}

	conn := getConnection()
	defer conn.Close()

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
	defer conn.Close()

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
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			var books []models.AggregateBook
			jsonErr := json.Unmarshal(d.Body, &books)
			if jsonErr != nil {
				log.Fatal(jsonErr)
			}

			if len(books) > 0 {
				datastorage.StoreBooks(books)
				startArbitrage <- books[0].Exchange_id
			}

			<- waitArbitrage
		}
	}()
	<-forever
}
