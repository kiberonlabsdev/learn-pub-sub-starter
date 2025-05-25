package main

import (
	"fmt"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerLogs() func(string) pubsub.AckType {
	return func(log string) pubsub.AckType {
		defer fmt.Print("> ")

		fmt.Println("Received log:", log)

		err := gamelogic.WriteLog(routing.GameLog{
			CurrentTime: time.Now(),
			Username:    "Server",
			Message:     log,
		})

		if err != nil {
			fmt.Println("Failed to write log:", err)
			return pubsub.NACK_REQUEUE
		}

		return pubsub.ACK
	}
}

func main() {
	connstring := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connstring)
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ:", err)
		return
	}
	fmt.Println("Starting Peril server...")

	gamelogic.PrintServerHelp()

	rabbit, err := conn.Channel()

	if err != nil {
		fmt.Println("Failed to open a channel:", err)
		return
	}

	defer rabbit.Close()

	err = pubsub.SubscribeGob(conn, routing.ExchangePerilTopic, "game_logs", routing.GameLogSlug+".#", pubsub.DURABLE, handlerLogs())

	if err != nil {
		fmt.Println("Failed to declare and bind game logs", err)
		return
	}

	for {
		words := gamelogic.GetInput()

		if len(words) == 0 {
			continue
		}
		switch expression := words[0]; expression {
		case "pause":
			fmt.Println("Pausing game...")
			err = pubsub.PublishJSON(rabbit, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			})

			if err != nil {
				fmt.Println("Failed to publish message:", err)
				return
			}
		case "resume":
			fmt.Println("Resuming game...")
			err = pubsub.PublishJSON(rabbit, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			})

			if err != nil {
				fmt.Println("Failed to publish message:", err)
				return
			}
		case "quit":
			fmt.Println("Closing game...")
			return
		default:
			fmt.Println("Unknown command:", expression)
		}

	}

	// signalChan := make(chan os.Signal, 1)
	// signal.Notify(signalChan, os.Interrupt)
	// <-signalChan
}
