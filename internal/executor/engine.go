package executor

import (
	"context"
	"sync"

	"github.com/jamesonhm/gochain/internal/dxlink"
	"github.com/jamesonhm/gochain/internal/strategy"
	"github.com/jamesonhm/gochain/internal/tasty"
)

type Engine struct {
	apiClient      *tasty.TastyAPI
	optionProvider *dxlink.DxLinkClient
	orderQueue     chan tasty.NewOrder
	wg             sync.WaitGroup
	workerCount    int
	ctx            context.Context
}

func NewEngine(
	apiClient *tasty.TastyAPI,
	optionProvider *dxlink.DxLinkClient,
	workerCount int,
	ctx context.Context,
) *Engine {
	e := &Engine{
		apiClient:      apiClient,
		optionProvider: optionProvider,
		orderQueue:     make(chan tasty.NewOrder, 10),
		workerCount:    workerCount,
		ctx:            ctx,
	}

	e.startWorkers()
	return e
}

// Submit order converts a strategy to an order and queues it
// called by monitor engine when conditions are met
// TODO: submit to open vs submit to close...
func (e *Engine) SubmitOrder(s strategy.Strategy) {
	order := e.orderFromStrategy(s)
	e.orderQueue <- order
}

func (e *Engine) orderFromStrategy(s strategy.Strategy) tasty.NewOrder {
	// for each leg, calculate strike price
	// create leg(s)
	// create order struct

	//type NewOrder struct {
	//	TimeInForce  TimeInForce   `json:"time-in-force"`
	//	GtcDate      string        `json:"gtc-date"`
	//	OrderType    OrderType     `json:"order-type"`
	//	StopTrigger  float32       `json:"stop-trigger,omitempty"`
	//	Price        float32       `json:"price,omitempty"`
	//	PriceEffect  PriceEffect   `json:"price-effect,omitempty"`
	//	Value        float32       `json:"value,omitempty"`
	//	ValueEffect  PriceEffect   `json:"value-effect,omitempty"`
	//	Source       string        `json:"source,omitempty"`
	//	PartitionKey string        `json:"partition-key,omitempty"`
	//	PreflightID  string        `json:"preflight-id,omitempty"`
	//	Legs         []NewOrderLeg `json:"legs"`
	//	Rules        NewOrderRules `json:"rules,omitempty"`
	//}

	//type NewOrderLeg struct {
	//	InstrumentType InstrumentType `json:"instrument-type"`
	//	Symbol         string         `json:"symbol"`
	//	Quantity       float32        `json:"quantity,omitempty"`
	//	Action         OrderAction    `json:"action"` (STO, BTO, STC, BTC, ...)
	//}
	return tasty.NewOrder{
		TimeInForce: "Day",
		OrderType:   "Limit",
		// price = sum(legs_midpoint) +/- slippage
		// price-effect = Credit if price > 0
		// legs:
	}
}

func (e *Engine) startWorkers() {
	for i := 0; i < e.workerCount; i++ {
		e.wg.Add(1)
		go e.worker()
	}
}

func (e *Engine) worker() {
	defer e.wg.Done()

	for order := range e.orderQueue {
		resp, err := e.apiClient.SubmitOrderDryRun(e.ctx, acctNum, &order)
	}
}
