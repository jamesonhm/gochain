package tasty

import (
	"time"

	"github.com/shopspring/decimal"
)

type EquityOptionsParams struct {
	Active      *bool    `url:"active"`
	Symbol      []string `url:"symbol[]"`
	WithExpired *bool    `url:"with-expired"`
}

type EquityOptionSymbol struct {
	// the symbol of the equity option using OCC symbology, i.e. FB 180629C00200000
	Active *bool `url:"active"`
}

type EquityOptionsResponse struct {
	Data struct {
		EquityOptions []EquityOption `json:"items"`
	} `json:"data"`
}

type EquityOptionResponse struct {
	EquityOption EquityOption `json:"data"`
}

type EquityOption struct {
	Symbol                         string          `json:"symbol"`
	InstrumentType                 InstrumentType  `json:"instrument-type"`
	Active                         bool            `json:"active"`
	ListedMarket                   string          `json:"listed-market"`
	StrikePrice                    decimal.Decimal `json:"strike-price"`
	RootSymbol                     string          `json:"root-symbol"`
	UnderlyingSymbol               string          `json:"underlying-symbol"`
	ExpirationDate                 string          `json:"expiration-date"`
	ExerciseStyle                  string          `json:"exercise-style"`
	SharesPerContract              int             `json:"shares-per-contract"`
	OptionType                     OptionType      `json:"option-type"`
	OptionChainType                string          `json:"option-chain-type"`
	ExpirationType                 string          `json:"expiration-type"`
	SettlementType                 string          `json:"settlement-type"`
	HaltedAt                       string          `json:"halted-at"`
	StopsTradingAt                 time.Time       `json:"stops-trading-at"`
	MarketTimeInstrumentCollection string          `json:"market-time-instrument-collection"`
	DaysToExpiration               int             `json:"days-to-expiration"`
	ExpiresAt                      time.Time       `json:"expires-at"`
	IsClosingOnly                  bool            `json:"is-closing-only"`
	OldSecurityNumber              string          `json:"old-security-number"`
	StreamerSymbol                 string          `json:"streamer-symbol"`
}
