package pubsub

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueType int

const (
	Durable QueueType = iota
	Transient
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	valBytes, err := json.Marshal(val)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        valBytes,
	})
	if err != nil {
		return err
	}

	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType QueueType, // an enum to represent durable or transient
) (*amqp.Channel, amqp.Queue, error) {
	newChannel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	durable := true
	autoDelete := false
	exclusive := false
	if simpleQueueType == Transient {
		durable = false
		autoDelete = true
		exclusive = true
	}

	newQueue, err := newChannel.QueueDeclare(
		queueName,
		durable,
		autoDelete,
		exclusive,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	err = newChannel.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	return newChannel, newQueue, nil
}
