package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/jamesonhm/gochain/internal/dt"
	"github.com/jamesonhm/gochain/internal/dxlink"
	"github.com/jamesonhm/gochain/internal/options"
	"github.com/jamesonhm/gochain/internal/strategy"
	"github.com/jamesonhm/gochain/internal/tasty"
)

type Engine struct {
	apiClient      *tasty.TastyAPI
	acctNum        string
	optionProvider *dxlink.DxLinkClient
	stratStates    *strategy.Status
	orderQueue     chan tasty.NewOrder
	wg             sync.WaitGroup
	workerCount    int
	ctx            context.Context
}

func NewEngine(
	apiClient *tasty.TastyAPI,
	acctNum string,
	optionProvider *dxlink.DxLinkClient,
	stratStates *strategy.Status,
	workerCount int,
	ctx context.Context,
) *Engine {

	e := &Engine{
		apiClient:      apiClient,
		acctNum:        acctNum,
		optionProvider: optionProvider,
		stratStates:    stratStates,
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
	order, err := e.orderFromStrategy(s)
	if err != nil {
		slog.Error("Unable to create order from strategy:", "error", err)
		return
	}
	pfid := strconv.Itoa(e.stratStates.NextPFID())
	order.PreflightID = pfid
	bytes, _ := json.MarshalIndent(order, "", "\t")
	fmt.Printf("This is where the order goes into the queue: %+v\n", string(bytes))
	e.stratStates.Submit(s.Name, time.Now().In(dt.TZNY()), pfid, order)
	e.stratStates.PPrint()
	e.orderQueue <- order
}

func (e *Engine) orderFromStrategy(s strategy.Strategy) (tasty.NewOrder, error) {
	// for each leg, calculate strike price
	// create leg(s)
	// create order struct
	var price float64
	var effect tasty.PriceEffect
	var holidays []time.Time
	var err error
	holidays, err = e.apiClient.GetMarketHolidaysDT(e.ctx)
	if err != nil {
		slog.Error("(orderFromStrategy) Unable to get market holidays:", "error", err)
		holidays = []time.Time{}
	}

	orderLegs := make([]tasty.NewOrderLeg, 0)
	for i, leg := range s.Legs {
		var action tasty.OrderAction
		var midPrice float64
		var optSymbol *options.OptionSymbol
		var err error

		switch leg.StrikeMethod {
		case strategy.Delta:
			slog.Debug("(orderFromStrategy) Leg delta",
				"underlying", s.Underlying,
				"dte:", leg.DTE,
				"opt type:", options.OptionType(leg.OptType),
				"round:", leg.Round,
				"strike meth val:", leg.StrikeMethVal,
			)
			optData, err := e.optionProvider.OptionDataByDelta(
				s.Underlying,
				leg.DTE,
				options.OptionType(leg.OptType),
				leg.Round,
				leg.StrikeMethVal,
				holidays,
			)
			if err != nil {
				return tasty.NewOrder{}, fmt.Errorf("Error getting option data: %w", err)
			}
			optSymbol, err = options.ParseDxLinkOption(optData.Greek.Symbol)
			if err != nil {
				return tasty.NewOrder{},
					fmt.Errorf("Error parsing optData.Greek.Symbol: %s, %w", optData.Greek.Symbol, err)
			}
			midPrice = (*optData.Quote.AskPrice + *optData.Quote.BidPrice) / 2
			fmt.Printf("(orderFromStrategy) mid price for leg %d: %.2f\n", i+1, midPrice)
		case strategy.Relative:
			if i == 0 {
				return tasty.NewOrder{}, fmt.Errorf("Strike Method `Relative` cannot be the first leg")
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
			fmt.Printf("(orderFromStrategy) mid price for leg %d: %.2f\n", i+1, midPrice)
		}

		if leg.Side == strategy.Buy {
			action = tasty.BTO
			price -= midPrice
		} else {
			action = tasty.STO
			price += midPrice
		}
		fmt.Printf("(orderFromStrategy) updated Price: %.2f\n", price)

		if optSymbol == nil {
			return tasty.NewOrder{}, fmt.Errorf("optSymbol is nil for leg %d with strikeMethod %v", i, leg.StrikeMethod)
		}
		fmt.Printf("(orderFromStrategy) optSymbol for order leg: %s\n", optSymbol.OCCString())
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
		Price:       fmt.Sprintf("%.2f", price),
		PriceEffect: effect,
		Legs:        orderLegs,
	}, nil
}

func (e *Engine) startWorkers() {
	for i := 1; i < e.workerCount+1; i++ {
		e.wg.Add(1)
		go e.worker(i)
	}
}

func (e *Engine) worker(id int) {
	defer e.wg.Done()

	for order := range e.orderQueue {
		resp, err := e.apiClient.SubmitOrderDryRun(e.ctx, e.acctNum, &order)
		if err != nil {
			slog.Error("(executor.worker) order dry run", "workerid", id, "order", order, "error", err)
			continue
		}
		// TODO: do something with the returned buying power effect or fees?
		respbyt, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			slog.Error("(executor.worker) unable to marshal dry run response", "workerid", id, "error", err)
		} else {
			fmt.Println(string(respbyt))
		}

		resp, err = e.apiClient.SubmitOrder(e.ctx, e.acctNum, &order)
		if err != nil {
			slog.Error("(executor.worker) order submit", "workerid", id, "order", order, "error", err)
			continue
		}
	}
}
