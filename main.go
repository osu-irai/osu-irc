package main

import (
	"context"
	"log"
)

func main() {
	const (
		url = "http://localhost:5076/api/ws/notifications"
	)

	client, err := NewClient(url)
	if err != nil {
		log.Fatalf("%v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		log.Fatalf("%v", err)
	}
	defer func(client *Client) {
		err := client.Disconnect()
		if err != nil {
			log.Fatalf("%v", err)
		}
	}(client)

	if err := client.JoinUserGroup(ctx); err != nil {
		log.Fatalf("%v", err)
	}
}
