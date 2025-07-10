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
	return tasty.NewOrder{}
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
