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

type UserChange struct {
	username string
	isAdded  bool
}

func main() {
	client := CreateClientConfig()
	users, err := client.GetUserIdList()
	requests := make(chan RequestWithTarget)
	userChanges := make(chan UserChange)

	if err != nil {
		log.Fatalf("%v", err)
	}
	user_set := set.From(users)
	log.Printf("found %d users", user_set.Size())

	sock, err := NewClient(url, requests, userChanges)
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
	irc, err := InitNewClient(*user_set)
	if err != nil {
		log.Fatalf("failed to initialize client, %v", err)
	}

	go irc.MessageLoop(requests)
	go irc.ReceiveUserChanges(userChanges)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	<-sigChan
}
