package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	rabbitCon := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(rabbitCon)
	if err != nil {
		log.Printf("Error creating rabbit mq connection: %v\n", err)
		return
	}
	defer connection.Close()
	fmt.Println("Connection to rabbitmq succesful")

	pauseResumeCh, err := connection.Channel()
	if err != nil {
		log.Printf("Error creating pause/resume channel: %v\n", err)
		return
	}

	err = pauseResumeCh.ExchangeDeclare(
		"peril_direct", //Exchange name
		"direct",       //type
		true,           //durable
		false,          //auto-deleted
		false,          //internal
		false,          //no-wait
		nil,            //args
	)

	if err != nil {
		log.Printf("Error declaring exchange: %v\n", err)
		return
	}

	testQueue, err := pauseResumeCh.QueueDeclare(
		"pause_test",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Error in creating queue: %v\n", err)
		return
	}
	fmt.Println("Queue created")
	err = pauseResumeCh.QueueBind(
		testQueue.Name,
		routing.PauseKey,
		"peril_direct",
		false,
		nil,
	)
	if err != nil {
		log.Printf("Error in binding test queue")
		return
	}
	fmt.Println("Queue bound to peril_direct")

	pubsub.PublishJSON(pauseResumeCh, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("Connection closing")
}
