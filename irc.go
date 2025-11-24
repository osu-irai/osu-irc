package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/hashicorp/go-set/v3"
	"gopkg.in/irc.v4"
)

type OsuClient struct {
	client  irc.Client
	allowed set.Set[string]
}

func CreateClient() irc.ClientConfig {
	nick := os.Getenv("IRC_NICK")
	pass := os.Getenv("IRC_PASS")
	log.Printf("creating client for %v", nick)
	return irc.ClientConfig{
		Nick: nick,
		Pass: pass,
		Name: nick,
		Handler: irc.HandlerFunc(func(c *irc.Client, m *irc.Message) {
			if m.Command == "PRIVMSG" {
				log.Printf("message from %v: %v", m.Name, m.String())
				if strings.HasPrefix(m.Param(1), "!") {
					log.Printf("called command %v", handleCommands(m.Name, m.Param(1)))
				}
			}
		}),
	}
}

func handleCommands(target string, command string) string {
	switch command {
	case "!info":
		return "This is iraibot, IRC interface for Irai"
	}
	return ""
}

func JoinIrcServer() (net.Conn, error) {
	log.Printf("joining irc.ppy.sh")
	return net.Dial("tcp", net.JoinHostPort("irc.ppy.sh", "6667"))
}

func InitNewClient(allowed set.Set[string]) (*OsuClient, error) {
	conf := CreateClient()
	conn, err := JoinIrcServer()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to osu irc")
	}

	client := irc.NewClient(conn, conf)
	go client.Run()
	log.Printf("client is running")
	osu_client := OsuClient{
		client:  *client,
		allowed: allowed,
	}
	return &osu_client, nil
}

func (client *OsuClient) MessageLoop(requestChannel <-chan RequestWithTarget) {
	go func() {
		for {
			req, more := <-requestChannel
			if more {
				client.sendRequest(req)
			} else {
				fmt.Printf("Channel closed")
				return
			}
		}
	}()

}

func (client *OsuClient) ReceiveUserChanges(userChangesChan <-chan UserChange) {
	go func() {
		for {
			req, more := <-userChangesChan
			if more {
				if req.isAdded {
					client.allowed.Insert(req.username)
				} else {
					client.allowed.Remove(req.username)
				}
			} else {
				fmt.Printf("Channel closed")
				return
			}
		}
	}()
}

func (c *OsuClient) SendCommandResponse(target string, response string) error {
	message := irc.Message{
		Command: "PRIVMSG",
		Params: []string{
			target,
			response,
		},
	}

	err := c.client.WriteMessage(&message)
	if err != nil {
		return fmt.Errorf("failed to write to IRC: %v", err)
	}
	log.Printf("sent message to %v", target)

	return nil
}

func (c *OsuClient) sendPrivateMessage(message string, target string) error {
	ircMessage := irc.Message{
		Command: "PRIVMSG",
		Params: []string{
			target,
			message,
		},
	}
	err := c.client.WriteMessage(&ircMessage)
	if err != nil {
		return fmt.Errorf("failed to write to IRC: %v", err)
	}
	log.Printf("sent message to %v", target)
	return nil
}

func (c *OsuClient) sendRequest(request RequestWithTarget) error {
	if !c.allowed.Contains(request.Target) {
		log.Printf("%v does not allow IRC requests", request.Target)
		return nil
	}
	err := c.sendPrivateMessage(request.Target, formatMessage(request))
	if err != nil {
		return fmt.Errorf("failed to write a request: %v", err)
	}
	return nil
}

func formatMessage(request RequestWithTarget) string {
	url := formatUrl(request.Request.Beatmap.BeatmapId, request.Request.Beatmap.BeatmapsetId)
	embed := formatUrlEmbed(url, request.Request.Beatmap.Artist, request.Request.Beatmap.Title, request.Request.Beatmap.Difficulty)
	return fmt.Sprintf("\x01ACTION %v requested %v\x01", request.Request.From.Username, embed)
}

func formatUrl(mapId int, mapsetId int) string {
	return fmt.Sprintf("https://osu.ppy.sh/beatmapsets/%d#/%d", mapsetId, mapId)
}

func formatUrlEmbed(url string, artist string, title string, difficulty string) string {
	return fmt.Sprintf("[%v %v - %v]", url, artist, title)
}
