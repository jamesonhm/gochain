package tasty

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

type Address struct {
	StreetOne   string `json:"street-one"`
	City        string `json:"city"`
	StateRegion string `json:"state-region"`
	PostalCode  string `json:"postal-code"`
	Country     string `json:"country"`
	IsForeign   bool   `json:"is-foreign"`
	IsDomestic  bool   `json:"is-domestic"`
}

type CustomerSuitability struct {
	ID                                int    `json:"id"`
	MaritalStatus                     string `json:"marital-status"`
	NumberOfDependents                int    `json:"number-of-dependents"`
	EmploymentStatus                  string `json:"employment-status"`
	Occupation                        string `json:"occupation"`
	EmployerName                      string `json:"employer-name"`
	JobTitle                          string `json:"job-title"`
	AnnualNetIncome                   int    `json:"annual-net-income"`
	NetWorth                          int    `json:"net-worth"`
	LiquidNetWorth                    int    `json:"liquid-net-worth"`
	StockTradingExperience            string `json:"stock-trading-experience"`
	CoveredOptionsTradingExperience   string `json:"covered-options-trading-experience"`
	UncoveredOptionsTradingExperience string `json:"uncovered-options-trading-experience"`
	FuturesTradingExperience          string `json:"futures-trading-experience"`
}

type Person struct {
	ExternalID         string `json:"external-id"`
	FirstName          string `json:"first-name"`
	LastName           string `json:"last-name"`
	BirthDate          string `json:"birth-date"`
	CitizenshipCountry string `json:"citizenship-country"`
	USACitizenshipType string `json:"usa-citizenship-type"`
	MaritalStatus      string `json:"marital-status"`
	NumberOfDependents int    `json:"number-of-dependents"`
	EmploymentStatus   string `json:"employment-status"`
	Occupation         string `json:"occupation"`
	EmployerName       string `json:"employer-name"`
	JobTitle           string `json:"job-title"`
}

type CustomerData struct {
	ID                              string              `json:"id"`
	FirstName                       string              `json:"first-name"`
	LastName                        string              `json:"last-name"`
	Address                         Address             `json:"address"`
	MailingAddress                  Address             `json:"mailing-address"`
	CustomerSuitability             CustomerSuitability `json:"customer-suitability"`
	USACitizenshipType              string              `json:"usa-citizenship-type"`
	IsForeign                       bool                `json:"is-foreign"`
	MobilePhoneNumber               string              `json:"mobile-phone-number"`
	Email                           string              `json:"email"`
	TaxNumberType                   string              `json:"tax-number-type"`
	TaxNumber                       string              `json:"tax-number"`
	BirthDate                       string              `json:"birth-date"`
	ExternalID                      string              `json:"external-id"`
	CitizenshipCountry              string              `json:"citizenship-country"`
	SubjectToTaxWithholding         bool                `json:"subject-to-tax-withholding"`
	AgreedToMargining               bool                `json:"agreed-to-margining"`
	AgreedToTerms                   bool                `json:"agreed-to-terms"`
	HasIndustryAffiliation          bool                `json:"has-industry-affiliation"`
	HasPoliticalAffiliation         bool                `json:"has-political-affiliation"`
	HasListedAffiliation            bool                `json:"has-listed-affiliation"`
	IsProfessional                  bool                `json:"is-professional"`
	HasDelayedQuotes                bool                `json:"has-delayed-quotes"`
	HasPendingOrApprovedApplication bool                `json:"has-pending-or-approved-application"`
	IdentifiableType                string              `json:"identifiable-type"`
	Person                          Person              `json:"person"`
}

type CustomerResponse struct {
	Context string       `json:"context"`
	Data    CustomerData `json:"data"`
}

