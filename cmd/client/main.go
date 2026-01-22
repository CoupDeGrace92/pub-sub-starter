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
	fmt.Println("Starting Peril client...")
	rabbitCon := routing.RabbitConString
	connection, err := amqp.Dial(rabbitCon)
	if err != nil {
		log.Printf("Error creating rabbitmq connection: %v\n", err)
		return
	}
	defer connection.Close()
	fmt.Println("Connection to rabbitmq successful")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Error welcoming client: %v\n", err)
	}

	gs := gamelogic.NewGameState(username)

	q := fmt.Sprintf("pause.%s", username)
	pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilDirect,
		q,
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gs),
	)

	moveChan, err := connection.Channel()
	if err != nil {
		log.Fatalf("Error creating a channel for %s's moves: %v\n", username, err)
	}

	pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		fmt.Sprintf("army_moves.%s", username),
		"army_moves.*",
		pubsub.Transient,
		handlerMove(gs, moveChan),
	)

	pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		"war",
		routing.WarRecognitionsPrefix+".*",
		pubsub.Durable,
		handlerWar(gs),
	)

repl:
	for {
		cmd := gamelogic.GetInput()
		switch cmd[0] {
		case "move":
			mv, err := gs.CommandMove(cmd)
			if err != nil {
				fmt.Println(err)
			}
			log.Println("Sending move: ", mv)
			pubsub.PublishJSON(moveChan, routing.ExchangePerilTopic, fmt.Sprintf("army_moves.%s", username), mv)
			log.Println("Move published succesfully")
		case "spawn":
			err = gs.CommandSpawn(cmd)
			if err != nil {
				fmt.Println(err)
			}
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming is not allowed")
		case "quit":
			fmt.Println("Exiting Peril...")
			break repl
		default:
			fmt.Println("Error: Unrecognized command")
		}
	}
	fmt.Printf("\nConnection closing\n")

}
