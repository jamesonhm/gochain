package executor

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/jamesonhm/gochain/internal/dxlink"
	"github.com/jamesonhm/gochain/internal/options"
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
	fmt.Printf("strat to submit: %+v\n", s)
	order, err := e.orderFromStrategy(s)
	if err != nil {
		slog.Error("Unable to create order from strategy:", "error", err)
		return
	}
	fmt.Printf("This is where the order goes into the queue: %+v\n", order)
	// e.orderQueue <- order
}

func (e *Engine) orderFromStrategy(s strategy.Strategy) (tasty.NewOrder, error) {
	// for each leg, calculate strike price
	// create leg(s)
	// create order struct
	var price float64
	var effect tasty.PriceEffect

	orderLegs := make([]tasty.NewOrderLeg, 0)
	for i, leg := range s.Legs {
		var action tasty.OrderAction
		var midPrice float64
		var optSymbol *options.OptionSymbol
		var err error

		switch leg.StrikeMethod {
		case strategy.Delta:
			fmt.Printf("got strike method delta\n")
			fmt.Printf("under: %s\n", s.Underlying)
			fmt.Printf("dte: %d\n", leg.DTE)
			fmt.Printf("opt type: %s\n", options.OptionType(leg.OptType))
			fmt.Printf("round: %d\n", leg.Round)
			fmt.Printf("strike meth val: %f\n", leg.StrikeMethVal)
			optData, err := e.optionProvider.OptionDataByDelta(
				s.Underlying,
				leg.DTE,
				options.OptionType(leg.OptType),
				leg.Round,
				leg.StrikeMethVal,
			)
			if err != nil {
				return tasty.NewOrder{}, fmt.Errorf("Error getting option data: %w", err)
			}
			optSymbol, err = options.ParseDxLinkOption(optData.Greek.Symbol)
			if err != nil {
				return tasty.NewOrder{},
					fmt.Errorf("Error parsing optData.Quote.Symbol: %s, %w", optData.Quote.Symbol, err)
			}
			midPrice = (*optData.Quote.AskPrice + *optData.Quote.BidPrice) / 2
			fmt.Printf("mid price for leg %d: %.2f\n", i+1, midPrice)
		case strategy.Relative:
			if i == 0 {
				return tasty.NewOrder{}, fmt.Errorf("Strike Method `Offset` cannot be the first leg")
			}
			prevSymbol := orderLegs[i-1].Symbol
			optSymbol, err = options.ParseOCCOption(prevSymbol)
			if err != nil {
				return tasty.NewOrder{}, fmt.Errorf("Unable to parse OCC Option: %s, %w", prevSymbol, err)
			}
			optSymbol.IncrementStrike(leg.StrikeMethVal)
			optData, err := e.optionProvider.GetOptData(optSymbol.DxLinkString())
			if err != nil {
				return tasty.NewOrder{}, fmt.Errorf("Unable to get Opt Data with symbol: %s, %w", optSymbol.DxLinkString(), err)
			}
			midPrice = (*optData.Quote.AskPrice + *optData.Quote.BidPrice) / 2
			fmt.Printf("mid price for leg %d: %.2f\n", i+1, midPrice)
		}

		if leg.Side == strategy.Buy {
			action = tasty.BTO
			price -= midPrice
		} else {
			action = tasty.STO
			price += midPrice
		}
		fmt.Printf("updated Price: %.2f\n", price)

		if optSymbol == nil {
			return tasty.NewOrder{}, fmt.Errorf("optSymbol is nil for leg %d with strikeMethod %v", i, leg.StrikeMethod)
		}
		fmt.Printf("optSymbol for order leg: %s\n", optSymbol.OCCString())
		orderLegs = append(orderLegs, tasty.NewOrderLeg{
			InstrumentType: tasty.EquityOptionIT,
			Symbol:         optSymbol.OCCString(),
			Quantity:       float64(leg.Quantity),
			Action:         action,
		})
	}

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
	if price > 0.0 {
		effect = tasty.Credit
	} else {
		effect = tasty.Debit
	}
	// TODO: add slippage to price
	return tasty.NewOrder{
		TimeInForce: "Day",
		OrderType:   "Limit",
		Price:       price,
		PriceEffect: effect,
		Legs:        orderLegs,
	}, nil
}

func (e *Engine) startWorkers() {
	for i := 0; i < e.workerCount; i++ {
		e.wg.Add(1)
		go e.worker()
	}
}

func (e *Engine) worker() {
	defer e.wg.Done()

	//for order := range e.orderQueue {
	//resp, err := e.apiClient.SubmitOrderDryRun(e.ctx, acctNum, &order)
	//}
}
