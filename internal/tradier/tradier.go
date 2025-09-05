package tradier

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	//"github.com/google/go-querystring/query"
	"github.com/jamesonhm/gochain/internal/rate"
)

const (
	API_URL       = "https://api.tradier.com"
	SANDBOX_URL   = "https://sandbox.tradier.com"
	STREAMING_URL = "https://stream.tradier.com"
)

type Client struct {
	baseurl string
	apiKey  string
	httpC   http.Client
	//uriBuilder *uri.URIBuilder
	limiter *rate.Limiter
}

func New(apiKey string, timeout time.Duration, rate_period time.Duration, rate_count int) Client {
	return Client{
		baseurl: API_URL,
		apiKey:  apiKey,
		httpC: http.Client{
			Timeout: timeout,
		},
		//uriBuilder: uri.New(),
		limiter: rate.New(rate_period, rate_count),
	}
}

// Call makes API call based on path and params
//func (c *Client) Call(ctx context.Context, path string, params, response any) error {
//uri := c.uriBuilder.EncodeParams(path, params)
//if err != nil {
//	return err
//}
//slog.LogAttrs(ctx, slog.LevelInfo, "Tradier Call", slog.String("URI", uri))
//return c.CallURL(ctx, uri, response)
//}

func (c *Client) CallURL(ctx context.Context, uri string, response any) error {
	err := c.limiter.Wait(ctx)
	if err != nil {
		return err
	}
	uri = c.baseurl + uri
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+c.apiKey)
	resp, err := c.httpC.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return fmt.Errorf("error decoding json: %w", err)
	}

	return nil
}
