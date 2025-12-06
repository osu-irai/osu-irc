package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

type UserChange struct {
	Username string `json:"username"`
	isAdded  bool   `json:"isEnabled"`
}

func initSettingsChangeBus(channel chan<- UserChange) {
	queueName := "irc-settings"
	// Connect to RabbitMQ
	conn, err := amqp.Dial(rabbitmqUrl)
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

	// Declare the queue (idempotent - safe to call even if it exists)
	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	// Set QoS to process one message at a time
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
	}

	// Start consuming messages
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack (set to false for manual acknowledgment)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	log.Printf("Connected to RabbitMQ at %s", rabbitmqUrl)
	log.Printf("Waiting for messages on queue '%s'. To exit press CTRL+C", queueName)

	// Channel to handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Process messages in a goroutine
	go func() {
		for d := range msgs {
			var message UserChange
			log.Printf("Received a message")

			// Parse the JSON message
			err := json.Unmarshal(d.Body, &message)
			if err != nil {
				log.Printf("Error parsing message: %v", err)
				log.Printf("Raw message: %s", string(d.Body))
				// Acknowledge even if parsing fails to remove it from queue
				d.Ack(false)
				continue
			}

			// Print the received message
			log.Printf("Received a settings change from %v", message.Username)
			channel <- message

			// Acknowledge the message
			d.Ack(false)
		}
	}()
	<-sigChan
}

func initMessageBus(channel chan<- RequestWithTarget) {
	exchangeName := "request-contract"
	messageType := "request-contract"

	conn, err := amqp.Dial(rabbitmqUrl)
	if err != nil {
		log.Fatalf("Failed to connect to RMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		messageType,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare exchange: %v", err)
	}

	queue, err := ch.QueueDeclare(
		"",
		false,
		true,
		true,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	err = ch.QueueBind(
		queue.Name,
		messageType,
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to bind queue to exchange: %v", err)
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
	}

	// Start consuming messages
	msgs, err := ch.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack (set to false for manual acknowledgment)
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	log.Printf("Connected to RabbitMQ at %s", rabbitmqUrl)
	log.Printf("Listening for messages of type '%s' on exchange '%s'", messageType, exchangeName)
	log.Println("To exit press CTRL+C")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		for d := range msgs {
			var message RequestWithTarget

			// Parse the JSON message
			err := json.Unmarshal(d.Body, &message)
			if err != nil {
				log.Printf("Error parsing message: %v", err)
				log.Printf("Raw message: %s", string(d.Body))
				d.Ack(false)
				continue
			}

			// Print the received message
			fmt.Println("\nReceived request message")
			channel <- message
			fmt.Printf("\n %+v", message)

			// Acknowledge the message
			d.Ack(false)
		}
	}()
	<-sigChan
}
