package tasty

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

const (
	SandboxURL = "wss://streamer.cert.tastyworks.com"
	ProdURL    = "wss://streamer.tastyworks.com"
)

type AccountStreamer struct {
	accts          []string
	conn           *websocket.Conn
	url            string
	token          string
	connected      bool
	messageCounter int
	ctx            context.Context
	cancel         context.CancelFunc
	retries        int
	delay          time.Duration
	expBackoff     bool
}

type ActionMsg struct {
	Action    string   `json:"action"`
	Value     []string `json:"value,omitempty"`
	AuthToken string   `json:"auth-token"`
	RequestID int      `json:"request-id"`
}

type ConnectRespMsg struct {
	Status      string   `json:"status"`
	Action      string   `json:"action"`
	WSSessionID string   `json:"web-socket-session-id"`
	Value       []string `json:"value,omitempty"`
	RequestID   int      `json:"request-id"`
}

func (c *TastyAPI) NewAccountStreamer(ctx context.Context, accts []string, prod bool) *AccountStreamer {
	ctx, cancel := context.WithCancel(ctx)
	var url string
	if prod {
		url = ProdURL
	} else {
		url = SandboxURL
	}
	return &AccountStreamer{
		accts:          accts,
		ctx:            ctx,
		cancel:         cancel,
		delay:          1 * time.Second,
		expBackoff:     false,
		retries:        3,
		token:          *c.session.Data.SessionToken,
		url:            url,
		messageCounter: 1,
	}
}

func (as *AccountStreamer) Connect() error {
	if as.connected {
		return fmt.Errorf("client already connected")
	}

	u, err := url.Parse(as.url)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("dial error for url: %s : %w", u.String(), err)
	}
	as.conn = conn
	as.connected = true

	// Start message handler
	go as.handleMessages()
	go as.keepAlive()

	setupMsg := ActionMsg{
		Action:    "connect",
		Value:     []string{},
		AuthToken: as.token,
		RequestID: as.messageCounter,
	}

	err = as.sendMessage(setupMsg)
	if err != nil {
		as.connected = false
		as.conn.Close()
		return fmt.Errorf("failed to send setup message: %w", err)
	}

	return nil
}

func (as *AccountStreamer) Close() error {
	if !as.connected {
		return fmt.Errorf("client not connected")
	}

	as.cancel()

	err := as.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		slog.Error("Error sending close message", "err", err)
	}

	as.connected = false
	err = as.conn.Close()
	if err != nil {
		return fmt.Errorf("error closing connection: %w", err)
	}

	return nil
}

func (as *AccountStreamer) sendMessage(msg ActionMsg) error {
	if as.conn == nil {
		return fmt.Errorf("unable to send message, no connection")
	}

	as.messageCounter++
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error marshaling message: %w", err)
	}

	slog.Info("ACCT STREAMER ->", "", msg)

	err = as.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}

	return nil
}

func (as *AccountStreamer) keepAlive() {
	for {
		select {
		case <-as.ctx.Done():
			return
		default:
			time.Sleep(30 * time.Second)
			as.sendMessage(ActionMsg{
				Action:    "heartbeat",
				AuthToken: as.token,
				RequestID: as.messageCounter,
			})
		}
	}
}

// handleMessages reads and processes incoming
func (as *AccountStreamer) handleMessages() {
	for {
		select {
		case <-as.ctx.Done():
			return
		default:
			if as.conn == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			_, message, err := as.conn.ReadMessage()
			if err != nil {
				slog.Error("Error reading message", "err", err)
				continue
			}

			go as.processMessage(message)
		}
	}
}

func (as *AccountStreamer) processMessage(message []byte) {
	var msgMap map[string]interface{}
	if err := json.Unmarshal(message, &msgMap); err != nil {
		slog.Error("Error unmarshaling message", "err", err)
		return
	}

	var msgType string
	var ok bool
	msgType, ok = msgMap["type"].(string)
	if !ok {
		msgType, ok = msgMap["action"].(string)
		if !ok {
			slog.Info("Unknown message format", "msg", string(message))
			return
		}
	}

	switch msgType {
	case "connect":
		resp := ConnectRespMsg{}
		err := json.Unmarshal(message, &resp)
		if err != nil {
			slog.Error("unable to unmarshal", "connect msg", string(message))
			return
		}
		slog.Info("ACCT STREAMER <-", "", resp)
	case "heartbeat":
		resp := ConnectRespMsg{}
		err := json.Unmarshal(message, &resp)
		if err != nil {
			slog.Error("unable to unmarshal", "heartbeat msg", string(message))
			return
		}
		slog.Info("ACCT STREAMER <-", "", resp)
	}
}
