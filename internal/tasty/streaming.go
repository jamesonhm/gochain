package tasty

import (
	"context"
	"net/http"
)

const (
	StreamingPath = "/api-quote-tokens"
)

func (c *TastyAPI) GetQuoteStreamerToken(ctx context.Context) (*QuoteStreamerToken, error) {
	res := &QuoteStreamerTokenResult{}
	path := c.baseurl + StreamingPath
	err := c.request(ctx, http.MethodGet, auth, path, nil, nil, res)
	return &res.QuoteStreamerToken, err
}
