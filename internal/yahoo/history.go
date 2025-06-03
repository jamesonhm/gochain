package yahoo

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jamesonhm/gochain/internal/dt"
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
	Meta  HistoryMeta        `json:"meta"`
	Body  map[int64]OHLCItem `json:"body"`
	Error string             `json:"error"`
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

func (c *YahooAPI) ONMove(symbol string) (float64, error) {
	reqNewData := false
	midnight := dt.Midnight(time.Now()).Unix()
	if hist, ok := c.cache[symbol]; !ok {
		reqNewData = true
	} else if hist.Meta.RegularMarketTime < midnight {
		reqNewData = true
	}
	if reqNewData {
		ctx := context.TODO()
		histParams := HistoryParams{
			Symbol:        symbol,
			Interval:      "1d",
			DiffAndSplits: false,
		}
		res, err := c.GetOHLCHistory(ctx, &histParams)
		if err != nil {
			return 0, err
		}
		c.cache[symbol] = res
	}
	var currOpen, prevClose float64
	var prevTS int64 = 0

	for ts, ohlc := range c.cache[symbol].Body {
		if ts > midnight {
			currOpen = ohlc.Open
			fmt.Printf("Current opent %.2f at TS %d, %s\n", currOpen, ts, time.Unix(ts, 0))
		}
		if ts > prevTS && ts < midnight {
			prevTS = ts
		}
	}
	prevClose = c.cache[symbol].Body[prevTS].Close
	fmt.Printf("Prev Close %.2f at TS %d, %s\n", prevClose, prevTS, time.Unix(prevTS, 0))

	return currOpen - prevClose, nil
}
