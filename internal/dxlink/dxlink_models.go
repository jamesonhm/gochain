package dxlink

import (
	"encoding/json"
	"fmt"
)

type MsgType string
type ChannelContract string
type ChannelService string
type FeedDataFormat string

const (
	// Primary Message Types
	Error            MsgType = "ERROR"
	Setup            MsgType = "SETUP"
	AuthState        MsgType = "AUTH_STATE"
	Auth             MsgType = "AUTH"
	KeepAlive        MsgType = "KEEPALIVE"
	ChannelRequest   MsgType = "CHANNEL_REQUEST"
	ChannelOpened    MsgType = "CHANNEL_OPENED"
	FeedSetup        MsgType = "FEED_SETUP"
	FeedConfig       MsgType = "FEED_CONFIG"
	FeedSubscription MsgType = "FEED_SUBSCRIPTION"
	FeedData         MsgType = "FEED_DATA"
	ChannelCancel    MsgType = "CHANNEL_CANCEL"
	ChannelClosed    MsgType = "CHANNEL_CLOSED"
	// contract parameter
	ChannelHistory ChannelContract = "HISTORY"
	ChannelTicker  ChannelContract = "TICKER"
	ChannelStream  ChannelContract = "STREAM"
	ChannelAuto    ChannelContract = "AUTO"
	// channel service type
	FeedService ChannelService = "FEED"
	// Feed Data Format
	FullFormat    FeedDataFormat = "FULL"
	CompactFormat FeedDataFormat = "COMPACT"
)

// MessageCallback is a function type for handling received messages
type MessageCallback func(message []byte)

type ErrorMsg struct {
	Type    MsgType `json:"type"`
	Channel int     `json:"channel"`
	Error   string  `json:"error"`
	Message string  `json:"message"`
}

type SetupMsg struct {
	Type                   MsgType `json:"type"`
	Channel                int     `json:"channel"`
	KeepAliveTimeout       int     `json:"keepaliveTimeout"`
	AcceptKeepAliveTimeout int     `json:"acceptKeepaliveTimeout"`
	Version                string  `json:"version"`
}

type AuthStateMsg struct {
	Type    MsgType `json:"type"`
	Channel int     `json:"channel"`
	State   string  `json:"state"`
	UserID  string  `json:"userId,omitempty"`
}

type AuthMsg struct {
	Type    MsgType `json:"type"`
	Channel int     `json:"channel"`
	Token   string  `json:"token"`
}

type KeepAliveMsg struct {
	Type    MsgType `json:"type"`
	Channel int     `json:"channel"`
}

type ChannelReqRespMsg struct {
	Type       MsgType        `json:"type"`
	Channel    int            `json:"channel"`
	Service    ChannelService `json:"service"`
	Parameters Parameters     `json:"parameters"`
}

type Parameters struct {
	// Allowed values: "HISTORY", "TICKER", "STREAM", "AUTO"
	Contract ChannelContract `json:"contract"`
}

type FeedSetupMsg struct {
	Type    MsgType `json:"type"`
	Channel int     `json:"channel"`
	// aggregation perion in seconds
	AcceptAggregationPeriod int             `json:"acceptAggregationPeriod"`
	AcceptEventFields       FeedEventFields `json:"acceptEventFields"`
	AcceptDataFormat        FeedDataFormat  `json:"acceptDataFormat"`
}

type FeedSubscriptionMsg struct {
	Type    MsgType       `json:"type"`
	Channel int           `json:"channel"`
	Add     []FeedSubItem `json:"add,omitempty"`
	Remove  []FeedSubItem `json:"remove,omitempty"`
	// Remove all subs when true
	Reset bool `json:"reset,omitempty"`
}

type FeedSubItem struct {
	// FeedRegularSubscription, these are the base fields that are required
	// Type example vals: "Quote", "Trade", "Candle"
	// Symbol example vals: "*" wildcard for "STREAM" and "AUTO" contracts only, or "symbol"
	Type   string `json:"type"`
	Symbol string `json:"symbol"`
	// FeedOrderBookSubscription adds a source field,
	// see https://kb.dxfeed.com/en/data-model/market-events/qd-model-of-market-events.html#order-x
	Source string `json:"source,omitempty"`
	// FeedTimeSeriesSubscription adds a fromTime field, unix timestamp start time
	FromTime int64 `json:"fromTime,omitempty"`
}

type FeedConfigMsg struct {
	Type              MsgType         `json:"type"`
	Channel           int             `json:"channel"`
	AggregationPeriod float64         `json:"aggregationPeriod"`
	DataFormat        FeedDataFormat  `json:"dataFormat"`
	EventFields       FeedEventFields `json:"eventFields,omitempty"`
}

type FeedEventFields struct {
	Quote  []string `json:"Quote"`
	Trade  []string `json:"Trade,omitempty"`
	Candle []string `json:"Candle,omitempty"`
	Greeks []string `json:"Greeks,omitempty"`
}

type FeedDataMsg struct {
	Type    MsgType           `json:"type"`
	Channel int               `json:"channel"`
	Data    ProcessedFeedData `json:"data"`
}

type ProcessedFeedData struct {
	Quotes []QuoteData
	Trades []TradeData
}

// QuoteData represents quote data received from the server
type QuoteData struct {
	Symbol   string
	Type     string
	BidPrice float64
	AskPrice float64
}

type TradeData struct {
	Symbol string
	Type   string
	Price  float64
	Volume float64
}

func (d *ProcessedFeedData) UnmarshalJSON(data []byte) error {
	var content []interface{}
	if err := json.Unmarshal(data, &content); err != nil {
		return err
	}

	for i := 0; i < len(content); i += 2 {
		typeName, ok := content[i].(string)
		if !ok || i+1 >= len(content) {
			continue
		}

		values, ok := content[i+1].([]interface{})
		if !ok {
			continue
		}

		switch typeName {
		case "Quote":
			for j := 0; j < len(values); j += 4 {
				if j+3 > len(values) {
					break
				}
				symbol, ok := values[j].(string)
				evtType, ok := values[j+1].(string)
				askPrice, ok := values[j+2].(float64)
				bidPrice, ok := values[j+3].(float64)
				if !ok {
					return fmt.Errorf("unable to unmarshal Quote values")
				}

				quote := QuoteData{
					Symbol:   symbol,
					Type:     evtType,
					BidPrice: bidPrice,
					AskPrice: askPrice,
				}
				d.Quotes = append(d.Quotes, quote)
			}
		}

	}
	return nil
}
