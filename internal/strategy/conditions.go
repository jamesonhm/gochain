package strategy

import (
	"log/slog"
	"time"
	//"github.com/jamesonhm/gochain/internal/dxlink"
	//"github.com/jamesonhm/gochain/internal/tasty"
	//"github.com/jamesonhm/gochain/internal/yahoo"
)

type EntryCondition func(
	options OptionsProvider,
	candles CandlesProvider,
	portfolio PortfolioProvider,
) bool
type ExitCondition func(
	options OptionsProvider,
	candles CandlesProvider,
	portfolio PortfolioProvider,
) bool
type StrikeCalc func(options OptionsProvider) float64

func EntryDayOfWeeek(allowedDays []time.Weekday) EntryCondition {
	return func(_ OptionsProvider, _ CandlesProvider, _ PortfolioProvider) bool {
		today := time.Now().Weekday()
		for _, day := range allowedDays {
			if day == today {
				return true
			}
		}
		return false
	}
}

func VixONMoveMin(threshold float64) EntryCondition {
	return func(_ OptionsProvider, candles CandlesProvider, _ PortfolioProvider) bool {
		vixMove, err := candles.ONMove("^VIX")
		if err != nil {
			slog.Error("VixONMoveMin Entry Condition", "error", err)
			return false
		}
		return vixMove >= threshold
	}
}

func VixONMoveMax(threshold float64) EntryCondition {
	return func(_ OptionsProvider, candles CandlesProvider, _ PortfolioProvider) bool {
		vixMove, err := candles.ONMove("^VIX")
		if err != nil {
			slog.Error("VixONMoveMax Entry Condition", "error", err)
			return false
		}
		return vixMove <= threshold
	}
}
