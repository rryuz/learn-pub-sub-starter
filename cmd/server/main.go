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
	fmt.Println("Starting Peril server...")

	// Connect to RabbitMQ
	conn, err := amqp.Dial(connString)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Println("Successfully connected to RabbitMQ")

	connChannel, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}

	gamelogic.PrintServerHelp()

mainLoop:
	for {
		userInput := gamelogic.GetInput()
		if len(userInput) == 0 {
			continue
		}
		switch userInput[0] {
		case "pause":
			fmt.Println("Sending pause message")
			err = pubsub.PublishJSON(connChannel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			})
			if err != nil {
				log.Fatal(err)
			}
		case "resume":
			fmt.Println("Sending pause message")
			err = pubsub.PublishJSON(connChannel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			})
			if err != nil {
				log.Fatal(err)
			}
		case "quit":
			fmt.Println("Exiting")
			break mainLoop
		default:
			fmt.Println("Command not supported")
		}
	}
}
