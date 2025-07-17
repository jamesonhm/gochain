package tasty

import (
	"time"
)

type MarketDataParams struct {
	Crypto       []string `url:"cryptocurrency"`
	Equity       []string `url:"equity"`
	EquityOption []string `url:"equity-option"`
	Index        []string `url:"index"`
	Future       []string `url:"future"`
	FutureOption []string `url:"future-option"`
}

type MarketDataResponse struct {
	Data struct {
		Items []MarketData `json:"items"`
	} `json:"data"`
	Pagination interface{} `json:"pagination"`
}

type MarketData struct {
	Symbol             string    `json:"symbol"`
	InstrumentType     string    `json:"instrument-type"`
	UpdatedAt          time.Time `json:"updated-at"`
	Bid                string    `json:"bid"`
	BidSize            string    `json:"bid-size"`
	Ask                string    `json:"ask"`
	AskSize            string    `json:"ask-size"`
	Mid                string    `json:"mid"`
	Mark               string    `json:"mark"`
	Last               string    `json:"last"`
	LastMkt            string    `json:"last-mkt"`
	Open               string    `json:"open"`
	DayHighPrice       string    `json:"day-high-price"`
	DayLowPrice        string    `json:"day-low-price"`
	ClosePriceType     string    `json:"close-price-type"`
	PrevClose          string    `json:"prev-close"`
	PrevClosePriceType string    `json:"prev-close-price-type"`
	SummaryDate        string    `json:"summary-date"`
	PrevCloseDate      string    `json:"prev-close-date"`
	IsTradingHalted    bool      `json:"is-trading-halted"`
	HaltStartTime      int       `json:"halt-start-time"`
	HaltEndTime        int       `json:"halt-end-time"`
	YearLowPrice       string    `json:"year-low-price"`
	YearHighPrice      string    `json:"year-high-price"`
	Beta               string    `json:"beta,omitempty"`
	DividendAmount     string    `json:"dividend-amount,omitempty"`
	DividendFrequency  string    `json:"dividend-frequency,omitempty"`
	Close              string    `json:"close,omitempty"`
	LowLimitPrice      string    `json:"low-limit-price,omitempty"`
	HighLimitPrice     string    `json:"high-limit-price,omitempty"`
	Volume             string    `json:"volume,omitempty"`
}
