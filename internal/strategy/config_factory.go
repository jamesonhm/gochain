package strategy

import (
	"fmt"
	"time"
)

type ConditionFactory struct {
	factories map[string]FactoryFunc
}

type FactoryFunc func(params map[string]interface{}) (Condition, error)

func NewConditionFactory() *ConditionFactory {
	factory := &ConditionFactory{
		factories: make(map[string]FactoryFunc),
	}

	factory.RegisterFactory("day-of-week", createDayOfWeekCondition)

	return factory
}

func (f *ConditionFactory) RegisterFactory(name string, factory FactoryFunc) {
	f.factories[name] = factory
}

//func (f *ConditionFactory) CreateFromJSON(jsonData []byte) (map[string]Condition, error) {
//
//}

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

	return EntryDayOfWeek(weekdays...), nil
}
