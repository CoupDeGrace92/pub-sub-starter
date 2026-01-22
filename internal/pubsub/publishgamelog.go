package pubsub

import (
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishGameLog(ch *amqp.Channel, initUser, currentUser, message string) error {
	log := routing.GameLog{
		CurrentTime: time.Now(),
		Message:     message,
		Username:    currentUser,
	}
	err := PublishGob(ch, routing.ExchangePerilTopic, routing.GameLogSlug+"."+initUser, log)
	return err
}
