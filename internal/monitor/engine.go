package monitor

import (
	"context"
	"log/slog"
	"time"

	"github.com/jamesonhm/gochain/internal/dxlink"
	"github.com/jamesonhm/gochain/internal/strategy"
	"github.com/jamesonhm/gochain/internal/tasty"
	"github.com/jamesonhm/gochain/internal/yahoo"
)

type Engine struct {
	portfolio  *tasty.TastyAPI
	options    *dxlink.DxLinkClient
	candles    *yahoo.YahooAPI
	strategies []strategy.Strategy
	//executor *executor.Engine
	scanInterval time.Duration
}

func NewEngine(
	portfolio *tasty.TastyAPI,
	options *dxlink.DxLinkClient,
	candles *yahoo.YahooAPI,
	scanInterval time.Duration,
) *Engine {
	return &Engine{
		portfolio:    portfolio,
		options:      options,
		candles:      candles,
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
		if s.CheckEntryConditions(e.portfolio, e.candles, e.options) {
			slog.LogAttrs(ctx, slog.LevelInfo, "Entry Conditions met", slog.String("strat name", s.Name))
		}
	}
}
