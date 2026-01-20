package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

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

	welcome, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Error welcoming client: %v\n", err)
	}

	q := fmt.Sprintf("pause.%s", welcome)
	_, _, err = pubsub.DeclareAndBind(connection, "peril_direct", q, routing.PauseKey, pubsub.Transient)
	if err != nil {
		log.Fatalf("Error creating and binding queue: %v\n", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Printf("\nConnection closing\n")

}
