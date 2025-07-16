package tasty

import (
	"context"
	"net/http"
)

const (
	MarketDataPath = "/market-data/by-type"
)

func (c *TastyAPI) GetMarketData(ctx context.Context, params *MarketDataParams) ([]MarketData, error) {
	res := &MarketDataResponse{}
	path := c.baseurl + MarketDataPath
	err := c.request(ctx, http.MethodGet, auth, path, params, nil, res)
	return res.Data.EquityOptions, err
}
