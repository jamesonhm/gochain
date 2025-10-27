package monitor

import (
	"context"
	"fmt"
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
	stratStates  StatusTracker
	scanInterval time.Duration
}

type StatusTracker interface {
	LastSubmitted(string) (time.Time, error)
}

func NewEngine(
	portfolio *tasty.TastyAPI,
	options *dxlink.DxLinkClient,
	candles *yahoo.YahooAPI,
	executor *executor.Engine,
	stratStates StatusTracker,
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

	fmt.Printf("-------Monitor Started------\n")
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
			slog.LogAttrs(
				ctx,
				slog.LevelDebug,
				"(checkAllStrategies) now not within entry time",
				slog.String("Strategy", s.Name),
				slog.Time("now", time.Now().In(dt.TZNY())),
				slog.String("min time", s.EntryTime.MinTime),
				slog.String("max time", s.EntryTime.MaxTime),
			)
			continue
		}
		slog.Info("(checkAllStrategies) now within entry time")
		// is the last submit time within the entry window
		if subTime, err := e.stratStates.LastSubmitted(s.Name); err == nil {
			if s.TimeInEntry(subTime) {
				slog.LogAttrs(
					ctx,
					slog.LevelInfo,
					"(checkAllStrategies) Already submitted",
					slog.String("Strategy", s.Name),
					slog.Time("LastSubmitTime", subTime),
				)
				continue
			}
		}
		slog.Info("(checkAllStrategies) last submitted not within entry time")
		if s.CheckEntryConditions(e.portfolio, e.candles, e.options) {
			slog.LogAttrs(
				ctx,
				slog.LevelInfo,
				"(checkAllStrategies) Entry Conditions met",
				slog.String("Strategy", s.Name),
			)
			e.executor.SubmitOrder(s)
		}
	}
}
