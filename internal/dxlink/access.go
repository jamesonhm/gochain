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
	fmt.Printf("StrikeFromDelta: Exp Date: %s\n", exp)
	c.mu.RLock()
	defer c.mu.RUnlock()

	s := roundNearest(currentPrice, round)
	opt := options.OptionSymbol{
		Underlying: underlying,
		Date:       exp,
		Strike:     s,
		OptionType: optType,
	}
	delta, err := c.getOptDelta(opt.DxLinkString())
	if err != nil {
		return "", fmt.Errorf("unable to find INITIAL option in subscription data: %s, %w", opt.DxLinkString(), err)
	}

	dist := math.Abs(delta - targetDelta)
	fmt.Printf("StrikeFromDelta: Initial values, option: %s, delta: %.6f, distance: %.6f\n", opt.DxLinkString(), delta, dist)
	// start a loop and increment to closest delta or delta = 0
	if delta > targetDelta {
		for s += float64(round); s < s*1.1; s += float64(round) {
			opt.IncrementStrike(float64(round))
			delta, err := c.getOptDelta(opt.DxLinkString())
			if err != nil {
				return "", fmt.Errorf("unable to find option in subscription data: %s, %w", opt.DxLinkString(), err)
			}
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
			delta, err := c.getOptDelta(opt.DxLinkString())
			if err != nil {
				return "", fmt.Errorf("unable to find option in subscription data: %s, %w", opt.DxLinkString(), err)
			}
			if math.Abs(delta-targetDelta) > dist {
				opt.IncrementStrike(float64(round))
				return opt.DxLinkString(), nil
			}
			dist = math.Abs(delta - targetDelta)
		}
		return "", fmt.Errorf("no option found decrementing")
	}
}

func (c *DxLinkClient) getOptDelta(opt string) (float64, error) {
	optionDataPtr, ok := c.optionSubs[opt]
	if !ok {
		return 0, fmt.Errorf("unable to find option in subscription data: %s", opt)
	}
	if optionDataPtr == nil {
		return 0, fmt.Errorf("optionData is a nil ptr")
	} else if optionDataPtr.Greek.Delta == nil {
		fmt.Printf("OptionDataPtr: %+v\n", optionDataPtr)
		return 0, fmt.Errorf("optionData.Greek.Delta is a nil ptr")
	}
	delta := *optionDataPtr.Greek.Delta
	return delta, nil
}

func roundNearest(price float64, round int) float64 {
	if round == 0 || round == 1 {
		return math.Floor(price)
	}
	diff := int(math.Floor(price)) % round
	return math.Floor(price) - float64(diff)
}
