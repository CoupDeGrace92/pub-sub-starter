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

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.Acknowledge {
	return func(war gamelogic.RecognitionOfWar) pubsub.Acknowledge {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(war)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			log.Printf("Not involved in war between %s and %s\n", war.Attacker.Username, war.Defender.Username)
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			log.Println("War outcome: No units")
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			msg := fmt.Sprintf("%s won a war against %s", winner, loser)
			log.Print(msg)
			pubsub.PublishGameLog(ch, war.Attacker.Username, gs.GetUsername(), msg)
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			msg := fmt.Sprintf("%s won a war against %s", winner, loser)
			log.Print(msg)
			pubsub.PublishGameLog(ch, war.Attacker.Username, gs.GetUsername(), msg)
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			msg := fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
			log.Print(msg)
			pubsub.PublishGameLog(ch, war.Attacker.Username, gs.GetUsername(), msg)
			return pubsub.Ack
		default:
			log.Printf("Error, war outcome unrecognized %v", outcome)
			return pubsub.NackDiscard
		}
	}
}
