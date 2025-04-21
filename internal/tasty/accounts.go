package tasty

import (
	"context"
	"net/http"
)

const (
	CustomerInfoPath         = "/customers/me"
	AccountsPath             = "/customers/me/accounts"
	AccountPath              = "/customers/me/accounts{account_number}"
	AccountTradingStatusPath = "/accounts{account_number}/trading-status"
)

func (c *TastyAPI) GetCustomer(ctx context.Context) (*CustomerResponse, error) {
	res := &CustomerResponse{}
	path := c.baseurl + CustomerInfoPath
	err := c.request(ctx, http.MethodGet, auth, path, nil, nil, res)
	return res, err
}

func (c *TastyAPI) GetAccounts(ctx context.Context) (*AccountsResponse, error) {
	res := &AccountsResponse{}
	path := c.baseurl + AccountsPath
	err := c.request(ctx, http.MethodGet, auth, path, nil, nil, res)
	return res, err
}

func (c *TastyAPI) GetAccount(ctx context.Context, params *AcctNumParams) (*AccountResponse, error) {
	res := &AccountResponse{}
	path := c.baseurl + AccountPath
	err := c.request(ctx, http.MethodGet, auth, path, params, nil, res)
	return res, err
}

func (c *TastyAPI) GetAccountTradingStatus(ctx context.Context, params *AcctNumParams) (*AccountTradingStatus, error) {
	res := &AccountTradingStatus{}
	path := c.baseurl + AccountPath
	err := c.request(ctx, http.MethodGet, auth, path, params, nil, res)
	return res, err
}
