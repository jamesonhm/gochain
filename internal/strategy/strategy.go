package strategy

type strategy struct {
	name            string
	underlying      string
	legs            []OptionLeg
	entryConditions []EntryCondition
	riskParams      RiskParams
	exitConditions  []ExitCondition
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

func NewStrategy(name, underlying string, risk RiskParams, entries ...EntryCondition) *strategy {

}

func (*strategy) WithLeg() *strategy       {}
func (*strategy) WithLinkedLeg() *strategy {}

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
