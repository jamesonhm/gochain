package tasty

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type EquityOptionsParams struct {
	Active      *bool    `url:"active,omitempty"`
	Symbols     []string `url:"symbol[]"`
	WithExpired *bool    `url:"with-expired,omitempty"`
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

type TickSize struct {
	Value     decimal.Decimal `json:"value"`
	Threshold decimal.Decimal `json:"threshold"`
	Symbol    string          `json:"symbol"`
}

type Deliverable struct {
	ID              int             `json:"id"`
	RootSymbol      string          `json:"root-symbol"`
	DeliverableType string          `json:"deliverable-type"`
	Description     string          `json:"description"`
	Amount          decimal.Decimal `json:"amount"`
	Symbol          string          `json:"symbol"`
	InstrumentType  InstrumentType  `json:"instrument-type"`
	Percent         decimal.Decimal `json:"percent"`
}

type Expiration struct {
	ExpirationType   string   `json:"expiration-type"`
	ExpirationDate   string   `json:"expiration-date"`
	DaysToExpiration int      `json:"days-to-expiration"`
	SettlementType   string   `json:"settlement-type"`
	Strikes          []Strike `json:"strikes"`
}

type Strike struct {
	StrikePrice        decimal.Decimal `json:"strike-price"`
	Call               string          `json:"call"`
	CallStreamerSymbol string          `json:"call-streamer-symbol"`
	Put                string          `json:"put"`
	PutStreamerSymbol  string          `json:"put-streamer-symbol"`
}

type NestedOptionChainsResponse struct {
	Data struct {
		NestedOptionChains []NestedOptionChains `json:"items"`
	} `json:"data"`
}

type NestedOptionChains struct {
	UnderlyingSymbol  string        `json:"underlying-symbol"`
	RootSymbol        string        `json:"root-symbol"`
	OptionChainType   string        `json:"option-chain-type"`
	SharesPerContract int           `json:"shares-per-contract"`
	TickSizes         []TickSize    `json:"tick-sizes"`
	Deliverables      []Deliverable `json:"deliverables"`
	Expirations       []Expiration  `json:"expirations"`
}

type CompactOptionChainsResponse struct {
	Data struct {
		CompactOptionChains []CompactOptionChains `json:"items"`
	} `json:"data"`
}

type CompactOptionChains struct {
	UnderlyingSymbol  string        `json:"underlying-symbol"`
	RootSymbol        string        `json:"root-symbol"`
	OptionChainType   string        `json:"option-chain-type"`
	SettlementType    string        `json:"settlement-type"`
	SharesPerContract int           `json:"shares-per-contract"`
	ExpirationType    string        `json:"expiration-type"`
	Deliverables      []Deliverable `json:"deliverables"`
	Symbols           []string      `json:"symbols"`
	StreamerSymbols   []string      `json:"streamer-symbols"`
}

// EquityOptionsSymbology is a struct to help build option symbol in correct OCC Symbology
// Root symbol of the underlying stock or ETF, padded with spaces to 6 characters.
// Expiration date, 6 digits in the format yymmdd. Option type, either P or C, for
// put or call.
type EquityOptionsSymbology struct {
	Symbol     string
	OptionType OptionType
	Strike     float32
	Expiration time.Time
}

// Builds the equity option into correct symbology.
func (sym EquityOptionsSymbology) Build() string {
	expiryString := sym.Expiration.Format("060102")
	strikeString := getStrikeWithPadding(sym.Strike)
	symbol := getSymbolWithPadding(sym.Symbol)
	return fmt.Sprintf("%s%s%s%s", symbol, expiryString, sym.OptionType, strikeString)
}

// convert the strike into a string with correct padding.
func getStrikeWithPadding(strike float32) string {
	strikeString := fmt.Sprintf("%d", int(strike*1000))
	for len(strikeString) < 8 {
		strikeString = "0" + strikeString
	}
	return strikeString
}

// convert the symbol into a string with correct padding.
func getSymbolWithPadding(symbol string) string {
	for len(symbol) < 6 {
		symbol += " "
	}

	return symbol
}