type Account struct {
	AccountNumber         string `json:"account-number"`
	ExternalID            string `json:"external-id"`
	OpenedAt              string `json:"opened-at"`
	Nickname              string `json:"nickname"`
	AccountTypeName       string `json:"account-type-name"`
	DayTraderStatus       bool   `json:"day-trader-status"`
	IsClosed              bool   `json:"is-closed"`
	IsFirmError           bool   `json:"is-firm-error"`
	IsFirmProprietary     bool   `json:"is-firm-proprietary"`
	IsFuturesApproved     bool   `json:"is-futures-approved"`
	IsTestDrive           bool   `json:"is-test-drive"`
	MarginOrCash          string `json:"margin-or-cash"`
	IsForeign             bool   `json:"is-foreign"`
	FundingDate           string `json:"funding-date"`
	InvestmentObjective   string `json:"investment-objective"`
	FuturesAccountPurpose string `json:"futures-account-purpose"`
	SuitableOptionsLevel  string `json:"suitable-options-level"`
	CreatedAt             string `json:"created-at"`
}

func (a *Account) String() string {
	j, _ := json.MarshalIndent(a, "", "  ")
	return string(j)
}

type AccountContainer struct {
	Account        Account `json:"account"`
	AuthorityLevel string  `json:"authority-level"`
}

type AccountData struct {
	Items []AccountContainer `json:"items"`
}

type AccountResponse struct {
	Context string  `json:"context"`
	Data    Account `json:"data"`
}

type AccountsResponse struct {
	Context string      `json:"context"`
	Data    AccountData `json:"data"`
}

type AcctNumParams struct {
	AcctNum string `path:"account_number"`
}

type AccountTradingStatusResponse struct {
	Data    AccountTradingStatus `json:"data"`
	Context string               `json:"context"`
}

type AccountTradingStatus struct {
	AccountNumber                            string          `json:"account-number"`
	AutotradeAccountType                     string          `json:"autotrade-account-type"`
	ClearingAccountNumber                    string          `json:"clearing-account-number"`
	ClearingAggregationIdentifier            string          `json:"clearing-aggregation-identifier"`
	DayTradeCount                            int             `json:"day-trade-count"`
	EquitiesMarginCalculationType            string          `json:"equities-margin-calculation-type"`
	FeeScheduleName                          string          `json:"fee-schedule-name"`
	FuturesMarginRateMultiplier              decimal.Decimal `json:"futures-margin-rate-multiplier"`
	HasIntradayEquitiesMargin                bool            `json:"has-intraday-equities-margin"`
	ID                                       int             `json:"id"`
	IsAggregatedAtClearing                   bool            `json:"is-aggregated-at-clearing"`
	IsClosed                                 bool            `json:"is-closed"`
	IsClosingOnly                            bool            `json:"is-closing-only"`
	IsCryptocurrencyClosingOnly              bool            `json:"is-cryptocurrency-closing-only"`
	IsCryptocurrencyEnabled                  bool            `json:"is-cryptocurrency-enabled"`
	IsFrozen                                 bool            `json:"is-frozen"`
	IsFullEquityMarginRequired               bool            `json:"is-full-equity-margin-required"`
	IsFuturesClosingOnly                     bool            `json:"is-futures-closing-only"`
	IsFuturesIntraDayEnabled                 bool            `json:"is-futures-intra-day-enabled"`
	IsFuturesEnabled                         bool            `json:"is-futures-enabled"`
	IsInDayTradeEquityMaintenanceCall        bool            `json:"is-in-day-trade-equity-maintenance-call"`
	IsInMarginCall                           bool            `json:"is-in-margin-call"`
	IsPatternDayTrader                       bool            `json:"is-pattern-day-trader"`
	IsPortfolioMarginEnabled                 bool            `json:"is-portfolio-margin-enabled"`
	IsRiskReducingOnly                       bool            `json:"is-risk-reducing-only"`
	IsSmallNotionalFuturesIntraDayEnabled    bool            `json:"is-small-notional-futures-intra-day-enabled"`
	IsRollTheDayForwardEnabled               bool            `json:"is-roll-the-day-forward-enabled"`
	AreFarOtmNetOptionsRestricted            bool            `json:"are-far-otm-net-options-restricted"`
	OptionsLevel                             string          `json:"options-level"`
	PdtResetOn                               string          `json:"pdt-reset-on"`
	ShortCallsEnabled                        bool            `json:"short-calls-enabled"`
	SmallNotionalFuturesMarginRateMultiplier decimal.Decimal `json:"small-notional-futures-margin-rate-multiplier"`
	CMTAOverride                             int             `json:"cmta-override"`
	IsEquityOfferingEnabled                  bool            `json:"is-equity-offering-enabled"`
	IsEquityOfferingClosingOnly              bool            `json:"is-equity-offering-closing-only"`
	EnhancedFraudSafeguardsEnabledAt         time.Time       `json:"enhanced-fraud-safeguards-enabled-at"`
	UpdatedAt                                time.Time       `json:"updated-at"`
}

