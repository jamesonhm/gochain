package main

import (
	"fmt"
	"log/slog"

	"github.com/jamesonhm/gochain/internal/strategy"
)

func loadStrategies() []strategy.Strategy {
	//entries := make(map[string]strategy.EntryCondition)
	//entries["days"] = strategy.EntryDayOfWeek(time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday)
	//entries["d2"] = strategy.EntryDayOfWeek(1, 2, 3)

	//strat := strategy.NewStrategy(
	//	"test",
	//	"^XSP",
	//	[]strategy.Leg{
	//		strategy.NewLeg(strategy.Put, strategy.Sell, 1, 45, strategy.Delta, 35, 5),
	//		strategy.NewLeg(strategy.Put, strategy.Buy, 1, 45, strategy.Offset, -5, 0),
	//	},
	//	strategy.RiskParams{
	//		PctPortfolio: 100,
	//		NumContracts: 1,
	//	},
	//	entries,
	//)
	strats := make([]strategy.Strategy, 0)
	conditionFactory := strategy.NewConditionFactory()
	strat, err := strategy.FromFile("examples/basic.json", conditionFactory)
	if err != nil {
		slog.Error("unable to create strategy from file", "err", err)
	}
	fmt.Printf("Strategy: %+v\n", strat)
	strats = append(strats, strat)

	return strats
}
