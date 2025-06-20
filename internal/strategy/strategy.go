package strategy

type Strategy struct {
	Name            string
	Underlying      string
	Legs            []OptionLeg
	EntryConditions map[string]EntryCondition
	RiskParams      RiskParams
	ExitConditions  map[string]ExitCondition
}

type OptionLeg struct {
	Type     string // call or put
	Side     string // sell or buy
	Quantity int
	DTE      int
	Strike   StrikeCalc
}

type RiskParams struct {
	PctPortfolio float64
	NumContracts int
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
	for _, condition := range s.EntryConditions {
		if !condition(options, candles, portfolio) {
			return false
		}
	}
	return true
}
