package pubsub

import (
	"bytes"
	"encoding/gob"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) Acknowledge,
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
			payload := bytes.NewReader(delivery.Body)
			err = gob.NewDecoder(payload).Decode(&rawResp)
			if err != nil {
				log.Printf("Error decoding %v: %v\n", rawResp, err)
			}
			a := handler(rawResp)

			switch a {
			case Ack:
				err = delivery.Ack(false)
				if err != nil {
					log.Println("Error acknowledging delivery from browser ", err)
				}
			case NackDiscard:
				err = delivery.Nack(false, false)
				if err != nil {
					log.Println("Error acknowledging delivery from browser ", err)
				}
			case NackRequeue:
				err = delivery.Nack(false, true)
				if err != nil {
					log.Println("Error acknowledging delivery from browser ", err)
				}
			}
		}
	}()

	return nil
}
