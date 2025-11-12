package main

import (
	"fmt"
	"net"
	"os"

	"gopkg.in/irc.v4"
)

func CreateClient() irc.ClientConfig {
	nick := os.Getenv("IRC_NICK")
	pass := os.Getenv("IRC_PASS")
	return irc.ClientConfig{
		Nick: nick,
		Pass: pass,
		Name: nick,
	}
}

func JoinIrcServer() (net.Conn, error) {
	return net.Dial("tcp", net.JoinHostPort("irc.ppy.sh", "6667"))
}

func InitNewClient() (*irc.Client, error) {
	conf := CreateClient()
	conn, err := JoinIrcServer()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to osu irc")
	}

	client := irc.NewClient(conn, conf)
	return client, nil
}
