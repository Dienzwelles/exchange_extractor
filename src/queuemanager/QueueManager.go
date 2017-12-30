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


	q := getQueue(ch)
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
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")

	return conn
}

func getChannel(conn *amqp.Connection) (*amqp.Channel){
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	return ch
}

func getQueue(channel *amqp.Channel) (amqp.Queue){
	q, err := channel.QueueDeclare(
		"trades", // name
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

	q := getQueue(ch)

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
