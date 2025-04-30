package dxlink

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type DxLinkClient struct {
	conn           *websocket.Conn
	url            string
	token          string
	subscriptions  map[string]bool
	mu             sync.Mutex
	connected      bool
	messageCounter int
	callbacks      map[string]MessageCallback
	ctx            context.Context
	cancel         context.CancelFunc
}

// MessageCallback is a function type for handling received messages
type MessageCallback func(message []byte)

type SetupMsg struct {
	Type                   string `json:"type"`
	Channel                int    `json:"channel"`
	KeepAliveTimeout       int    `json:"keepaliveTimeout"`
	AcceptKeepAliveTimeout int    `json:"acceptKeepaliveTimeout"`
	Version                string `json:"version"`
}

// QuoteData represents quote data received from the server
type QuoteData struct {
	EventSymbol string  `json:"eventSymbol"`
	BidPrice    float64 `json:"bidPrice"`
	AskPrice    float64 `json:"askPrice"`
	LastPrice   float64 `json:"lastPrice"`
	Volume      int64   `json:"volume"`
}

func New(url string) *DxLinkClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &DxLinkClient{
		url:           url,
		subscriptions: make(map[string]bool),
		callbacks:     make(map[string]MessageCallback),
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (c *DxLinkClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return fmt.Errorf("client already connected")
	}

	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("dial error: %w", err)
	}
	c.conn = conn
	c.connected = true

	// Start message handler
	go c.handleMessages()

	setupMsg := SetupMsg{
		Type:                   "SETUP",
		Channel:                0,
		KeepAliveTimeout:       60,
		AcceptKeepAliveTimeout: 60,
		Version:                "0.1-golang",
	}

	err = c.sendMessage(setupMsg)
	if err != nil {
		c.connected = false
		c.conn.Close()
		return fmt.Errorf("failed to send setup message: %w", err)
	}

	return nil
}

func (c *DxLinkClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return fmt.Errorf("client not connected")
	}

	c.cancel()

	disconnectMsg := map[string]interface{}{
		"type": "DISCONNECT",
	}

	err := c.sendMessage(disconnectMsg)
	if err != nil {
		slog.Error("Error sending disconnect message", "err", err)
	}

	err = c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		slog.Error("Error sending close message", "err", err)
	}

	c.connected = false
	err = c.conn.Close()
	if err != nil {
		return fmt.Errorf("error closing connection: %w", err)
	}

	return nil
}

func (c *DxLinkClient) sendMessage(msg interface{}) error {
	if c.conn == nil {
		return fmt.Errorf("unable to send message, no connection")
	}

	c.messageCounter++
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error marshaling message: %w", err)
	}
	fd, _ := json.MarshalIndent(msg, "", "  ")
	fmt.Printf("sent message: %s", string(fd))

	err = c.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}

	return nil
}

// handleMessages reads and processes incoming
func (c *DxLinkClient) handleMessages() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			if c.conn == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			_, message, err := c.conn.ReadMessage()
			if err != nil {
				slog.Error("Error reading message", "err", err)
				//TODO: try reconnect
				continue
			}

			go c.processMessage(message)
		}
	}
}

func (c *DxLinkClient) processMessage(message []byte) {
	var msgMap map[string]interface{}
	if err := json.Unmarshal(message, &msgMap); err != nil {
		slog.Error("Error unmarshaling message", "err", err)
		return
	}

	fmt.Println("from process message handler")
	fd, _ := json.MarshalIndent(message, "", "  ")
	fmt.Println(string(fd))

	msgType, ok := msgMap["type"].(string)
	if !ok {
		if _, hasData := msgMap["data"]; hasData {
			c.handleDataMessage(message)
			return
		}
		slog.Info("Unknown message format", "msg", string(message))
		return
	}

	switch msgType {
	case "CONNECTED":
		slog.Info("Connected to the DxLink server")
	case "AUTHENTICATED":
		slog.Info("Successfully authenticated")
	case "FEED_SUBSCRIBED":
		slog.Info("Successfully subscribed to feed")
	case "ERROR":
		slog.Info("Error from server", "err", msgMap)
	default:
		c.mu.Lock()
		callback, exists := c.callbacks[msgType]
		c.mu.Unlock()

		if exists {
			callback(message)
		} else {
			c.handleDataMessage(message)
		}
	}
}

func (c *DxLinkClient) handleDataMessage(message []byte) {
	var dataMap map[string]interface{}
	if err := json.Unmarshal(message, &dataMap); err != nil {
		slog.Error("Error unmarshaling data message", "err", err)
		return
	}

	fmt.Println("from data message handler")
	fd, _ := json.MarshalIndent(message, "", "  ")
	fmt.Println(string(fd))
}
