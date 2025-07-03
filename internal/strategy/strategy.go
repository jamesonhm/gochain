package strategy

type StrategyConfig struct {
	Name            string
	Underlying      string
	Legs            []*Leg
	RiskParams      RiskParams
	EntryParams     map[string]interface{}
	entryConditions map[string]EntryCondition
	ExitParams      map[string]interface{}
	exitConditions  []ExitCondition
}

type Leg struct {
	optType       OptType // call or put
	side          OptSide // sell or buy
	quantity      int
	dte           int
	strikeMethod  StrikeMethod
	strikeMethVal float64
	round         int
}

type RiskParams struct {
	PctPortfolio float64
	NumContracts int
}

func NewStrategy(
	name,
	underlying string,
	legs []*Leg,
	risk RiskParams,
	entries map[string]EntryCondition,
) *StrategyConfig {
	strat := &StrategyConfig{
		Name:            name,
		Underlying:      underlying,
		Legs:            legs,
		RiskParams:      risk,
		entryConditions: entries,
	}
	return strat
}

func StrategyFromFile(filpath string) (*StrategyConfig, error) {
	return nil, nil
}

func NewLeg(
	optType OptType,
	side OptSide,
	quantity int,
	dte int,
	strikeMethod StrikeMethod,
	strikeMethVal float64,
	round int,
) *Leg {
	return &Leg{
		optType:       optType,
		side:          side,
		quantity:      quantity,
		dte:           dte,
		strikeMethod:  strikeMethod,
		strikeMethVal: strikeMethVal,
		round:         round,
	}
}

type OptionsProvider interface {
}

type CandlesProvider interface {
	ONMove(string) (float64, error)
}

type PortfolioProvider interface {
}

func (s *StrategyConfig) CheckEntryConditions(
	options OptionsProvider,
	candles CandlesProvider,
	portfolio PortfolioProvider,
) bool {
	for _, condition := range s.entryConditions {
		if !condition(options, candles, portfolio) {
			return false
		}
	}
	return true
}
