package strategy

import (
	"encoding/json"
	"log/slog"
	"os"
)

type Strategy struct {
	Name            string `json:"name"`
	Underlying      string `json:"underlying"`
	Legs            []Leg  `json:"legs"`
	RiskParams      RiskParams
	EntryConditions map[string]map[string]interface{} `json:"entry-conditions"`
	entryConditions map[string]Condition
	exitConditions  map[string]Condition
}

type Leg struct {
	// call or put
	OptType OptType `json:"option-type"`
	// sell or buy
	Side     OptSide `json:"option-side"`
	Quantity int     `json:"quantity"`
	DTE      int     `json:"days-to-expiration"`
	// delta or offset
	StrikeMethod  StrikeMethod `json:"strike-selection-method"`
	StrikeMethVal float64      `json:"strike-selection-value"`
	Round         int          `json:"round-nearest"`
}

type RiskParams struct {
	PctPortfolio float64
	NumContracts int
}

func FromFile(fpath string, f *ConditionFactory) (Strategy, error) {
	var strat Strategy
	file, err := os.Open(fpath)
	if err != nil {
		return strat, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&strat); err != nil {
		return strat, err
	}
	if strat.EntryConditions != nil {
		conditions, err := f.FromConfig(strat.EntryConditions)
		if err != nil {
			return strat, err
		}
		strat.entryConditions = conditions
	}

	return strat, nil
}

func NewStrategy(
	name,
	underlying string,
	legs []Leg,
	risk RiskParams,
	entries map[string]Condition,
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
		OptType:       optType,
		Side:          side,
		Quantity:      quantity,
		DTE:           dte,
		StrikeMethod:  strikeMethod,
		StrikeMethVal: strikeMethVal,
		Round:         round,
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
