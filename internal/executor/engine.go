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
	stratStates    StatusTracker
	semaphore      chan struct{}
	wg             sync.WaitGroup
	workerCount    int
	ctx            context.Context
	liveOrder      bool
}

type StatusTracker interface {
	SubmitOrder(string, time.Time, string, tasty.Order)
	NextPFID() int
}

func NewEngine(
	apiClient *tasty.TastyAPI,
	acctNum string,
	optionProvider *dxlink.DxLinkClient,
	stratStates StatusTracker,
	workerCount int,
	ctx context.Context,
	liveOrder bool,
) *Engine {

	e := &Engine{
		apiClient:      apiClient,
		acctNum:        acctNum,
		optionProvider: optionProvider,
		stratStates:    stratStates,
		semaphore:      make(chan struct{}, workerCount),
		workerCount:    workerCount,
		ctx:            ctx,
		liveOrder:      liveOrder,
	}

	//e.startWorkers()
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
	order.Source = s.Name
	bytes, _ := json.MarshalIndent(order, "", "\t")
	fmt.Printf("This is where the order goes into the queue:\n%+v\n", string(bytes))
	//e.stratStates.Submit(s.Name, time.Now().In(dt.TZNY()), pfid, order)
	//e.stratStates.PPrint()
	e.semaphore <- struct{}{}
	e.wg.Add(1)
	go e.worker(order, s)
	e.wg.Wait()
}

func (e *Engine) orderFromStrategy(s strategy.Strategy) (tasty.NewOrder, error) {
	// for each leg, calculate strike price
	// create leg(s)
	// create order struct
	// TODO: change price/midPrice to decimal type
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

	// Add Entry Slippage
	if price > 0.0 {
		price -= (float64(s.EntrySlippage) / 100)
	} else {
		price += (float64(s.EntrySlippage) / 100)
	}
	if price > 0.0 {
		effect = tasty.Credit
	} else {
		effect = tasty.Debit
	}
	return tasty.NewOrder{
		TimeInForce: "Day",
		OrderType:   "Limit",
		Price:       fmt.Sprintf("%.2f", price),
		PriceEffect: effect,
		Legs:        orderLegs,
	}, nil
}

//func (e *Engine) startWorkers() {
//	for i := 1; i < e.workerCount+1; i++ {
//		e.wg.Add(1)
//		go e.worker(i)
//	}
//}

// func (e *Engine) worker(id int) {
func (e *Engine) worker(newOrder tasty.NewOrder, s strategy.Strategy) {
	defer func() {
		<-e.semaphore
		e.wg.Done()
	}()

	//for order := range e.semaphore {
	resp, err := e.apiClient.SubmitOrderDryRun(e.ctx, e.acctNum, &newOrder)
	if err != nil {
		slog.Error("(executor.worker) order dry run", "order", newOrder, "error", err)
		return
	}
	orderResp := resp.OrderResponse.Order
	e.stratStates.SubmitOrder(orderResp.Source, time.Now().In(dt.TZNY()), orderResp.PreflightID, orderResp)

	respbyt, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		slog.Error("(executor.worker) unable to marshal dry run response", "error", err)
	} else {
		fmt.Println(string(respbyt))
	}
	if len(resp.OrderResponse.Warnings) > 0 {
		slog.Warn(
			"(executor.worker) order dry run, will not go live",
			"warnings", resp.OrderResponse.Warnings,
		)
		return
	}
	// TODO: do something with the returned buying power effect or fees?

	// Update qty's for strat Allocation

	if e.liveOrder {
		resp, err = e.apiClient.SubmitOrder(e.ctx, e.acctNum, &newOrder)
		if err != nil {
			slog.Error("(executor.worker) order submit", "order", newOrder, "error", err)
			return
		}
	}
	//}
}

func allocationMultiple(s strategy.Strategy, orderResp tasty.OrderResponse) int {
	return 0
}
