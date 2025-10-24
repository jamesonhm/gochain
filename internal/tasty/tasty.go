package tasty

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
	API_URL      = "https://api.tastyworks.com"
	SANDBOX_URL  = "https://api.cert.tastyworks.com"
	BACKTEST_URL = "https://backtester.vast.tastyworks.com"
)

type TastyEnv string

const (
	TastyProd    TastyEnv = "PROD"
	TastySandbox TastyEnv = "SANDBOX"
)

type AuthReq bool

const (
	noAuth AuthReq = false
	auth   AuthReq = true
)

type TastyAPI struct {
	baseurl    string
	session    *Session
	httpClient *http.Client
	//uriBuilder *uri.URIBuilder
	limiter *rate.Limiter
	Env     TastyEnv
}

func New(timeout time.Duration, rate_period time.Duration, rate_count int, env TastyEnv) *TastyAPI {
	var base_url string
	switch env {
	case TastyProd:
		base_url = API_URL
	case TastySandbox:
		base_url = SANDBOX_URL
	}
	return &TastyAPI{
		baseurl: base_url,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		//uriBuilder: uri.New(),
		limiter: rate.New(rate_period, rate_count),
		Env:     env,
	}
}

// encodePath takes the path format string and embeds the path params and adds any query params
//func (c *TastyAPI) encodePath(path string, params any) string {
//	encPath := c.uriBuilder.EncodeParams(path, params)
//	return encPath
//}

func (c *TastyAPI) request(
	ctx context.Context,
	method string,
	auth AuthReq,
	path string,
	params,
	payload,
	response any,
) error {
	err := c.limiter.Wait(ctx)
	if err != nil {
		return err
	}

	if auth && c.session.Data.SessionToken == nil {
		return fmt.Errorf("invalid session")
	}

	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, path, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	if auth {
		req.Header.Add("Authorization", *c.session.Data.SessionToken)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "gochain-client/0.1")

	//fullURL := c.encodePath(path, params)
	if params != nil {
		qstring, qerr := query.Values(params)
		if qerr != nil {
			return fmt.Errorf("Query params error: %v", qerr)
		}
		req.URL.RawQuery = qstring.Encode()
	}
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		"TastyTrade Call",
		slog.String("URL", req.URL.String()),
		slog.String("method", req.Method),
	)

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

func (c *TastyAPI) GetUser() string {
	if c.session == nil {
		return "NOT LOGGED IN"
	}
	return *c.session.Data.User.Username
}

func (c *TastyAPI) GetToken() string {
	if c.session == nil {
		return "NOT LOGGED IN"
	}
	return *c.session.Data.SessionToken
}
