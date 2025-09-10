package monitor

import (
	"context"
	"log/slog"
	"time"

	"github.com/jamesonhm/gochain/internal/dt"
	"github.com/jamesonhm/gochain/internal/dxlink"
	"github.com/jamesonhm/gochain/internal/executor"
	"github.com/jamesonhm/gochain/internal/strategy"
	"github.com/jamesonhm/gochain/internal/tasty"
	"github.com/jamesonhm/gochain/internal/yahoo"
)

type Engine struct {
	portfolio    *tasty.TastyAPI
	options      *dxlink.DxLinkClient
	candles      *yahoo.YahooAPI
	strategies   []strategy.Strategy
	executor     *executor.Engine
	stratStates  *strategy.StratStates
	scanInterval time.Duration
}

func NewEngine(
	portfolio *tasty.TastyAPI,
	options *dxlink.DxLinkClient,
	candles *yahoo.YahooAPI,
	executor *executor.Engine,
	stratStates *strategy.StratStates,
	scanInterval time.Duration,
) *Engine {
	return &Engine{
		portfolio:    portfolio,
		options:      options,
		candles:      candles,
		executor:     executor,
		stratStates:  stratStates,
		scanInterval: scanInterval,
	}
}

func (e *Engine) AddStrategy(s strategy.Strategy) {
	e.strategies = append(e.strategies, s)
}

func (e *Engine) Run(ctx context.Context) {
	ticker := time.NewTicker(e.scanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.checkAllStrategies(ctx)
		}
	}
}

func (e *Engine) checkAllStrategies(ctx context.Context) {
	for _, s := range e.strategies {
		// is "now" within the entry window
		if !s.TimeInEntry(time.Now().In(dt.TZNY())) {
			continue
		}
		// is the last submit time within the entry window
		if subTime, err := e.stratStates.LastSubmitted(s.Name); err == nil {
			if s.TimeInEntry(subTime) {
				slog.Info("(checkAllStrategies) Already submitted", "Strategy", s.Name, "LastSubmitTime", subTime)
				continue
			}
		}
		if s.CheckEntryConditions(e.portfolio, e.candles, e.options) {
			slog.LogAttrs(ctx, slog.LevelInfo, "Entry Conditions met", slog.String("strat name", s.Name))
			e.executor.SubmitOrder(s)
		}
	}
}