type AccountPositionParams struct {
	AccountNumber string `path:"account_number"`
	// An array of underlying symbol(s) for positions (e.g. underlying-symbol[]={value1}&underlying-symbol[]={value2})
	UnderlyingSymbol *[]string `query:"underlying-symbol"`
	// A single symbol, stock ticker symbol (AAPL), OCC Option Symbon (AAPL 191004P00275000), TW Future symbol, TW Future Option Symbol
	Symbol *string `query:"symbol"`
	// The type of instrument
	InstrumentType         *InstrumentType `query:"instrument-type"`
	IncludeClosedPositions *bool           `query:"include-closed-positions"`
	UnderlyingProductCode  *string         `query:"underlying-product-code"`
	PartitionKeys          *[]string       `query:"partition-keys"`
	// Returns net positions grouped by instrument type and symbol
	NetPositions *bool `query:"net-positions"`
	// Include current quote mark (note: can decrease performance)
	IncludeMarks *bool `query:"include-marks"`
}

type AccountPositionResponse struct {
	Data struct {
		AccountPositions []AccountPosition `json:"items"`
	} `json:"data"`
}

type AccountPosition struct {
	AccountNumber                 string          `json:"account-number"`
	Symbol                        string          `json:"symbol"`
	InstrumentType                InstrumentType  `json:"instrument-type"`
	UnderlyingSymbol              string          `json:"underlying-symbol"`
	Quantity                      int             `json:"quantity"`
	QuantityDirection             Direction       `json:"quantity-direction"`
	ClosePrice                    decimal.Decimal `json:"close-price"`
	AverageOpenPrice              decimal.Decimal `json:"average-open-price"`
	AverageYearlyMarketClosePrice decimal.Decimal `json:"average-yearly-market-close-price"`
	AverageDailyMarketClosePrice  decimal.Decimal `json:"average-daily-market-close-price"`
	Mark                          decimal.Decimal `json:"mark"`
	MarkPrice                     decimal.Decimal `json:"mark-price"`
	Multiplier                    int             `json:"multiplier"`
	CostEffect                    PriceEffect     `json:"cost-effect"`
	IsSuppressed                  bool            `json:"is-suppressed"`
	IsFrozen                      bool            `json:"is-frozen"`
	RestrictedQuantity            int             `json:"restricted-quantity"`
	ExpiresAt                     time.Time       `json:"expires-at"`
	FixingPrice                   decimal.Decimal `json:"fixing-price"`
	DeliverableType               string          `json:"deliverable-type"`
	RealizedDayGain               decimal.Decimal `json:"realized-day-gain"`
	RealizedDayGainEffect         PriceEffect     `json:"realized-day-gain-effect"`
	RealizedDayGainDate           string          `json:"realized-day-gain-date"`
	RealizedToday                 decimal.Decimal `json:"realized-today"`
	RealizedTodayEffect           PriceEffect     `json:"realized-today-effect"`
	RealizedTodayDate             string          `json:"realized-today-date"`
	CreatedAt                     time.Time       `json:"created-at"`
	UpdatedAt                     time.Time       `json:"updated-at"`
}

func (ap *AccountPosition) String() string {
	j, _ := json.MarshalIndent(ap, "", "  ")
	return string(j)
}

type AccountBalanceResponse struct {
	AccountBalances AccountBalances `json:"data"`
}

