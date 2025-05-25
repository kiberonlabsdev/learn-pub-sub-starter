package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {

	json, err := json.Marshal(val)

	if err != nil {
		return err
	}
	// Publish the message to the specified exchange with the given routing key
	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        json,
	})
	if err != nil {
		return err
	}
	return nil
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	// Encode the value as Gob
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(val)
	if err != nil {
		return err
	}
	// Publish the message to the specified exchange with the given routing key
	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/gob",
		Body:        b.Bytes(),
	})
	if err != nil {
		return err
	}
	return nil
}
