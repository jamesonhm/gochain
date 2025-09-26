package tasty

import (
	"context"
	"net/http"
	"strings"
)

const (
	DryRunOrderPath = "/accounts/{account_number}/orders/dry-run"
	OrderPath       = "/accounts/{account_number}/orders"
)

func (c *TastyAPI) SubmitOrderDryRun(ctx context.Context, acctNum string, order *NewOrder) (*SubmitOrderResponse, error) {
	res := &SubmitOrderResponse{}
	path := c.baseurl + DryRunOrderPath
	path = strings.ReplaceAll(path, "{account_number}", acctNum)
	err := c.request(ctx, http.MethodPost, auth, path, nil, order, res)
	return res, err
}

func (c *TastyAPI) SubmitOrder(ctx context.Context, acctNum string, order *NewOrder) (*SubmitOrderResponse, error) {
	res := &SubmitOrderResponse{}
	path := c.baseurl + OrderPath
	path = strings.ReplaceAll(path, "{account_number}", acctNum)
	err := c.request(ctx, http.MethodPost, auth, path, nil, order, res)
	return res, err
}
