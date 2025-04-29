package dxlink

import (
	"context"
	"fmt"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Client struct {
	url       string
	token     string
	createdAt time.Time
}

type SetupMsg struct {
	Type                   string `json:"type"`
	Channel                int    `json:"channel"`
	KeepAliveTimeout       int    `json:"keepaliveTimeout"`
	AcceptKeepAliveTimeout int    `json:"acceptKeepaliveTimeout"`
	Version                string `json:"version"`
}

func New(ctx context.Context, url string, token string) {
	c, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	setup := SetupMsg{
		Type:                   "SETUP",
		Channel:                0,
		KeepAliveTimeout:       60,
		AcceptKeepAliveTimeout: 60,
		Version:                "0.1-golang",
	}
	err = wsjson.Write(ctx, c, setup)
}
func (c *Client) Connect(url string, token string)
