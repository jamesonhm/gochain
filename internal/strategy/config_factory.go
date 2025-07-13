package strategy

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

func (f *ConditionFactory) CreateFromJSON(jsonData []byte) (map[string]Condition, error) {

}
