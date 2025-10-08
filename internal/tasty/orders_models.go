package tasty

import (
	"time"

	"github.com/shopspring/decimal"
)

type Order struct {
	AccountNumber            string          `json:"account-number"`
	ContingentStatus         string          `json:"contingent-status"`
	ConfirmationStatus       string          `json:"confirmation-status"`
	Cancellable              bool            `json:"cancellable"`
	CancelledAt              time.Time       `json:"cancelled-at"`
	CancelUserID             string          `json:"cancel-user-id"`
	CancelUsername           string          `json:"cancel-username"`
	ComplexOrderID           int             `json:"complex-order-id"`
	ComplexOrderTag          string          `json:"complex-order-tag"`
	Editable                 bool            `json:"editable"`
	Edited                   bool            `json:"edited"`
	ExtExchangeOrderNumber   string          `json:"ext-exchange-order-number"`
	ExtClientOrderID         string          `json:"ext-client-order-id"`
	ExtGlobalOrderNumber     int             `json:"ext-global-order-number"`
	GtcDate                  string          `json:"gtc-date"`
	ID                       int             `json:"id"`
	InFlightAt               string          `json:"in-flight-at"`
	Legs                     []OrderLeg      `json:"legs"`
	LiveAt                   string          `json:"live-at"`
	OrderType                OrderType       `json:"order-type"`
	PreflightID              string          `json:"preflight-id"`
	Price                    decimal.Decimal `json:"price"`
	PriceEffect              PriceEffect     `json:"price-effect"`
	ReceivedAt               time.Time       `json:"received-at"`
	ReplacingOrderID         string          `json:"replacing-order-id"`
	ReplacesOrderID          string          `json:"replaces-order-id"`
	RejectReason             string          `json:"reject-reason"`
	Rules                    OrderRules      `json:"rules"`
	Size                     int             `json:"size"`
	Source                   string          `json:"source"`
	Status                   OrderStatus     `json:"status"`
	StopTrigger              decimal.Decimal `json:"stop-trigger"`
	TerminalAt               time.Time       `json:"terminal-at"`
	TimeInForce              TimeInForce     `json:"time-in-force"`
	UnderlyingSymbol         string          `json:"underlying-symbol"`
	UnderlyingInstrumentType InstrumentType  `json:"underlying-instrument-type"`
	UpdatedAt                int             `json:"updated-at"`
	UserID                   string          `json:"user-id"`
	Username                 string          `json:"username"`
	Value                    decimal.Decimal `json:"value"`
	ValueEffect              PriceEffect     `json:"value-effect"`
}

type OrderLeg struct {
	InstrumentType    InstrumentType `json:"instrument-type"`
	Symbol            string         `json:"symbol"`
	Quantity          float64        `json:"quantity"`
	RemainingQuantity float64        `json:"remaining-quantity"`
	Action            OrderAction    `json:"action"`
	Fills             []OrderFill    `json:"fills"`
}

type OrderFill struct {
	ExtGroupFillID   string          `json:"ext-group-fill-id"`
	ExtExecID        string          `json:"ext-exec-id"`
	FillID           string          `json:"fill-id"`
	Quantity         float64         `json:"quantity"`
	FillPrice        decimal.Decimal `json:"fill-price"`
	FilledAt         time.Time       `json:"filled-at"`
	DestinationVenue string          `json:"destination-venue"`
}

type OrderRules struct {
	RouteAfter  string           `json:"route-after"`
	RoutedAt    string           `json:"routed-at"`
	CancelAt    string           `json:"cancel-at"`
	CancelledAt string           `json:"cancelled-at"`
	Conditions  []OrderCondition `json:"conditions"`
}

type OrderCondition struct {
	ID                         int                   `json:"id"`
	Action                     OrderRuleAction       `json:"action"`
	Symbol                     string                `json:"symbol"`
	InstrumentType             InstrumentType        `json:"instrument-type"`
	Indicator                  Indicator             `json:"indicator"`
	Comparator                 Comparator            `json:"comparator"`
	Threshold                  decimal.Decimal       `json:"threshold"`
	IsThresholdBasedOnNotional bool                  `json:"is-threshold-based-on-notional"`
	TriggeredAt                string                `json:"triggered-at"`
	TriggeredValue             decimal.Decimal       `json:"triggered-value"`
	PriceComponents            []OrderPriceComponent `json:"price-components"`
}

type OrderPriceComponent struct {
	Symbol            string         `json:"symbol"`
	InstrumentType    InstrumentType `json:"instrument-type"`
	Quantity          float64        `json:"quantity"`
	QuantityDirection Direction      `json:"quantity-direction"`
}

type NewOrder struct {
	TimeInForce  TimeInForce   `json:"time-in-force"`
	GtcDate      string        `json:"gtc-date"`
	OrderType    OrderType     `json:"order-type"`
	StopTrigger  float64       `json:"stop-trigger,omitempty"`
	Price        string        `json:"price,omitempty"`
	PriceEffect  PriceEffect   `json:"price-effect,omitempty"`
	Value        float64       `json:"value,omitempty"`
	ValueEffect  PriceEffect   `json:"value-effect,omitempty"`
	Source       string        `json:"source,omitempty"`
	PartitionKey string        `json:"partition-key,omitempty"`
	PreflightID  string        `json:"preflight-id,omitempty"`
	Legs         []NewOrderLeg `json:"legs"`
	Rules        NewOrderRules `json:"rules,omitempty"`
}

type NewOrderLeg struct {
	InstrumentType InstrumentType `json:"instrument-type"`
	Symbol         string         `json:"symbol"`
	Quantity       float64        `json:"quantity,omitempty"`
	Action         OrderAction    `json:"action"`
}

