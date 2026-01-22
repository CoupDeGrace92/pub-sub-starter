package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(val)
	if err != nil {
		return err
	}
	payload := buf.Bytes()

	msg := amqp.Publishing{
		ContentType: "application/gob",
		Body:        payload,
	}

	ch.PublishWithContext(context.Background(), exchange, key, false, false, msg)
	return nil
}
