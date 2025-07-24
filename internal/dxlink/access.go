package dxlink

import (
	"fmt"
	"math"
	"time"

	"github.com/jamesonhm/gochain/internal/dt"
	"github.com/jamesonhm/gochain/internal/options"
)

// searches the map of optionSubs for the date, and strike nearest the delta based on the rounding value
func (c *DxLinkClient) StrikeFromDelta(
	underlying string,
	currentPrice float64,
	dte int,
	optType options.OptionType,
	round int,
	targetDelta float64,
) (string, error) {
	// find exp date
	exp := time.Now().AddDate(0, 0, dte)
	if exp.Weekday() < 1 || exp.Weekday() > 5 {
		exp = dt.NextWeekday(exp)
	}
	c.mu.RLock()
	defer c.mu.RUnlock()

	s := roundNearest(currentPrice, round)
	opt := options.OptionSymbol{
		Underlying: underlying,
		Date:       exp,
		Strike:     s,
		OptionType: optType,
	}
	data, ok := c.optionSubs[opt.DxLinkString()]
	if !ok {
		return "", fmt.Errorf("unable to find option in subscription data: %s", opt.DxLinkString())
	}
	delta := *data.Greek.Delta
	dist := math.Abs(delta - targetDelta)
	// start a loop and increment to closest delta or delta = 0
	if delta > targetDelta {
		for s += float64(round); s < s*1.1; s += float64(round) {
			opt.IncrementStrike(float64(round))
			data, ok = c.optionSubs[opt.DxLinkString()]
			if !ok {
				return "", fmt.Errorf("unable to find option in subscription data: %s", opt.DxLinkString())
			}
			delta = *data.Greek.Delta
			if math.Abs(delta-targetDelta) > dist {
				opt.DecrementStrike(float64(round))
				return opt.DxLinkString(), nil
			}
			dist = math.Abs(delta - targetDelta)
		}
		return "", fmt.Errorf("no option found incrementing")
	} else {
		for s -= float64(round); s > s*0.9; s -= float64(round) {
			opt.DecrementStrike(float64(round))
			data, ok = c.optionSubs[opt.DxLinkString()]
			if !ok {
				return "", fmt.Errorf("unable to find option in subscription data: %s", opt.DxLinkString())
			}
			delta = *data.Greek.Delta
			if math.Abs(delta-targetDelta) > dist {
				opt.IncrementStrike(float64(round))
				return opt.DxLinkString(), nil
			}
			dist = math.Abs(delta - targetDelta)
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
