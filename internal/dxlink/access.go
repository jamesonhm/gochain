package dxlink

import (
	"fmt"
	"math"
	"time"

	"github.com/jamesonhm/gochain/internal/dt"
	"github.com/jamesonhm/gochain/internal/options"
	"github.com/jamesonhm/gochain/internal/strategy"
)

// searches the map of optionSubs for the date, and strike nearest the delta based on the rounding value
func (c *DxLinkClient) StrikeFromDelta(
	underlying string,
	currentPrice float64,
	leg strategy.Leg,
) (string, error) {
	// find exp date
	exp := time.Now().AddDate(0, 0, leg.DTE)
	if exp.Weekday() < 1 || exp.Weekday() > 5 {
		exp = dt.NextWeekday(exp)
	}
	c.mu.RLock()
	defer c.mu.RUnlock()

	// TODO: change to func "roundNearest" (modulo round value then subtract)
	s := math.Floor(currentPrice)
	opt := options.OptionSymbol{
		Underlying: underlying,
		Date:       exp,
		Strike:     s,
		OptionType: options.OptionType(leg.OptType),
	}.DxLinkString()

	data, ok := c.optionSubs[opt]
	if !ok {
		return "", fmt.Errorf("unable to find option in subscription data: %s", opt)
	}
	delta := *data.Greek.Delta
	dist := math.Abs(delta - leg.StrikeMethVal)
	// start a loop and increment to closest delta or delta = 0
	if delta > leg.StrikeMethVal {
		for s += float64(leg.Round); s < s*1.1; s += float64(leg.Round) {
			opt = options.OptionSymbol{
				Underlying: underlying,
				Date:       exp,
				Strike:     s,
				OptionType: options.OptionType(leg.OptType),
			}.DxLinkString()
			data, ok = c.optionSubs[opt]
			if !ok {
				return "", fmt.Errorf("unable to find option in subscription data: %s", opt)
			}
			delta = *data.Greek.Delta
			if math.Abs(delta-leg.StrikeMethVal) > dist {
				return options.OptionSymbol{
					Underlying: underlying,
					Date:       exp,
					Strike:     s - float64(leg.Round),
					OptionType: options.OptionType(leg.OptType),
				}.DxLinkString(), nil
			}
			dist = math.Abs(delta - leg.StrikeMethVal)
		}
		return "", fmt.Errorf("no option found incrementing")
	} else {
		for s -= float64(leg.Round); s > s*0.9; s -= float64(leg.Round) {
			opt = options.OptionSymbol{
				Underlying: underlying,
				Date:       exp,
				Strike:     s,
				OptionType: options.OptionType(leg.OptType),
			}.DxLinkString()
			data, ok = c.optionSubs[opt]
			if !ok {
				return "", fmt.Errorf("unable to find option in subscription data: %s", opt)
			}
			delta = *data.Greek.Delta
			if math.Abs(delta-leg.StrikeMethVal) > dist {
				return options.OptionSymbol{
					Underlying: underlying,
					Date:       exp,
					Strike:     s + float64(leg.Round),
					OptionType: options.OptionType(leg.OptType),
				}.DxLinkString(), nil
			}
			dist = math.Abs(delta - leg.StrikeMethVal)
		}
		return "", fmt.Errorf("no option found decrementing")
	}
}

func roundNearest(price float64, round int) float64 {
	if round == 0 || round == 1 {
		return math.Floor(price)
	}
	diff := int(math.Floor(price)) % round
	return math.Floor(price) - float64(diff)
}
