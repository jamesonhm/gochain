package strategy

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
)

type Condition func(
	options OptionsProvider,
	candles CandlesProvider,
	portfolio PortfolioProvider,
) bool

// Checks that may require `stratStates` to be part of the monitor
// Max Open Trades

// Factory functions for each condition type
func createDayOfWeekCondition(params map[string]interface{}) (Condition, error) {
	daysInterface, ok := params["days"]
	if !ok {
		return nil, fmt.Errorf("missing 'days' parameter")
	}

	// Handle both []string and []interface{} from JSON
	var dayStrings []string
	switch v := daysInterface.(type) {
	case []string:
		dayStrings = v
	case []interface{}:
		for _, day := range v {
			if dayStr, ok := day.(string); ok {
				dayStrings = append(dayStrings, dayStr)
			} else {
				return nil, fmt.Errorf("invalid day format: %v", day)
			}
		}
	default:
		return nil, fmt.Errorf("days must be an array of strings")
	}

	// Convert string days to time.Weekday
	var weekdays []time.Weekday
	dayMap := map[string]time.Weekday{
		"sun": time.Sunday, "sunday": time.Sunday,
		"mon": time.Monday, "monday": time.Monday,
		"tues": time.Tuesday, "tuesday": time.Tuesday,
		"weds": time.Wednesday, "wednesday": time.Wednesday,
		"thurs": time.Thursday, "thursday": time.Thursday,
		"fri": time.Friday, "friday": time.Friday,
		"sat": time.Saturday, "saturday": time.Saturday,
	}

	for _, dayStr := range dayStrings {
		if weekday, exists := dayMap[dayStr]; exists {
			weekdays = append(weekdays, weekday)
		} else {
			return nil, fmt.Errorf("invalid day: %s", dayStr)
		}
	}

	return func(_ OptionsProvider, _ CandlesProvider, _ PortfolioProvider) bool {
		today := time.Now().Weekday()
		for _, day := range weekdays {
			if day == today {
				return true
			}
		}
		return false
	}, nil
}

func createVixONMoveCondition(params map[string]interface{}) (Condition, error) {
	minInter, minOk := params["min"]
	maxInter, maxOk := params["max"]
	if !minOk && !maxOk {
		return nil, fmt.Errorf("VIX ON Move Condition requires at least one of `min` or `max`")
	}

	var minParam decimal.Decimal
	var maxParam decimal.Decimal
	var err error
	if minOk {
		if minStr, ok := minInter.(string); ok {
			minParam, err = decimal.NewFromString(minStr)
			if err != nil {
				return nil, fmt.Errorf("VIX ON Move unable to convert `min` param to decimal: %s", minStr)
			}
		} else {
			return nil, fmt.Errorf("VIX ON Move `min` parameters should be entered as strings: %v", minInter)
		}
	} else {
		minParam = decimal.NewFromInt(-999)
	}
	if maxOk {
		if maxStr, ok := maxInter.(string); ok {
			maxParam, err = decimal.NewFromString(maxStr)
			if err != nil {
				return nil, fmt.Errorf("VIX ON Move unable to convert `max` param to decimal: %s", maxStr)
			}
		} else {
			return nil, fmt.Errorf("VIX ON Move `max` parameters should be entered as strings: %v", maxInter)
		}
	} else {
		maxParam = decimal.NewFromInt(999)
	}
	return func(_ OptionsProvider, candles CandlesProvider, _ PortfolioProvider) bool {
		vixMove, err := candles.ONMove("^VIX")
		if err != nil {
			slog.Error("Unable to get VixONMove for Entry Condition", "error", err)
			return false
		}
		vixMoveD := decimal.NewFromFloat(vixMove)
		return minParam.LessThanOrEqual(vixMoveD) && maxParam.GreaterThanOrEqual(vixMoveD)
	}, nil
}
