package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	rabbitCon := routing.RabbitConString
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
	gamelogic.PrintServerHelp()

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
inf:
	for {
		cmd := gamelogic.GetInput()
		switch cmd[0] {
		case "pause":
			log.Println("Sending pause message")
			pubsub.PublishJSON(pauseResumeCh, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
		case "resume":
			log.Println("Sending resume message")
			pubsub.PublishJSON(pauseResumeCh, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
		case "quit":
			log.Println("Exiting...")
			break inf
		default:
			log.Println("I don't understand the command: ", cmd[0])
		}
	}

	fmt.Println("Connection closing")
}
