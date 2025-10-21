package tasty

import (
	"context"
	"net/http"
	"strings"
)

const (
	CustomerInfoPath         = "/customers/me"
	AccountsPath             = "/customers/me/accounts"
	AccountPath              = "/customers/me/accounts/{account_number}"
	AccountTradingStatusPath = "/accounts{account_number}/trading-status"
	AccountPositionsPath     = "/accounts/{account_number}/positions"
	AccountBalancesPath      = "/accounts/{account_number}/balances/{currency}"
	DoNotExercisesPath       = "/accounts/{account_number}/do-not-exercises"
)

func (c *TastyAPI) GetCustomer(ctx context.Context) (*CustomerResponse, error) {
	res := &CustomerResponse{}
	path := c.baseurl + CustomerInfoPath
	err := c.request(ctx, http.MethodGet, auth, path, nil, nil, res)
	return res, err
}

func (c *TastyAPI) GetAccounts(ctx context.Context) ([]AccountContainer, error) {
	res := &AccountsResponse{}
	path := c.baseurl + AccountsPath
	err := c.request(ctx, http.MethodGet, auth, path, nil, nil, res)
	return res.Data.Items, err
}

func (c *TastyAPI) GetAccount(ctx context.Context, acctNum string) (*AccountResponse, error) {
	res := &AccountResponse{}
	path := c.baseurl + AccountPath
	path = strings.ReplaceAll(path, "{account_number}", acctNum)
	err := c.request(ctx, http.MethodGet, auth, path, nil, nil, res)
	return res, err
}

func (c *TastyAPI) GetAccountTradingStatus(ctx context.Context, acctNum string) (*AccountTradingStatus, error) {
	res := &AccountTradingStatusResponse{}
	path := c.baseurl + AccountPath
	path = strings.ReplaceAll(path, "{account_number}", acctNum)
	err := c.request(ctx, http.MethodGet, auth, path, nil, nil, res)
	return &res.Data, err
}

func (c *TastyAPI) GetAccountPositions(ctx context.Context, acctNum string, params *AccountPositionParams) ([]AccountPosition, error) {
	res := &AccountPositionResponse{}
	path := c.baseurl + AccountPositionsPath
	path = strings.ReplaceAll(path, "{account_number}", acctNum)
	err := c.request(ctx, http.MethodGet, auth, path, params, nil, res)
	return res.Data.AccountPositions, err
}

func (c *TastyAPI) GetAccountBalances(ctx context.Context, acctNum string) (*AccountBalances, error) {
	res := &AccountBalanceResponse{}
	path := c.baseurl + AccountBalancesPath
	path = strings.ReplaceAll(path, "{account_number}", acctNum)
	path = strings.ReplaceAll(path, "{currency}", "USD")
	err := c.request(ctx, http.MethodGet, auth, path, nil, nil, res)
	return &res.AccountBalances, err
}

func (c *TastyAPI) PostDoNotExercise(ctx context.Context, acctNum string, payload *DoNotExerciseBody) (*DoNotExerciseResponse, error) {
	res := &DoNotExerciseResponse{}
	path := c.baseurl + DoNotExercisesPath
	path = strings.ReplaceAll(path, "{account_number}", acctNum)
	err := c.request(ctx, http.MethodPost, auth, path, nil, payload, res)
	return res, err
}
