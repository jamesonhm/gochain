package monitor

import (
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
	strategies []strategy.StrategyConfig
	//executor *executor.Engine
	scanInterval time.Duration
}

func NewEngine(
	portfolio *tasty.TastyAPI,
	options *dxlink.DxLinkClient,
	candles *yahoo.YahooAPI,
	strategies []strategy.StrategyConfig,
	scanInterval time.Duration,
) *Engine {
	return &Engine{
		portfolio:    portfolio,
		options:      options,
		candles:      candles,
		strategies:   strategies,
		scanInterval: scanInterval,
	}
}
