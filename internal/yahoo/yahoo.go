package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	cache      *Cache
}

func New(
	key string,
	timeout time.Duration,
	rate_period time.Duration,
	rate_count int,
	cache_retention time.Duration,
) *YahooAPI {
	return &YahooAPI{
		apikey:  key,
		baseurl: API_URL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		limiter: rate.New(rate_period, rate_count),
		cache:   NewCache(cache_retention),
	}
}

func (c *YahooAPI) cachedRequest(
	ctx context.Context,
	path string,
	params,
	response any,
) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}

	if params != nil {
		qstring, qerr := query.Values(params)
		if qerr != nil {
			return fmt.Errorf("Query params error: %v", qerr)
		}
		req.URL.RawQuery = qstring.Encode()
	}

	if data, ok := c.cache.Get(req.URL.String()); ok {
		slog.LogAttrs(ctx, slog.LevelInfo, "Yahoo Cache Get", slog.String("URL", req.URL.String()))
		if err := json.Unmarshal(data, &response); err != nil {
			return fmt.Errorf("Cache Unmarshal Error: %w", err)
		}
		return nil
	}

	req.Header.Add("x-rapidapi-key", c.apikey)
	req.Header.Add("x-rapidapi-host", "yahoo-finance15.p.rapidapi.com")

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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error Reading resp.Body: %w", err)
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return fmt.Errorf("Body Unmarshal Error: %w", err)
	}

	c.cache.Add(req.URL.String(), body)
	return nil
}

func (c *YahooAPI) request(
	ctx context.Context,
	path string,
	params,
	response any,
) error {
	err := c.limiter.Wait(ctx)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
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
