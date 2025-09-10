package strategy

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jamesonhm/gochain/internal/dt"
)

const (
	MIN_TIME = "9:32AM"
	MAX_TIME = "3:58PM"
)

type Strategy struct {
	Name            string    `json:"name"`
	Underlying      string    `json:"underlying"`
	Legs            []Leg     `json:"legs"`
	EntryTime       EntryTime `json:"entry-time"`
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

// Entry window times in the "Kitchen" format/layout (3:04PM)
type EntryTime struct {
	MinTime string `json:"min-time"`
	MaxTime string `json:"max-time"`
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

	if err := strat.validateEntryTimes(); err != nil {
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

func (s *Strategy) ListDTEs() []int {
	var dtes []int
	for _, leg := range s.Legs {
		dtes = append(dtes, leg.DTE)
	}
	return dtes
}

// check if a time is within the current entry time window
func (s *Strategy) TimeInEntry(t time.Time) bool {
	if t.After(dt.ParseTimeAsToday(s.EntryTime.MinTime)) &&
		t.Before(dt.ParseTimeAsToday(s.EntryTime.MaxTime)) {
		return true
	}
	return false
}

func (s *Strategy) validateEntryTimes() error {
	var t time.Time
	var err error
	if s.EntryTime.MinTime == "" {
		return fmt.Errorf("(strategy: `%s`) EntryTime.MinTime is required", s.Name)
	}
	if t, err = time.Parse(time.Kitchen, s.EntryTime.MinTime); err != nil {
		return fmt.Errorf("(strategy: `%s`) Invalid format for EntryTime.Min: %s, should be `3:40PM`", s.Name, s.EntryTime.MinTime)
	}
	minmin, _ := time.Parse(time.Kitchen, MIN_TIME)
	maxmax, _ := time.Parse(time.Kitchen, MAX_TIME)
	if t.Before(minmin) {
		return fmt.Errorf("(strategy: `%s`) EntryTime.MinTime is before %s", s.Name, MIN_TIME)
	} else if t.After(maxmax) {
		return fmt.Errorf("(strategy: `%s`) EntryTime.MinTime is after %s", s.Name, MAX_TIME)
	}

	if s.EntryTime.MaxTime == "" {
		s.EntryTime.MaxTime = dt.ParseTimeAsToday(s.EntryTime.MinTime).Add(1 * time.Minute).Format(time.Kitchen)
	}
	if t, err = time.Parse(time.Kitchen, s.EntryTime.MaxTime); err != nil {
		return fmt.Errorf("(strategy: `%s`) Invalid format for EntryTime.MaxTime: %s, should be `3:40PM`", s.Name, s.EntryTime.MaxTime)
	}
	if t.After(maxmax) {
		return fmt.Errorf("(strategy: `%s`) EntryTime.MaxTime is after %s", s.Name, MAX_TIME)
	} else if t.Before(minmin) {
		return fmt.Errorf("(strategy: `%s`) EntryTime.MaxTime is before %s", s.Name, MIN_TIME)
	}

	return nil
}