type NewOrderRules struct {
	// RouteAfter Earliest time an order should route at
	RouteAfter string `json:"route-after"`
	// CancelAt Latest time an order should be canceled at
	CancelAt   string              `json:"cancel-at"`
	Conditions []NewOrderCondition `json:"conditions"`
}

type NewOrderCondition struct {
	// The action in which the trigger is enacted. i.e. route and cancel
	Action OrderRuleAction `json:"action"`
	// The symbol to apply the condition to.
	Symbol string `json:"symbol"`
	// The instrument's type in relation to the condition.
	InstrumentType string `json:"instrument-type"`
	// The indicator for the trigger, currently only supports last
	Indicator Indicator `json:"indicator"`
	// How to compare against the threshold.
	Comparator Comparator `json:"comparator"`
	// The price at which the condition triggers.
	Threshold       float32                  `json:"threshold"`
	PriceComponents []NewOrderPriceComponent `json:"price-components"`
}

type NewOrderPriceComponent struct {
	// The symbol to apply the condition to.
	Symbol string `json:"symbol"`
	// The instrument's type in relation to the symbol.
	InstrumentType InstrumentType `json:"instrument-type"`
	// The Ratio quantity in relation to the symbol
	Quantity float32 `json:"quantity"`
	// The quantity direction(ie Long or Short) in relation to the symbol
	QuantityDirection Direction `json:"quantity-direction"`
}

type SubmitOrderResponse struct {
	OrderResponse OrderResponse       `json:"data"`
	OrderError    *OrderErrorResponse `json:"error"`
}

type OrderResponse struct {
	Order             Order             `json:"order"`
	ComplexOrder      ComplexOrder      `json:"complex-order"`
	Warnings          []OrderInfo       `json:"warnings"`
	Errors            []OrderInfo       `json:"errors"`
	BuyingPowerEffect BuyingPowerEffect `json:"buying-power-effect"`
	FeeCalculation    FeeCalculation    `json:"fee-calculation"`
}

type OrderErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Errors  []OrderInfo `json:"errors"`
}

type ComplexOrder struct {
	ID                                   int             `json:"id"`
	AccountNumber                        string          `json:"account-number"`
	Type                                 string          `json:"type"`
	TerminalAt                           string          `json:"terminal-at"`
	RatioPriceThreshold                  decimal.Decimal `json:"ratio-price-threshold"`
	RatioPriceComparator                 string          `json:"ratio-price-comparator"`
	RatioPriceIsThresholdBasedOnNotional bool            `json:"ratio-price-is-threshold-based-on-notional"`
	// RelatedOrders Non-current orders. This includes replaced orders, unfilled orders, and terminal orders.
	RelatedOrders []RelatedOrder `json:"related-orders"`
	// Orders with complex-order-tag: '::order'. For example, 'OTO::order' for OTO complex orders.
	Orders []Order `json:"orders"`
	// TriggerOrder Order with complex-order-tag: '::trigger-order'. For example, 'OTO::trigger-order for OTO complex orders.
	TriggerOrder Order `json:"trigger-order"`
}

type OrderInfo struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	PreflightID string `json:"preflight-id"`
}

type BuyingPowerEffect struct {
	ChangeInMarginRequirement            decimal.Decimal `json:"change-in-margin-requirement"`
	ChangeInMarginRequirementEffect      PriceEffect     `json:"change-in-margin-requirement-effect"`
	ChangeInBuyingPower                  decimal.Decimal `json:"change-in-buying-power"`
	ChangeInBuyingPowerEffect            PriceEffect     `json:"change-in-buying-power-effect"`
	CurrentBuyingPower                   decimal.Decimal `json:"current-buying-power"`
	CurrentBuyingPowerEffect             PriceEffect     `json:"current-buying-power-effect"`
	NewBuyingPower                       decimal.Decimal `json:"new-buying-power"`
	NewBuyingPowerEffect                 PriceEffect     `json:"new-buying-power-effect"`
	IsolatedOrderMarginRequirement       decimal.Decimal `json:"isolated-order-margin-requirement"`
	IsolatedOrderMarginRequirementEffect PriceEffect     `json:"isolated-order-margin-requirement-effect"`
	IsSpread                             bool            `json:"is-spread"`
	Impact                               decimal.Decimal `json:"impact"`
	Effect                               PriceEffect     `json:"effect"`
}

type FeeCalculation struct {
	RegulatoryFees                   decimal.Decimal `json:"regulatory-fees"`
	RegulatoryFeesEffect             PriceEffect     `json:"regulatory-fees-effect"`
	ClearingFees                     decimal.Decimal `json:"clearing-fees"`
	ClearingFeesEffect               PriceEffect     `json:"clearing-fees-effect"`
	Commission                       decimal.Decimal `json:"commission"`
	CommissionEffect                 PriceEffect     `json:"commission-effect"`
	ProprietaryIndexOptionFees       decimal.Decimal `json:"proprietary-index-option-fees"`
	ProprietaryIndexOptionFeesEffect PriceEffect     `json:"proprietary-index-option-fees-effect"`
	TotalFees                        decimal.Decimal `json:"total-fees"`
	TotalFeesEffect                  PriceEffect     `json:"total-fees-effect"`
}

type RelatedOrder struct {
	ID               int    `json:"id"`
	ComplexOrderID   int    `json:"complex-order-id"`
	ComplexOrderTag  string `json:"complex-order-tag"`
	ReplacesOrderID  string `json:"replaces-order-id"`
	ReplacingOrderID string `json:"replacing-order-id"`
	Status           string `json:"status"`
}
