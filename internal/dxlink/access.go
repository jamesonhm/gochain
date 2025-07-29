package dxlink

import (
	"fmt"
	"math"

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
) (*OptionData, error) {
	// find exp date
	exp := dt.DTEToDate(dte)
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
		return nil, fmt.Errorf("unable to find INITIAL option in subscription data: %s, %w", opt.DxLinkString(), err)
	}

	dist := math.Abs(delta - targetDelta)
	fmt.Printf("StrikeFromDelta: Initial values, option: %s, delta: %.6f, distance: %.6f\n", opt.DxLinkString(), delta, dist)
	// start a loop and increment to closest delta or delta = 0
	if delta > targetDelta {
		for s += float64(round); s < s*1.1; s += float64(round) {
			opt.IncrementStrike(float64(round))
			delta, err := c.getOptDelta(opt.DxLinkString())
			if err != nil {
				return nil, fmt.Errorf("unable to find option in subscription data: %s, %w", opt.DxLinkString(), err)
			}
			if math.Abs(delta-targetDelta) > dist {
				opt.DecrementStrike(float64(round))
				optDataPtr, err := c.getOptData(opt.DxLinkString())
				if err != nil {
					return nil, err
				}
				fmt.Printf("Opt Data Ptr: %+v\n", *optDataPtr)
				return optDataPtr, nil
			}
			dist = math.Abs(delta - targetDelta)

			fmt.Printf("StrikeFromDelta: option: %s, delta: %.6f, distance: %.6f\n", opt.DxLinkString(), delta, dist)
		}
		return nil, fmt.Errorf("no option found incrementing")
	} else {
		for s -= float64(round); s > s*0.9; s -= float64(round) {
			opt.DecrementStrike(float64(round))
			delta, err := c.getOptDelta(opt.DxLinkString())
			if err != nil {
				return nil, fmt.Errorf("unable to find option in subscription data: %s, %w", opt.DxLinkString(), err)
			}
			if math.Abs(delta-targetDelta) > dist {
				opt.IncrementStrike(float64(round))
				optDataPtr, err := c.getOptData(opt.DxLinkString())
				if err != nil {
					return nil, err
				}
				fmt.Printf("Opt Data Ptr: %+v\n", optDataPtr)
				return optDataPtr, nil
			}
			dist = math.Abs(delta - targetDelta)

			fmt.Printf("StrikeFromDelta: option: %s, delta: %.6f, distance: %.6f\n", opt.DxLinkString(), delta, dist)
		}
		return nil, fmt.Errorf("no option found decrementing")
	}
}

func (c *DxLinkClient) getOptData(opt string) (*OptionData, error) {
	optionDataPtr, ok := c.optionSubs[opt]
	if !ok {
		return nil, fmt.Errorf("unable to find option in subscription data: %s", opt)
	}
	if optionDataPtr == nil {
		return nil, fmt.Errorf("optionData is a nil ptr")
	}
	return optionDataPtr, nil
}

func (c *DxLinkClient) getOptDelta(opt string) (float64, error) {
	optionDataPtr, err := c.getOptData(opt)
	if err != nil {
		return 0, err
	}
	if optionDataPtr.Greek.Delta == nil {
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
