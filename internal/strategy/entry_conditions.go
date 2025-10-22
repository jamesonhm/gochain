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

func strInterToDec(strInter interface{}) (decimal.Decimal, error) {
	if val, ok := strInter.(string); ok {
		dec, err := decimal.NewFromString(val)
		if err != nil {
			return decimal.NewFromInt(0), err
		}
		return dec, nil
	} else {
		return decimal.NewFromInt(0), fmt.Errorf("unable to type cast to string: %v", strInter)
	}
}

func createVixONMoveCondition(params map[string]interface{}) (Condition, error) {
	minInter, minOk := params["min"]
	maxInter, maxOk := params["max"]
	if !minOk && !maxOk {
		return nil, fmt.Errorf("VIX ON Move Condition requires at least one of `min` or `max`")
	}
	var units string
	var ok bool
	if unitsInter, unitsOk := params["units"]; !unitsOk {
		units = "percent"
	} else {
		if units, ok = unitsInter.(string); !ok {
			units = "percent"
		}
	}

	var minParam decimal.Decimal
	var maxParam decimal.Decimal
	var err error
	if minOk {
		minParam, err = strInterToDec(minInter)
		if err != nil {
			return nil, fmt.Errorf("VIX ON Move unable to get decimal from min param: %v, %w", minInter, err)
		}
	} else {
		minParam = decimal.NewFromInt(-999)
	}
	if maxOk {
		maxParam, err = strInterToDec(maxInter)
		if err != nil {
			return nil, fmt.Errorf("VIX ON Move unable to get decimal from max param: %v, %w", maxInter, err)
		}
	} else {
		maxParam = decimal.NewFromInt(999)
	}
	return func(_ OptionsProvider, candles CandlesProvider, _ PortfolioProvider) bool {
		var vixMoveD decimal.Decimal
		switch units {
		case "percent":
			vixMove, err := candles.ONMovePct("^VIX")
			if err != nil {
				slog.Error("Unable to get VixONMovePct for Entry Condition", "error", err)
				return false
			}
			vixMoveD = decimal.NewFromFloat(vixMove)
		case "absolute":
			vixMove, err := candles.ONMove("^VIX")
			if err != nil {
				slog.Error("Unable to get VixONMove for Entry Condition", "error", err)
				return false
			}
			vixMoveD = decimal.NewFromFloat(vixMove)
		}

		return minParam.LessThanOrEqual(vixMoveD) && maxParam.GreaterThanOrEqual(vixMoveD)
	}, nil
}
