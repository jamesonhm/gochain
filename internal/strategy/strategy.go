package strategy

import (
	"github.com/jamesonhm/gochain/internal/dxlink"
	"github.com/jamesonhm/gochain/internal/tasty"
)

type Strategy struct {
	Name            string
	Underlying      string
	Legs            []OptionLeg
	EntryConditions map[string]EntryCondition
	RiskParams      RiskParams
	ExitConditions  map[string]ExitCondition
}

type OptionLeg struct {
	Quantity int
	Side     string // sell or buy
	Type     string // call or put
	DTE      int
	Strike   StrikeCalc
}

type RiskParams struct {
	PctPortfolio float64
	NumContracts int
}

type EntryCondition func(marketData dxlink.DxLinkClient, accountData tasty.TastyAPI) bool
type ExitCondition func(marketData dxlink.DxLinkClient, accountData tasty.TastyAPI) bool
type StrikeCalc func(marketData dxlink.DxLinkClient) float64
