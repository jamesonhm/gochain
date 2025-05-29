package yahoo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/go-querystring/query"
	"github.com/jamesonhm/gochain/internal/rate"
)

const (
	API_URL = "https://yahoo-finance15.p.rapidapi.com/api"
)

type YahooAPI struct {
	apikey     string
	baseurl    string
	httpClient *http.Client
	limiter    *rate.Limiter
	cache      map[string]*HistoryResponse
}

func New(key string, timeout time.Duration, rate_period time.Duration, rate_count int) *YahooAPI {
	return &YahooAPI{
		apikey:  key,
		baseurl: API_URL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		limiter: rate.New(rate_period, rate_count),
	}
}

func (c *YahooAPI) request(
	ctx context.Context,
	method string,
	path string,
	params,
	payload,
	response any,
) error {
	err := c.limiter.Wait(ctx)
	if err != nil {
		return err
	}

	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling payload: %w", err)
	}
	fmt.Println("request payload:", string(body))

	req, err := http.NewRequestWithContext(ctx, method, path, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Add("x-rapidapi-key", c.apikey)
	req.Header.Add("x-rapidapi-host", "yahoo-finance15.p.rapidapi.com")

	if params != nil {
		qstring, qerr := query.Values(params)
		if qerr != nil {
			return fmt.Errorf("Query params error: %v", qerr)
		}
		req.URL.RawQuery = qstring.Encode()
	}
	slog.LogAttrs(ctx, slog.LevelInfo, "Yahoo API Call", slog.String("URL", req.URL.String()))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return fmt.Errorf("client error occurred, status code: %d", resp.StatusCode)
	}
	if resp.StatusCode >= 500 {
		return fmt.Errorf("server error occurred, status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return fmt.Errorf("error decoding json: %w", err)
	}

	return nil
}
