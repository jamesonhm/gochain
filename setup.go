package main

import (
	"fmt"
	"log/slog"

	"github.com/jamesonhm/gochain/internal/strategy"
)

func loadStrategies() []strategy.Strategy {
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
