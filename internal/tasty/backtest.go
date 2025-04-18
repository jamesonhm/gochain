package tasty

import (
	"context"
	"fmt"
	"net/http"
)

const (
	backtestURL         = "https://backtester.vast.tastyworks.com"
	backtestSessionPath = "/sessions"
)

type BacktestRequest struct {
	TastyToken string `json:"tastytradeToken"`
}

type BacktestSessionResponse struct {
	Token *string `json:"token"`
}

func (c *TastyAPI) BacktestSession(ctx context.Context) (*BacktestSessionResponse, error) {
	backtestURL := fmt.Sprintf("%s%s", backtestURL, backtestSessionPath)
	authData := BacktestRequest{
		TastyToken: *c.session.Data.SessionToken,
	}

	btSession := &BacktestSessionResponse{}
	err := c.request(ctx, http.MethodPost, auth, backtestURL, nil, authData, btSession)
	if err != nil {
		return nil, err
	}
	return btSession, nil
}

func (c *TastyAPI) CancelBacktestSession(ctx context.Context) error {
	backtestURL := fmt.Sprintf("%s%s", backtestURL, backtestSessionPath)

	err := c.request(ctx, http.MethodDelete, noAuth, backtestURL, nil, nil, nil)
	if err != nil {
		return err
	}
	fmt.Println("Backtest Cancelled")
	return nil
}
