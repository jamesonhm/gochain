package tasty

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jamesonhm/gochain/internal/rate"
	"github.com/jamesonhm/gochain/internal/uri"
)

const (
	API_URL     = "https://api.tastyworks.com"
	SANDBOX_URL = "https://api.cert.tastyworks.com"
)

type TastyAPI struct {
	baseurl    string
	authToken  *string
	httpClient *http.Client
	uriBuilder *uri.URIBuilder
	limiter    *rate.Limiter
}

func New(timeout time.Duration, rate_period time.Duration, rate_count int, sb bool) *TastyAPI {
	base_url := API_URL
	if sb {
		base_url = SANDBOX_URL
	}
	return &TastyAPI{
		baseurl: base_url,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		uriBuilder: uri.New(),
		limiter:    rate.New(rate_period, rate_count),
	}
}

// encodePath takes the path format string and embeds the path params and adds any query params
func (c *TastyAPI) encodePath(path string, params any) string {
	encPath := c.uriBuilder.EncodeParams(path, params)
	return encPath
}

func (c *TastyAPI) request(
	ctx context.Context,
	method string,
	auth bool,
	path string,
	params,
	payload,
	response any,
) error {
	err := c.limiter.Wait(ctx)
	if err != nil {
		return err
	}

	if auth && c.authToken == nil {
		return fmt.Errorf("invalid session")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling payload: %w", err)
	}

	encPath := c.encodePath(path, params)
	fullURL := c.baseurl + encPath
	slog.LogAttrs(ctx, slog.LevelInfo, "TastyTrade Call", slog.String("URI", fullURL))

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	if auth {
		req.Header.Add("Authorization", *c.authToken)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return fmt.Errorf("client error occurred, status code: %d, err: %w", resp.StatusCode, err)
	}
	if resp.StatusCode >= 500 {
		return fmt.Errorf("server error occurred, status code: %d, err: %w", resp.StatusCode, err)
	}

	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return fmt.Errorf("error decoding json: %w", err)
	}

	return nil
}
