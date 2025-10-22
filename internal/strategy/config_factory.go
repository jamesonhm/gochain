package strategy

import (
	"fmt"
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
	factory.RegisterFactory("vix-overnight-move", createVixONMoveCondition)

	return factory
}

func (f *ConditionFactory) RegisterFactory(name string, factory FactoryFunc) {
	f.factories[name] = factory
}

func (f *ConditionFactory) FromConfig(raw map[string]map[string]interface{}) (map[string]Condition, error) {
	conditions := make(map[string]Condition)

	for name, params := range raw {
		factory, exists := f.factories[name]
		if !exists {
			return nil, fmt.Errorf("unknown condition type: %s", name)
		}

		condition, err := factory(params)
		if err != nil {
			return nil, fmt.Errorf("failed to create condition %s: %w", name, err)
		}

		conditions[name] = condition
	}

	return conditions, nil
}
