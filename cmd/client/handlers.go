package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acknowledge {
	return func(ps routing.PlayingState) pubsub.Acknowledge {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		log.Println("Ack type: Ack")
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.Acknowledge {
	return func(move gamelogic.ArmyMove) pubsub.Acknowledge {
		defer fmt.Print("> ")
		m := gs.HandleMove(move)
		switch m {
		case gamelogic.MoveOutComeSafe:
			log.Println("Ack type: Ack")
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			payload := gamelogic.RecognitionOfWar{
				Attacker: move.Player,
				Defender: gs.GetPlayerSnap(),
			}
			err := pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.WarRecognitionsPrefix, payload.Defender.Username),
				payload,
			)
			if err != nil {
				fmt.Printf("error: %s\n", err)
			}
			log.Println("Ack type: Nack and requeue")
			return pubsub.Ack
		default:
			log.Println("Ack type: Nack and discard")
			return pubsub.NackDiscard
		}

	}
}

func handlerWar(gs *gamelogic.GameState) func(gamelogic.RecognitionOfWar) pubsub.Acknowledge {
	return func(war gamelogic.RecognitionOfWar) pubsub.Acknowledge {
		defer fmt.Print("> ")
		outcome, _, _ := gs.HandleWar(war)
		fmt.Printf("WAR DEBUG: me=%+v\nattacker=%+v\ndefender=%+v\noutcome=%v\n\n",
			gs.GetPlayerSnap(), war.Attacker, war.Defender, outcome)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			fmt.Printf("Not involved in war between %s and %s\n", war.Attacker.Username, war.Defender.Username)
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			log.Println("War outcome: No units")
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			log.Printf("You LOST the war between %s and %s", war.Attacker.Username, war.Defender.Username)
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			log.Printf("You WON the war between %s and %s\n", war.Attacker.Username, war.Defender.Username)
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			log.Printf("The war between %s and %s ended in a DRAW\n", war.Attacker.Username, war.Defender.Username)
			return pubsub.Ack
		default:
			log.Printf("Error, war outcome unrecognized %v", outcome)
			return pubsub.NackDiscard
		}
	}
}
