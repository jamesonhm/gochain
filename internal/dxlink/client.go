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

func New(ctx context.Context, url string, token string) *DxLinkClient {
	ctx, cancel := context.WithCancel(ctx)
	return &DxLinkClient{
		url:           url,
		subscriptions: make(map[string]bool),
		callbacks:     make(map[string]MessageCallback),
		ctx:           ctx,
		cancel:        cancel,
		token:         token,
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

	slog.Info("CLIENT ->", "", msg)
	//fd, _ := json.MarshalIndent(msg, "", "  ")
	//fmt.Printf("sent message: %s\n", string(fd))

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
				// TODO: try reconnect
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

	msgType, ok := msgMap["type"].(string)
	if !ok {
		slog.Info("Unknown message format", "msg", string(message))
		return
	}

	switch msgType {
	case string(Setup):
		resp := SetupMsg{}
		err := json.Unmarshal(message, &resp)
		if err != nil {
			slog.Error("unable to unmarshal setup msg")
			return
		}
		slog.Info("SERVER <-", "", resp)
	case string(AuthState):
		resp := AuthStateMsg{}
		err := json.Unmarshal(message, &resp)
		if err != nil {
			slog.Error("unable to unmarshal auth state msg")
			return
		}
		slog.Info("SERVER <-", "", resp)
		if resp.State == "UNAUTHORIZED" {
			authMsg := AuthMsg{
				Type:    Auth,
				Channel: 0,
				Token:   c.token,
			}
			c.sendMessage(authMsg)
		} else if resp.State == "AUTHORIZED" {
			slog.Info("this is where we look for a callback to setup the new channels and feeds")
			chanReq := ChannelReqRespMsg{
				Type:    ChannelRequest,
				Channel: 1,
				Service: FeedService,
				Parameters: Parameters{
					Contract: ChannelAuto,
				},
			}
			c.sendMessage(chanReq)
		}
	case string(ChannelOpened):
		resp := ChannelReqRespMsg{}
		err := json.Unmarshal(message, &resp)
		if err != nil {
			slog.Error("unable to unmarshal channel open msg")
			return
		}
		slog.Info("SERVER <-", "", resp)
		feedSetup := FeedSetupMsg{
			Type:                    FeedSetup,
			Channel:                 1,
			AcceptAggregationPeriod: 10,
			AcceptDataFormat:        CompactFormat,
			AcceptEventFields: FeedEventFields{
				Quote: []string{"eventType", "eventSymbol", "bidPrice", "askPrice"},
			},
		}
		c.sendMessage(feedSetup)
	case string(FeedConfig):
		resp := FeedConfigMsg{}
		err := json.Unmarshal(message, &resp)
		if err != nil {
			slog.Error("unable to unmarshal feed config msg")
			fmt.Printf("%s\n", string(message))
			return
		}
		slog.Info("SERVER <-", "", resp)
		feedSub := FeedSubscriptionMsg{
			Type:    FeedSubscription,
			Channel: 1,
			Reset:   true,
			Add: []FeedSubItem{
				{
					Type:   "Trade",
					Symbol: "SPY",
				},
				{
					Type:   "Trade",
					Symbol: "XSP",
				},
			},
		}
		c.sendMessage(feedSub)
	case string(Error):
		resp := ErrorMsg{}
		err := json.Unmarshal(message, &resp)
		if err != nil {
			slog.Error("unable to unmarshal error msg")
			return
		}
		slog.Info("SERVER <-", "", resp)
	default:
		//c.mu.Lock()
		//callback, exists := c.callbacks[msgType]
		//c.mu.Unlock()
		//if exists {
		//	callback(message)
		//} else {
		//	c.handleDataMessage(message)
		//}
		//c.handleDataMessage(message)
		slog.Info("Unknown message type", "msg", string(message))
	}
}

func pprint(msg any) {
	fd, _ := json.MarshalIndent(msg, "", "  ")
	fmt.Println(string(fd))
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
