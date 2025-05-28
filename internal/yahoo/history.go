package yahoo

import (
	"context"
	"net/http"
)

const (
	HistoryPath = "/v1/markets/stock/history"
)

type HistoryParams struct {
	Symbol        string `url:"symbol"`
	Interval      string `url:"interval"`
	DiffAndSplits bool   `url:"diffandsplits"`
}

type HistoryResponse struct {
	Meta  HistoryMeta         `json:"meta"`
	Body  map[string]OHLCItem `json:"body"`
	Error string              `json:"error"`
}

type HistoryMeta struct {
	Currency             string  `json:"currency"`
	Symbol               string  `json:"symbol"`
	ExchangeName         string  `json:"exchangeName"`
	InstrumentType       string  `json:"instrumentName"`
	FirstTradeDate       int64   `json:"firstTradeDate"`
	RegularMarketTime    int64   `json:"regularMarketTime"`
	Gmtoffset            int64   `json:"gmtoffset"`
	Timezone             string  `json:"timezone"`
	ExchangeTimezoneName string  `json:"exchangeTimezoneName"`
	RegularMarketPrice   float64 `json:"regularMarketPrice"`
	ChartPreviousClose   float64 `json:"chartPreviousClose"`
	PreviousClose        float64 `json:"previousClose"`
	Scale                int64   `json:"scale"`
	PriceHint            int64   `json:"priceHint"`
	DataGranularity      string  `json:"dataGranularity"`
	Range                string  `json:"range"`
}

type OHLCItem struct {
	Date    string  `json:"date"`
	DateUtc int64   `json:"date_utc"`
	Open    float64 `json:"open"`
	High    float64 `json:"high"`
	Low     float64 `json:"low"`
	Close   float64 `json:"close"`
	Volume  float64 `json:"volume"`
}

func (c *YahooAPI) GetOHLCHistory(ctx context.Context, params *HistoryParams) (*HistoryResponse, error) {
	res := &HistoryResponse{}
	path := c.baseurl + HistoryPath
	err := c.request(ctx, http.MethodGet, path, params, nil, res)
	return res, err
}
