package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	InvType SimpleQueueType = iota
	Durable
	Transient
)

type durTrans struct {
	durable    bool
	autoDelete bool
	exclusive  bool
}

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	var d durTrans
	switch queueType {
	case Durable:
		d.durable = true
		d.autoDelete = false
		d.exclusive = false
	case Transient:
		d.durable = false
		d.autoDelete = true
		d.exclusive = true
	default:
		err := fmt.Errorf("Option for transient or durable is neither")
		return nil, amqp.Queue{}, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	dlxKey := amqp.Table{
		"x-dead-letter-exchange": "peril_dlx",
	}
	q, err := ch.QueueDeclare(
		queueName,
		d.durable,    //durable
		d.autoDelete, //autodelete
		d.exclusive,  //exclusive
		false,
		dlxKey,
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	err = ch.QueueBind(
		q.Name,
		key,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	return ch, q, nil
}
