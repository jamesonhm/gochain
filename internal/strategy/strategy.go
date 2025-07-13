package strategy

import "log/slog"

type Strategy struct {
	Name            string `json:"name"`
	Underlying      string `json:"underlying"`
	Legs            []Leg  `json:"legs"`
	RiskParams      RiskParams
	EntryConditions map[string]map[string]interface{} `json:"entry-conditions"`
	entryConditions map[string]EntryCondition
	exitConditions  map[string]ExitCondition
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
	legs []Leg,
	risk RiskParams,
	entries map[string]EntryCondition,
) Strategy {
	strat := Strategy{
		Name:            name,
		Underlying:      underlying,
		Legs:            legs,
		RiskParams:      risk,
		entryConditions: entries,
	}
	return strat
}

func NewLeg(
	optType OptType,
	side OptSide,
	quantity int,
	dte int,
	strikeMethod StrikeMethod,
	strikeMethVal float64,
	round int,
) Leg {
	return Leg{
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

func (s *Strategy) CheckEntryConditions(
	options OptionsProvider,
	candles CandlesProvider,
	portfolio PortfolioProvider,
) bool {
	for name, condition := range s.entryConditions {
		if !condition(options, candles, portfolio) {
			return false
		}
		slog.Info("Condition evaluated True", "strategy", s.Name, "condition", name)
	}
	return true
}
