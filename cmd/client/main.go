package main

import (
	"fmt"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func main() {
	fmt.Println("Starting Peril client...")

	connstring := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connstring)
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ:", err)
		return
	}

	defer conn.Close()

	username, err := gamelogic.ClientWelcome()

	if err != nil {
		fmt.Println("Failed to get username", err)
		return
	}

	channel, _, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, "pause."+username, routing.PauseKey, pubsub.TRANSIENT)

	if err != nil {
		fmt.Println("Failed to declare and bind", err)
		return
	}

	gamestate := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, "war", routing.WarRecognitionsPrefix+".#", pubsub.DURABLE, handlerWar(gamestate, channel))

	if err != nil {
		fmt.Println("Failed to subscribe to war", err)
		return
	}

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, "pause."+username, routing.PauseKey, pubsub.TRANSIENT, handlerPause(gamestate))

	if err != nil {
		fmt.Println("Failed to subscribe to pause", err)
		return
	}
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, "army_moves."+username, routing.ArmyMovesPrefix+".*", pubsub.TRANSIENT, handlerMove(gamestate, channel))

	if err != nil {
		fmt.Println("Failed to subscribe to army moves", err)
		return
	}

	if err != nil {
		fmt.Println("Failed to subscribe to pause", err)
		return
	}

	for {

		words := gamelogic.GetInput()

		if len(words) == 0 {
			continue
		}

		switch expression := words[0]; expression {

		case "spawn":
			err = gamestate.CommandSpawn(words)
			if err != nil {
				fmt.Println("Failed to spawn unit:", err)
			}

		case "move":
			move, err := gamestate.CommandMove(words)

			err = pubsub.PublishJSON(channel, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+".*", move)
			if err != nil {
				fmt.Println("Failed to move unit:", err)

			} else {
				fmt.Println("Moved unit:", move)
			}
		case "status":
			gamestate.CommandStatus()
		case "spam":
			if len(words) <= 1 {
				fmt.Println("Please provide a number to spam")
				continue
			}
			amount, err := strconv.Atoi(words[1])
			if err != nil {
				fmt.Println("Invalid number:", words[1])
				continue
			}
			malicious := gamelogic.GetMaliciousLog()

			for i := 0; i < amount; i++ {

				err = pubsub.PublishGob(channel, routing.ExchangePerilTopic, routing.GameLogSlug+"."+gamestate.GetUsername(), malicious)
				if err != nil {
					fmt.Println("Failed to pub malicious", err)
					continue
				}
			}

		case "help":
			gamelogic.PrintClientHelp()
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Unknown command:", expression)
		}

	}

}
