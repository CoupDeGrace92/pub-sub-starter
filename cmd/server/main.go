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

	err = pauseResumeCh.ExchangeDeclare(
		"peril_topic",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Error declaring exchange: %v\n", err)
	}

	err = pauseResumeCh.ExchangeDeclare(
		"peril_dlx",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)

	_, _, err = pubsub.DeclareAndBind(connection, "peril_dlx", "peril_dlq", "", pubsub.Durable)
	if err != nil {
		log.Fatalf("Error creating and binding dead letter queue: %v\n", err)
	}

	_, _, err = pubsub.DeclareAndBind(connection, routing.ExchangePerilTopic, routing.GameLogSlug, "game_logs.*", pubsub.Durable)
	if err != nil {
		log.Fatalf("Error creating and binding game log queue: %v", err)
	}

	forever := make(chan struct{})

	go func() {
		pubsub.SubscribeGob(
			connection,
			routing.ExchangePerilTopic,
			routing.GameLogSlug,
			"game_logs.*",
			pubsub.Durable,
			handlerGameLog,
		)
	}()

	<-forever
	/*
	   repl:

	   	for {
	   		cmd := gamelogic.GetInput()
	   		if len(cmd) == 0 {
	   			continue
	   		}
	   		switch cmd[0] {
	   		case "pause":
	   			log.Println("Sending pause message")
	   			pubsub.PublishJSON(pauseResumeCh, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
	   		case "resume":
	   			log.Println("Sending resume message")
	   			pubsub.PublishJSON(pauseResumeCh, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
	   		case "quit":
	   			log.Println("Exiting...")
	   			break repl
	   		default:
	   			log.Println("I don't understand the command: ", cmd[0])
	   		}
	   	}

	   	fmt.Println("Connection closing")
	*/
}
