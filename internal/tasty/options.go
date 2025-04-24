package tasty

import (
	"context"
	"net/http"
	"strings"
)

const (
	OptionChainPath   = "/option-chains/{symbol}"
	NestedOptionPath  = "/option-chains/{symbol}/nested"
	CompactOptionPath = "/option-chains/{symbol}/compact"
)

func (c *TastyAPI) GetOptionChain(ctx context.Context, symbol string) (*EquityOption, error) {
	res := &EquityOptionResponse{}
	path := c.baseurl + OptionChainPath
	path = strings.ReplaceAll(path, "{symbol}", symbol)
	err := c.request(ctx, http.MethodGet, auth, path, nil, nil, res)
	return &res.EquityOption, err
}

func (c *TastyAPI) GetOptionNested(ctx context.Context, symbol string) ([]NestedOptionChains, error) {
	res := &NestedOptionChainsResponse{}
	path := c.baseurl + NestedOptionPath
	path = strings.ReplaceAll(path, "{symbol}", symbol)
	err := c.request(ctx, http.MethodGet, auth, path, nil, nil, res)
	return res.Data.NestedOptionChains, err
}

func (c *TastyAPI) GetOptionCompact(ctx context.Context, symbol string) ([]CompactOptionChains, error) {
	res := &CompactOptionChainsResponse{}
	path := c.baseurl + CompactOptionPath
	path = strings.ReplaceAll(path, "{symbol}", symbol)
	err := c.request(ctx, http.MethodGet, auth, path, nil, nil, res)
	return res.Data.CompactOptionChains, err
}
