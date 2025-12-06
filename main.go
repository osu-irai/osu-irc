package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/hashicorp/go-set/v3"
)

const (
	rabbitmqUrl = "amqp://irai:iraipass@localhost:5672/"
)

type RequestWithTarget struct {
	Id      int     `json:"id"`
	Target  string  `json:"target"`
	Request Request `json:"request"`
}
type Request struct {
	Beatmap Beatmap `json:"beatmap"`
	From    From    `json:"from"`
	Source  int     `json:"source"`
}

type Beatmap struct {
	BeatmapId    int     `json:"beatmapId"`
	BeatmapsetId int     `json:"beatmapsetId"`
	Artist       string  `json:"artist"`
	Title        string  `json:"title"`
	Difficulty   string  `json:"difficulty"`
	Stars        float64 `json:"stars"`
}
type From struct {
	Id        int    `json:"id"`
	Username  string `json:"username"`
	AvatarUrl string `json:"avatarUrl"`
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

	go initMessageBus(requests)
	go initSettingsChangeBus(userChanges)

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
