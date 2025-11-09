package main

import (
	"fmt"
	"net"

	"gopkg.in/irc.v4"
)

func CreateClient() irc.ClientConfig {
	return irc.ClientConfig{
		Nick: "Chiffa",
		Pass: "",
		Name: "Chiffa",
	}
}

func JoinIrcServer() (net.Conn, error) {
	return net.Dial("tcp", net.JoinHostPort("irc.ppy.sh", "6667"))
}

func InitNewClient() *irc.Client {
	conf := CreateClient()
	conn, err := JoinIrcServer()
	if err != nil {
		fmt.Errorf("failed to connect to osu irc")
	}

	client := irc.NewClient(conn, conf)
	return client
}
