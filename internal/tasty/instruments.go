package tasty

import (
	"context"
	"net/http"
	"strings"
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

func (c *TastyAPI) GetEquityOption(ctx context.Context, symbol EquityOptionsSymbology, active bool) (*EquityOption, error) {
	occSym := symbol.Build()
	type activeQuery struct {
		Active bool `url:"active"`
	}
	params := activeQuery{Active: active}
	res := &EquityOptionResponse{}
	path := c.baseurl + InstrumentOptionPath
	path = strings.ReplaceAll(path, "{symbol}", occSym)
	err := c.request(ctx, http.MethodGet, auth, path, params, nil, res)
	return &res.EquityOption, err
}
