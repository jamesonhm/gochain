package main

import (
	"fmt"

	"github.com/jamesonhm/gochain/internal/strategy"
)

func loadStrategies() ([]strategy.Strategy, error) {
	strats := make([]strategy.Strategy, 0)
	conditionFactory := strategy.NewConditionFactory()
	strat, err := strategy.FromFile("examples/basic.json", conditionFactory)
	if err != nil {
		return nil, fmt.Errorf("err in strategy from file: %w", err)
	}
	fmt.Printf("Strategy: %+v\n", strat)
	strats = append(strats, strat)

	return strats, nil
}
