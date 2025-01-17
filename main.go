package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type Message struct {
	Pattern string `json:"pattern"`
	Data    struct {
		EnrollmentID      string   `json:"enrollmentID"`
		ChannelName       string   `json:"channelName"`
		ChainCodeName     string   `json:"chainCodeName"`
		TransactionName   string   `json:"transactionName"`
		TransactionParams []string `json:"transactionParams"`
	} `json:"data"`
	ID string `json:"id"`
}

func main() {
	// Connect to RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Create a channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// Declare a queue
	queueName := "submit-transaction"
	// _, err = ch.QueueDeclare(
	// 	queueName, // Name of the queue
	// 	true,      // Durable
	// 	false,     // Delete when unused
	// 	false,     // Exclusive
	// 	false,     // No-wait
	// 	nil,       // Arguments
	// )
	// if err != nil {
	// 	log.Fatalf("Failed to declare a queue: %v", err)
	// }
	for i := 0; i < 1000; i++ {
		id, err := generateHexString(22)
		if err != nil {
			log.Fatalf("Failed to generate random ID: %v", err)
		}
		// Create a JSON message
		message := Message{
			Pattern: "'process_transaction'",
			Data: struct {
				EnrollmentID      string   `json:"enrollmentID"`
				ChannelName       string   `json:"channelName"`
				ChainCodeName     string   `json:"chainCodeName"`
				TransactionName   string   `json:"transactionName"`
				TransactionParams []string `json:"transactionParams"`
			}{
				EnrollmentID:      "b0f04e1e701011bc117605a1d979c5cc8e312d3a",
				ChannelName:       "kalp",
				ChainCodeName:     "rpccu",
				TransactionName:   "AddPoints",
				TransactionParams: []string{"b0f04e1e701011bc117605a1d979c5cc8e312d3a", "1"},
			},
			ID: id,
		}

		// Convert the message to JSON
		messageBody, err := json.Marshal(message)
		if err != nil {
			log.Fatalf("Failed to marshal message to JSON: %v", err)
		}

		// Publish the JSON message
		err = ch.Publish(
			"",        // Exchange
			queueName, // Routing key (queue name)
			false,     // Mandatory
			false,     // Immediate
			amqp.Publishing{
				ContentType: "application/json", // Specify JSON content type
				Body:        messageBody,
			},
		)
		if err != nil {
			log.Fatalf("Failed to publish a message: %v", err)
		}
	}

	log.Printf("JSON message sent to queue %s", queueName)
}

func generateHexString(length int) (string, error) {
	if length%2 != 0 {
		return "", fmt.Errorf("length must be even to represent bytes as hexadecimal")
	}

	// Calculate the number of bytes needed
	byteLength := length / 2
	bytes := make([]byte, byteLength)

	// Fill the byte slice with random data
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode the bytes to hexadecimal
	return hex.EncodeToString(bytes), nil
}
