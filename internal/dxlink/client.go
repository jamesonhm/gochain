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
	optionSubs     map[string]*OptionData
	underlyingSubs map[string]*UnderlyingData
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
		url:            url,
		optionSubs:     make(map[string]*OptionData),
		underlyingSubs: make(map[string]*UnderlyingData),
		callbacks:      make(map[string]MessageCallback),
		ctx:            ctx,
		cancel:         cancel,
		token:          token,
	}
}

func (c *DxLinkClient) ResetData() {
	clear(c.optionSubs)
	clear(c.underlyingSubs)
}

func (c *DxLinkClient) UpdateOptionSubs(symbol string, options []string, days int) error {
	today, err := endOfDay(time.Now())
	if err != nil {
		return err
	}
	cut_date := today.AddDate(0, 0, days)
	c.underlyingSubs[symbol] = &UnderlyingData{}
	for _, option := range options {
		opt, err := ParseOption(option)
		if err != nil {
			return err
		}
		if opt.Date.After(cut_date) {
			continue
		}
		c.optionSubs[option] = &OptionData{}
		fmt.Printf("Added %s to subs\n", option)
	}
	return nil
}

func (c *DxLinkClient) Connect() error {
	//c.mu.Lock()
	//defer c.mu.Unlock()

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

	err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
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
	c.mu.Lock()
	defer c.mu.Unlock()

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
			// setup a channel for underlying (indices, equities) and options
			chanReq := ChannelReqRespMsg{
				Type:    ChannelRequest,
				Channel: 1,
				Service: FeedService,
				Parameters: Parameters{
					Contract: ChannelAuto,
				},
			}
			c.sendMessage(chanReq)

			// setup a channel for options
			chanReq = ChannelReqRespMsg{
				Type:    ChannelRequest,
				Channel: 3,
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
		// TODO: add eventTime to quote for age validation
		var feedSetup FeedSetupMsg
		if resp.Channel == 1 {
			// UNDERLYING
			feedSetup = FeedSetupMsg{
				Type:                    FeedSetup,
				Channel:                 1,
				AcceptAggregationPeriod: 60,
				AcceptDataFormat:        CompactFormat,
				AcceptEventFields: FeedEventFields{
					//Quote: []string{"eventType", "eventSymbol", "bidPrice", "askPrice"},
					Trade:  []string{"eventType", "eventSymbol", "price", "size"},
					Candle: []string{"eventType", "eventSymbol", "time", "open", "high", "low", "close", "volume", "impVolatility"},
				},
			}
		} else if resp.Channel == 3 {
			// OPTIONS
			feedSetup = FeedSetupMsg{
				Type:                    FeedSetup,
				Channel:                 3,
				AcceptAggregationPeriod: 60,
				AcceptDataFormat:        CompactFormat,
				AcceptEventFields: FeedEventFields{
					Quote:  []string{"eventType", "eventSymbol", "bidPrice", "askPrice"},
					Greeks: []string{"eventType", "eventSymbol", "price", "volatility", "delta", "gamma", "theta", "rho", "vega"},
				},
			}
		}
		c.sendMessage(feedSetup)
	case string(FeedConfig):
		resp := FeedConfigMsg{}
		err := json.Unmarshal(message, &resp)
		if err != nil {
			slog.Error("unable to unmarshal feed config msg", "err", err)
			fmt.Printf("%s\n\n", string(message))
			return
		}
		slog.Info("SERVER <-", "", resp)
		var feedSub FeedSubscriptionMsg
		if resp.Channel == 1 {
			feedSub = c.underlyingFeedSub()
		} else if resp.Channel == 3 {
			fmt.Println("Channel 3 response, getting option feed subs")
			feedSub = c.optionFeedSub()
		}
		c.sendMessage(feedSub)
	case string(FeedData):
		resp := FeedDataMsg{}
		err := json.Unmarshal(message, &resp)
		if err != nil {
			slog.Error("unable to unmarshal feed data msg", "err", err)
			fmt.Printf("%s\n", string(message))
			return
		}
		slog.Info("SERVER <-", "", resp)

		c.mu.Lock()
		defer c.mu.Unlock()
		switch resp.Channel {
		case 1:
			for _, trade := range resp.Data.Trades {
				c.underlyingSubs[trade.Symbol].Trade = trade
			}
		case 3:
			for _, quote := range resp.Data.Quotes {
				c.optionSubs[quote.Symbol].Quote = quote
			}
			for _, greek := range resp.Data.Greeks {
				c.optionSubs[greek.Symbol].Greek = greek
			}
		}
	case string(Error):
		resp := ErrorMsg{}
		err := json.Unmarshal(message, &resp)
		if err != nil {
			slog.Error("unable to unmarshal error msg", "err", err)
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

func (c *DxLinkClient) underlyingFeedSub() FeedSubscriptionMsg {
	feedSub := FeedSubscriptionMsg{
		Type:    FeedSubscription,
		Channel: 1,
		Reset:   true,
		Add:     []FeedSubItem{},
	}
	fromTime := time.Now().AddDate(0, 0, -2).Unix()

	for under := range c.underlyingSubs {
		candle_symbol := under + "{=30m}"
		//feedSub.Add = append(feedSub.Add, FeedSubItem{Type: "Quote", Symbol: under})
		feedSub.Add = append(feedSub.Add, FeedSubItem{Type: "Candle", Symbol: candle_symbol, FromTime: fromTime})
		feedSub.Add = append(feedSub.Add, FeedSubItem{Type: "Trade", Symbol: under})
	}
	return feedSub
}

func (c *DxLinkClient) optionFeedSub() FeedSubscriptionMsg {
	feedSub := FeedSubscriptionMsg{
		Type:    FeedSubscription,
		Channel: 3,
		Reset:   true,
		Add:     []FeedSubItem{},
	}
	for opt := range c.optionSubs {
		slog.Info("OptionFeedSub method", "OptionSub iter:", opt)
		feedSub.Add = append(feedSub.Add, FeedSubItem{Type: "Quote", Symbol: opt})
		feedSub.Add = append(feedSub.Add, FeedSubItem{Type: "Greeks", Symbol: opt})
	}
	return feedSub
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

func endOfDay(date time.Time) (*time.Time, error) {
	nytz, err := time.LoadLocation("America/New_York")
	if err != nil {
		return nil, err
	}
	end := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, nytz)
	return &end, nil
}
