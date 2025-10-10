package yahoo

import (
	"context"
	"fmt"
	"time"
)

const (
	QuotesPath = "/v1/markets/stock/quotes"
)

type QuotesParams struct {
	Symbol string `url:"symbol"`
}

type QuotesResponse struct {
	Meta QuotesMeta  `json:"meta"`
	Body []QuoteItem `json:"body"`
}

type QuotesMeta struct {
	Version       string    `json:"version"`
	Status        int       `json:"status"`
	Copywrite     string    `json:"copywrite"`
	Symbol        string    `json:"symbol"`
	ProcessedTime time.Time `json:"processedTime"`
}

type QuoteItem struct {
	Language                          string  `json:"language"`
	Region                            string  `json:"region"`
	QuoteType                         string  `json:"quoteType"`
	TypeDisp                          string  `json:"typeDisp"`
	QuoteSourceName                   string  `json:"quoteSourceName"`
	Triggerable                       bool    `json:"triggerable"`
	CustomPriceAlertConfidence        string  `json:"customPriceAlertConfidence"`
	Currency                          string  `json:"currency"`
	RegularMarketTime                 int     `json:"regularMarketTime"`
	UnderlyingSymbol                  string  `json:"underlyingSymbol"`
	Exchange                          string  `json:"exchange"`
	MessageBoardID                    string  `json:"messageBoardId"`
	ExchangeTimezoneName              string  `json:"exchangeTimezoneName"`
	ExchangeTimezoneShortName         string  `json:"exchangeTimezoneShortName"`
	GmtOffSetMilliseconds             int     `json:"gmtOffSetMilliseconds"`
	Market                            string  `json:"market"`
	EsgPopulated                      bool    `json:"esgPopulated"`
	RegularMarketChangePercent        float64 `json:"regularMarketChangePercent"`
	RegularMarketPrice                float64 `json:"regularMarketPrice"`
	HasPrePostMarketData              bool    `json:"hasPrePostMarketData"`
	FirstTradeDateMilliseconds        int64   `json:"firstTradeDateMilliseconds"`
	PriceHint                         int     `json:"priceHint"`
	RegularMarketChange               float64 `json:"regularMarketChange"`
	RegularMarketDayHigh              float64 `json:"regularMarketDayHigh"`
	RegularMarketDayRange             string  `json:"regularMarketDayRange"`
	RegularMarketDayLow               float64 `json:"regularMarketDayLow"`
	RegularMarketVolume               int     `json:"regularMarketVolume"`
	RegularMarketPreviousClose        float64 `json:"regularMarketPreviousClose"`
	Bid                               float64 `json:"bid"`
	Ask                               float64 `json:"ask"`
	BidSize                           int     `json:"bidSize"`
	AskSize                           int     `json:"askSize"`
	FullExchangeName                  string  `json:"fullExchangeName"`
	RegularMarketOpen                 float64 `json:"regularMarketOpen"`
	AverageDailyVolume3Month          int     `json:"averageDailyVolume3Month"`
	AverageDailyVolume10Day           int     `json:"averageDailyVolume10Day"`
	FiftyTwoWeekLowChange             float64 `json:"fiftyTwoWeekLowChange"`
	FiftyTwoWeekRange                 string  `json:"fiftyTwoWeekRange"`
	FiftyTwoWeekHighChange            float64 `json:"fiftyTwoWeekHighChange"`
	FiftyTwoWeekHighChangePercent     float64 `json:"fiftyTwoWeekHighChangePercent"`
	FiftyTwoWeekLow                   float64 `json:"fiftyTwoWeekLow"`
	FiftyTwoWeekHigh                  float64 `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekChangePercent         float64 `json:"fiftyTwoWeekChangePercent"`
	MarketState                       string  `json:"marketState"`
	ShortName                         string  `json:"shortName"`
	LongName                          string  `json:"longName"`
	FiftyDayAverage                   float64 `json:"fiftyDayAverage"`
	FiftyDayAverageChange             float64 `json:"fiftyDayAverageChange"`
	FiftyDayAverageChangePercent      float64 `json:"fiftyDayAverageChangePercent"`
	TwoHundredDayAverage              float64 `json:"twoHundredDayAverage"`
	TwoHundredDayAverageChange        float64 `json:"twoHundredDayAverageChange"`
	TwoHundredDayAverageChangePercent float64 `json:"twoHundredDayAverageChangePercent"`
	SourceInterval                    int     `json:"sourceInterval"`
	ExchangeDataDelayedBy             int     `json:"exchangeDataDelayedBy"`
	Tradeable                         bool    `json:"tradeable"`
	CryptoTradeable                   bool    `json:"cryptoTradeable"`
	Symbol                            string  `json:"symbol"`
}

func (c *YahooAPI) getQuote(
	ctx context.Context,
	params *QuotesParams,
	cache_lifetime time.Duration,
) (*QuotesResponse, error) {
	res := &QuotesResponse{}
	path := c.baseurl + QuotesPath
	err := c.cachedRequest(ctx, path, params, res, cache_lifetime)
	return res, err
}

func (c *YahooAPI) ONMove(symbol string) (float64, error) {
	const cache_lifetime = 8 * time.Hour

	quoteParams := QuotesParams{
		Symbol: symbol,
	}
	ctx := context.TODO()
	res, err := c.getQuote(ctx, &quoteParams, cache_lifetime)
	if err != nil {
		return 0, err
	}

	currOpen := res.Body[0].RegularMarketOpen
	fmt.Printf("Current opent %.2f\n", currOpen)
	prevClose := res.Body[0].RegularMarketPreviousClose
	fmt.Printf("Prev Close %.2f\n", prevClose)

	return currOpen - prevClose, nil
}
