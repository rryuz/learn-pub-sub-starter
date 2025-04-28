package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

const connString = "amqp://guest:guest@localhost:5672/"

func main() {
	fmt.Println("Starting Peril client...")

	// Connect to RabbitMQ
	conn, err := amqp.Dial(connString)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	pauseQueueName := fmt.Sprintf("%s.%s", routing.PauseKey, userName)
	armyMovesQueueName := fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, userName)
	armyMovesKey := routing.ArmyMovesPrefix + ".*"

	_, _, err = pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		pauseQueueName,
		routing.PauseKey,
		pubsub.Transient,
	)
	if err != nil {
		log.Fatal(err)
	}

	armyMovesChannel, _, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		armyMovesQueueName,
		armyMovesKey,
		pubsub.Transient,
	)
	if err != nil {
		log.Fatal(err)
	}

	gameState := gamelogic.NewGameState(userName)

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, pauseQueueName, routing.PauseKey, pubsub.Transient, handlerPause(gameState))
	if err != nil {
		log.Fatal(err)
	}

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, armyMovesQueueName, armyMovesKey, pubsub.Transient, handlerMove(gameState))
	if err != nil {
		log.Fatal(err)
	}

mainLoop:
	for {
		userInput := gamelogic.GetInput()
		if len(userInput) == 0 {
			continue
		}
		switch userInput[0] {
		case "spawn":
			err := gameState.CommandSpawn(userInput)
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "move":
			armyMove, err := gameState.CommandMove(userInput)
			if err != nil {
				fmt.Println(err)
				continue
			}
			pubsub.PublishJSON(armyMovesChannel, routing.ExchangePerilTopic, armyMovesKey, armyMove)
			fmt.Println("Units moved")
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			break mainLoop
		default:
			fmt.Println("Command not allowed")
		}
	}
}
