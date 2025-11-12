package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/hashicorp/go-set/v3"
)

const (
	url = "http://localhost:5076/api/ws/notifications"
)

func main() {
	client := CreateClientConfig()
	users, err := client.GetUserIdList()
	requests := make(chan RequestWithTarget)

	if err != nil {
		log.Fatalf("%v", err)
	}
	user_set := set.From(users)
	log.Printf("found %d users", user_set.Size())

	sock, err := NewClient(url, requests)
	if err != nil {
		log.Fatalf("%v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := sock.Connect(ctx); err != nil {
		log.Fatalf("%v", err)
	}
	defer func(client *Client) {
		err := client.Disconnect()
		if err != nil {
			log.Fatalf("%v", err)
		}
	}(sock)

	if err := sock.JoinUserGroup(ctx); err != nil {
		log.Fatalf("%v", err)
	}
	// go InitNewClient()
	irc, err := InitNewClient(*user_set)
	if err != nil {
		log.Fatalf("failed to initialize client, %v", err)
	}

	irc.MessageLoop(requests)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	<-sigChan
}
