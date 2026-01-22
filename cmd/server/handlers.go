package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerGameLog(gameLog routing.GameLog) pubsub.Acknowledge {
	defer fmt.Print("> ")
	err := gamelogic.WriteLog(gameLog)
	if err != nil {
		log.Println("Error writing log: ", err)
		return pubsub.NackDiscard
	}
	return pubsub.Ack
}
