package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
) error {

	return subscribe(conn, exchange, queueName, key, simpleQueueType, handler, func(data []byte) (T, error) {
		var val T
		err := json.Unmarshal(data, &val)
		if err != nil {
			return val, err
		}
		return val, nil
	})
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
) error {

	return subscribe(conn, exchange, queueName, key, simpleQueueType, handler, func(data []byte) (T, error) {
		var val T
		var buf bytes.Buffer
		buf.Write(data)

		err := gob.NewDecoder(&buf).Decode(&val)
		if err != nil {
			return val, err
		}
		return val, nil
	})
}

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
	unmarshaller func([]byte) (T, error),
) error {
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)

	if err != nil {
		return err
	}

	channel.Qos(10, 0, false)
	delivery, err := channel.Consume(queue.Name, "", false, false, false, false, nil)

	if err != nil {
		return err
	}

	go func() {
		for msg := range delivery {

			val, err := unmarshaller(msg.Body)

			if err != nil {
				// Should probably do something more sophisticated here
				fmt.Println("Failed to unmarshal message:", err)
				continue
			}

			ackType := handler(val)

			switch ackType {
			case ACK:
				// Acknowledge the message
				msg.Ack(false)
			case NACK_REQUEUE:
				// Reject the message and requeue it
				msg.Nack(false, true)
			case NACK_DROP:
				// Reject the message and drop it
				msg.Nack(false, false)
			default:
				// Assume reject and drop
				msg.Nack(false, false)
			}
		}
	}()
	return nil

}
