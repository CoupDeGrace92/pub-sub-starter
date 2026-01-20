package pubsub

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Acknowledge int

const (
	InvalidType Acknowledge = iota
	Ack
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T),
) error {
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	consumeCh, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for delivery := range consumeCh {
			var rawResp T
			err = json.Unmarshal(delivery.Body, &rawResp)
			if err != nil {
				log.Printf("Error unmarshalling %v: %v\n", rawResp, err)
			}
			handler(rawResp)
			err = delivery.Ack(false)
			if err != nil {
				log.Printf("Error acknowledging delivery from broker: %v\n", err)
			}
		}
	}()

	return nil
}
