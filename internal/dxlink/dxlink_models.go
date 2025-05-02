package dxlink

//import "log/slog"

type MsgType string
type ChannelContract string
type ChannelService string

const (
	// Primary Message Types
	Error            MsgType = "ERROR"
	Setup            MsgType = "SETUP"
	AuthState        MsgType = "AUTH_STATE"
	Auth             MsgType = "AUTH"
	ChannelRequest   MsgType = "CHANNEL_REQUEST"
	ChannelOpened    MsgType = "CHANNEL_OPENED"
	FeedSubscription MsgType = "FEED_SUBSCRIPTION"
	FeedConfig       MsgType = "FEED_CONFIG"
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

// QuoteData represents quote data received from the server
type QuoteData struct {
	EventSymbol string  `json:"eventSymbol"`
	BidPrice    float64 `json:"bidPrice"`
	AskPrice    float64 `json:"askPrice"`
	LastPrice   float64 `json:"lastPrice"`
	Volume      int64   `json:"volume"`
}
