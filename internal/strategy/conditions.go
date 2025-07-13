package strategy

import (
	"log/slog"
	"time"
)

type Condition func(
	options OptionsProvider,
	candles CandlesProvider,
	portfolio PortfolioProvider,
) bool

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

func EntryDayOfWeek(allowedDays ...time.Weekday) EntryCondition {
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

func TimeRange(startHour, startMin, endHour, endMin int) EntryCondition {
	return func(_ OptionsProvider, _ CandlesProvider, _ PortfolioProvider) bool {
		nytz, err := time.LoadLocation("America/New_York")
		if err != nil {
			return false
		}
		now := time.Now().In(nytz)
		start := time.Date(now.Year(), now.Month(), now.Day(), startHour, startMin, 0, 0, nytz)
		end := time.Date(now.Year(), now.Month(), now.Day(), endHour, endMin, 0, 0, nytz)
		return now.After(start) && now.Before(end)
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
