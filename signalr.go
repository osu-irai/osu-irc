package main

import (
	"context"
	"fmt"
	"log"

	"github.com/philippseith/signalr"
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
type Client struct {
	conn    signalr.Client
	channel chan<- RequestWithTarget
}

func NewClient(url string, channel chan<- RequestWithTarget) (*Client, error) {
	ctx := context.Background()
	connection, err := signalr.NewHTTPConnection(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to create signalr connection")
	}
	conn, err := signalr.NewClient(ctx,
		signalr.WithConnection(connection), signalr.WithReceiver(&Receiver{channel: channel}), signalr.Logger(nil, false))

	if err != nil {
		return nil, fmt.Errorf("failed to create a signalr client")
	}
	log.Printf("created a client")
	return &Client{
		conn:    conn,
		channel: channel,
	}, nil
}

func (c *Client) Connect(ctx context.Context) error {
	c.conn.Start()
	return nil
}

func (c *Client) Disconnect() error {
	c.conn.Stop()
	close(c.channel)
	return nil
}

func (c *Client) JoinUserGroup(ctx context.Context) error {
	result := <-c.conn.Invoke("JoinServiceGroup", "osubot")

	if result.Error != nil {
		return fmt.Errorf("failed to join user group")
	}
	log.Printf("joined service usergroup")
	return nil
}

type Receiver struct {
	signalr.Receiver
	channel chan<- RequestWithTarget
}

func (r *Receiver) ReceiveFullRequest(req RequestWithTarget) {
	log.Printf("%v received %v - %v", req.Target, req.Request.Beatmap.Artist, req.Request.Beatmap.Title)
	r.channel <- req
}
