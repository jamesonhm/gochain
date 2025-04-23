package tasty

import (
	"context"
	"net/http"
)

const (
	InstrumentOptionsPath = "/instruments/equity-options"
	InstrumentOptionPath  = "/instruments/equity-options/{symbol}"
)

func (c *TastyAPI) GetEquityOptions(ctx context.Context, params *EquityOptionsParams) ([]EquityOption, error) {
	res := &EquityOptionsResponse{}
	path := c.baseurl + InstrumentOptionsPath
	err := c.request(ctx, http.MethodGet, auth, path, params, nil, res)
	return res.Data.EquityOptions, err
}

func (c *TastyAPI) GetEquityOption(ctx context.Context, params *EquityOptionSymbol) (*EquityOption, error) {
	res := &EquityOptionResponse{}
	path := c.baseurl + InstrumentOptionPath
	err := c.request(ctx, http.MethodGet, auth, path, params, nil, res)
	return &res.EquityOption, err
}