type AccountBalances struct {
	AccountNumber                      string          `json:"account-number"`
	CashBalance                        decimal.Decimal `json:"cash-balance"`
	LongEquityValue                    decimal.Decimal `json:"long-equity-value"`
	ShortEquityValue                   decimal.Decimal `json:"short-equity-value"`
	LongDerivativeValue                decimal.Decimal `json:"long-derivative-value"`
	ShortDerivativeValue               decimal.Decimal `json:"short-derivative-value"`
	LongFuturesValue                   decimal.Decimal `json:"long-futures-value"`
	ShortFuturesValue                  decimal.Decimal `json:"short-futures-value"`
	LongFuturesDerivativeValue         decimal.Decimal `json:"long-futures-derivative-value"`
	ShortFuturesDerivativeValue        decimal.Decimal `json:"short-futures-derivative-value"`
	LongMargineableValue               decimal.Decimal `json:"long-margineable-value"`
	ShortMargineableValue              decimal.Decimal `json:"short-margineable-value"`
	MarginEquity                       decimal.Decimal `json:"margin-equity"`
	EquityBuyingPower                  decimal.Decimal `json:"equity-buying-power"`
	DerivativeBuyingPower              decimal.Decimal `json:"derivative-buying-power"`
	DayTradingBuyingPower              decimal.Decimal `json:"day-trading-buying-power"`
	FuturesMarginRequirement           decimal.Decimal `json:"futures-margin-requirement"`
	AvailableTradingFunds              decimal.Decimal `json:"available-trading-funds"`
	MaintenanceRequirement             decimal.Decimal `json:"maintenance-requirement"`
	MaintenanceCallValue               decimal.Decimal `json:"maintenance-call-value"`
	RegTCallValue                      decimal.Decimal `json:"reg-t-call-value"`
	DayTradingCallValue                decimal.Decimal `json:"day-trading-call-value"`
	DayEquityCallValue                 decimal.Decimal `json:"day-equity-call-value"`
	NetLiquidatingValue                decimal.Decimal `json:"net-liquidating-value"`
	CashAvailableToWithdraw            decimal.Decimal `json:"cash-available-to-withdraw"`
	DayTradeExcess                     decimal.Decimal `json:"day-trade-excess"`
	PendingCash                        decimal.Decimal `json:"pending-cash"`
	PendingCashEffect                  PriceEffect     `json:"pending-cash-effect"`
	LongCryptocurrencyValue            decimal.Decimal `json:"long-cryptocurrency-value"`
	ShortCryptocurrencyValue           decimal.Decimal `json:"short-cryptocurrency-value"`
	CryptocurrencyMarginRequirement    decimal.Decimal `json:"cryptocurrency-margin-requirement"`
	UnsettledCryptocurrencyFiatAmount  decimal.Decimal `json:"unsettled-cryptocurrency-fiat-amount"`
	UnsettledCryptocurrencyFiatEffect  PriceEffect     `json:"unsettled-cryptocurrency-fiat-effect"`
	ClosedLoopAvailableBalance         decimal.Decimal `json:"closed-loop-available-balance"`
	EquityOfferingMarginRequirement    decimal.Decimal `json:"equity-offering-margin-requirement"`
	LongBondValue                      decimal.Decimal `json:"long-bond-value"`
	BondMarginRequirement              decimal.Decimal `json:"bond-margin-requirement"`
	SnapshotDate                       string          `json:"snapshot-date"`
	TimeOfDay                          string          `json:"time-of-day"`
	RegTMarginRequirement              decimal.Decimal `json:"reg-t-margin-requirement"`
	FuturesOvernightMarginRequirement  decimal.Decimal `json:"futures-overnight-margin-requirement"`
	FuturesIntradayMarginRequirement   decimal.Decimal `json:"futures-intraday-margin-requirement"`
	MaintenanceExcess                  decimal.Decimal `json:"maintenance-excess"`
	PendingMarginInterest              decimal.Decimal `json:"pending-margin-interest"`
	ApexStartingDayMarginEquity        decimal.Decimal `json:"apex-starting-day-margin-equity"`
	BuyingPowerAdjustment              decimal.Decimal `json:"buying-power-adjustment"`
	BuyingPowerAdjustmentEffect        PriceEffect     `json:"buying-power-adjustment-effect"`
	EffectiveCryptocurrencyBuyingPower decimal.Decimal `json:"effective-cryptocurrency-buying-power"`
	UpdatedAt                          time.Time       `json:"updated-at"`
}
