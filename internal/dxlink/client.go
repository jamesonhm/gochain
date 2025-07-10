package dxlink

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	//"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jamesonhm/gochain/internal/dt"
)

type DxLinkClient struct {
	conn           *websocket.Conn
	url            string
	token          string
	optionSubs     map[string]*OptionData
	underlyingSubs map[string]*UnderlyingData
	mu             sync.RWMutex
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
	today, err := dt.EndOfDay(time.Now())
	if err != nil {
		return err
	}
	cut_date := today.AddDate(0, 0, days)
	c.underlyingSubs[symbol] = NewUnderlying()
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
					Trade: []string{"eventType", "eventSymbol", "price", "size"},
					//Candle: []string{"eventType", "eventSymbol", "time", "open", "high", "low", "close", "volume", "impVolatility", "openInterest"},
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
		//slog.Info("SERVER <-", "", resp)

		c.mu.Lock()
		defer c.mu.Unlock()
		switch resp.Channel {
		case 1:
			if len(resp.Data.Trades) > 0 {
				slog.Info("SERVER <-", "symbol", resp.Data.Trades[0].Symbol, "trades", resp.Data.Trades)
				for _, trade := range resp.Data.Trades {
					if _, ok := c.underlyingSubs[trade.Symbol]; !ok {
						c.underlyingSubs[trade.Symbol] = NewUnderlying()
					}
					c.underlyingSubs[trade.Symbol].Trade = trade
				}
			}
			//if len(resp.Data.Candles) > 0 {
			//	symbol := resp.Data.Candles[0].Symbol[0:strings.Index(resp.Data.Candles[0].Symbol, "{")]
			//	//if symbol == "VIX" {
			//	//	for _, c := range resp.Data.Candles {
			//	//		fmt.Println(time.UnixMilli(int64(*c.Time)))
			//	//	}
			//	//}
			//	slog.Info(
			//		"SERVER <-",
			//		"symbol",
			//		symbol,
			//		"start",
			//		time.UnixMilli(int64(*resp.Data.Candles[0].Time)),
			//		"startValClose",
			//		resp.Data.Candles[0].Close,
			//		"end",
			//		time.UnixMilli(int64(*resp.Data.Candles[len(resp.Data.Candles)-1].Time)),
			//		"endValClose",
			//		resp.Data.Candles[len(resp.Data.Candles)-1].Close,
			//	)

			//	if _, ok := c.underlyingSubs[symbol]; !ok {
			//		c.underlyingSubs[symbol] = NewUnderlying()
			//	}
			//	for _, candle := range resp.Data.Candles {
			//		c.underlyingSubs[symbol].Candles[int64(*candle.Time)] = candle
			//	}
			//copy(c.underlyingSubs[symbol].Candles, resp.Data.Candles)
			//}
		case 3:
			if len(resp.Data.Quotes) > 0 {
				slog.Info("SERVER <-", "symbol", resp.Data.Quotes[0].Symbol, "quotes", resp.Data.Quotes)
				for _, quote := range resp.Data.Quotes {
					c.optionSubs[quote.Symbol].Quote = quote
				}
			}
			if len(resp.Data.Greeks) > 0 {
				slog.Info("SERVER <-", "symbol", resp.Data.Greeks[0].Symbol, "greeks", resp.Data.Greeks)
				for _, greek := range resp.Data.Greeks {
					c.optionSubs[greek.Symbol].Greek = greek
				}
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

//func (c *DxLinkClient) VixONMove() (float64, error) {
//	c.mu.RLock()
//	defer c.mu.RUnlock()
//	vix, ok := c.underlyingSubs["VIX"]
//	if !ok {
//		return 0, fmt.Errorf("no VIX underlying data")
//	}
//	candles := vix.Candles
//	if len(candles) < 2 {
//		slog.Info("VIX", "Candles", candles)
//		return 0, fmt.Errorf("not enough VIX candles")
//	}
//	now := time.Now()
//	previous_dt := dt.PreviousWeekday(
//		time.Date(
//			now.Year(),
//			now.Month(),
//			now.Day(),
//			18, 0, 0, 0,
//			time.Local,
//		),
//	)
//	previous_ts := previous_dt.UnixMilli()
//	var prev_candle CandleEvent
//	if prev_candle, ok = candles[previous_ts]; !ok {
//		return 0, fmt.Errorf("no candle found for previous day: %d, %s", previous_ts, previous_dt)
//	}
//	var curr_candle CandleEvent
//	for ts, c := range candles {
//		if ts > previous_ts {
//			curr_candle = c
//		}
//	}
//	if curr_candle.EventType == "" {
//		return 0, fmt.Errorf("no candle found for current day")
//	}
//	return *curr_candle.Open - *prev_candle.Close, nil
//}

func (c *DxLinkClient) underlyingFeedSub() FeedSubscriptionMsg {
	feedSub := FeedSubscriptionMsg{
		Type:    FeedSubscription,
		Channel: 1,
		Reset:   true,
		Add:     []FeedSubItem{},
	}
	//fromTime := time.Now().AddDate(0, 0, -3).UnixMilli()

	for under := range c.underlyingSubs {
		//candle_symbol := under + "{=30m}"
		//feedSub.Add = append(feedSub.Add, FeedSubItem{Type: "Quote", Symbol: under})
		//feedSub.Add = append(feedSub.Add, FeedSubItem{Type: "Candle", Symbol: candle_symbol, FromTime: fromTime})
		feedSub.Add = append(feedSub.Add, FeedSubItem{Type: "Trade", Symbol: under})
	}
	//feedSub.Add = append(
	//	feedSub.Add,
	//	FeedSubItem{
	//		Type:     "Candle",
	//		Symbol:   "VIX{=1d}",
	//		FromTime: fromTime,
	//	},
	//)
	//feedSub.Add = append(
	//	feedSub.Add,
	//	FeedSubItem{
	//		Type:   "Trade",
	//		Symbol: "VIX",
	//	},
	//)
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
