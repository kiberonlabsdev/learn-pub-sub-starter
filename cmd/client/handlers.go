package main

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerMove(gs *gamelogic.GameState, channel *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(move)

		switch outcome {
		case gamelogic.MoveOutcomeSamePlayer:
			fmt.Println("You are moving your own units!")
			return pubsub.NACK_DROP
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(channel, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix+"."+gs.GetUsername(), gamelogic.RecognitionOfWar{
				Attacker: move.Player,
				Defender: gs.GetPlayerSnap(),
			})
			if err != nil {
				return pubsub.NACK_REQUEUE
			}
			return pubsub.ACK

		case gamelogic.MoveOutComeSafe:
			return pubsub.ACK
		default:
			return pubsub.NACK_DROP

		}

	}
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.ACK
	}
}

func handlerWar(gs *gamelogic.GameState, channel *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")

		res, _, _ := gs.HandleWar(rw)
		switch res {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NACK_REQUEUE
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NACK_DROP
		case gamelogic.WarOutcomeDraw:

			data := fmt.Sprintf("A war between  %v and %v resulted in a draw", rw.Attacker.Username, rw.Defender.Username)

			err := pubsub.PublishGob(channel, routing.ExchangePerilTopic, routing.GameLogSlug+"."+rw.Attacker.Username, data)
			if err != nil {
				return pubsub.NACK_REQUEUE
			}

			return pubsub.ACK
		case gamelogic.WarOutcomeOpponentWon:

			data := fmt.Sprintf("%v won a war against %v", rw.Defender.Username, rw.Attacker.Username)

			err := pubsub.PublishGob(channel, routing.ExchangePerilTopic, routing.GameLogSlug+"."+rw.Attacker.Username, data)
			if err != nil {
				return pubsub.NACK_REQUEUE
			}

			return pubsub.ACK

		case gamelogic.WarOutcomeYouWon:
			data := fmt.Sprintf("%v won a war against %v", rw.Attacker.Username, rw.Defender.Username)

			err := pubsub.PublishGob(channel, routing.ExchangePerilTopic, routing.GameLogSlug+"."+rw.Attacker.Username, data)
			if err != nil {
				return pubsub.NACK_REQUEUE
			}

			return pubsub.ACK
		default:
			fmt.Println("Unknown war outcome:", res)
			return pubsub.NACK_DROP
		}
	}
}
