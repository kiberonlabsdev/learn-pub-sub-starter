package pubsub

import (
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int
type AckType int

// SimpleQueueType is an enum to represent
const (
	DURABLE SimpleQueueType = iota
	TRANSIENT
)

const (
	ACK AckType = iota
	NACK_REQUEUE
	NACK_DROP
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {

	channel, err := conn.Channel()

	if err != nil {
		return nil, amqp.Queue{}, err
	}

	var durable, autoDelete, exclusive bool
	if simpleQueueType == DURABLE {
		autoDelete = false
		durable = true
		exclusive = false
	} else {
		autoDelete = true
		exclusive = true
		durable = false
	}

	queue, err := channel.QueueDeclare(queueName, durable, autoDelete, exclusive, false, amqp.Table{

		"x-dead-letter-exchange": routing.ExchangeDLQ,
	})
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	err = channel.QueueBind(queue.Name, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	return channel, queue, nil
}
