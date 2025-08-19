package dxlink

import (
	"fmt"
	"math"
	"time"

	"github.com/jamesonhm/gochain/internal/dt"
	"github.com/jamesonhm/gochain/internal/options"
)

func (c *DxLinkClient) OptionDataByOffset(
	underlying string,
	dte int,
	optType options.OptionType,
	offsetFrom float64,
	offsetBy int,
) (*OptionData, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	exp := dt.DTEToDate(dte)
	s := float64(int(offsetFrom) + offsetBy)
	opt := options.OptionSymbol{
		Underlying: underlying,
		Date:       exp,
		Strike:     s,
		OptionType: optType,
	}
	data, err := c.GetOptData(opt.DxLinkString())
	if err != nil {
		return nil, fmt.Errorf("OptionDataByOffset: unable to find INITIAL option in subscription data: %s, %w", opt.DxLinkString(), err)
	}
	return data, nil
}

// searches the map of optionSubs for the date, and strike nearest the delta based on the rounding value
func (c *DxLinkClient) OptionDataByDelta(
	underlying string,
	dte int,
	optType options.OptionType,
	round int,
	targetDelta float64,
) (*OptionData, error) {
	fmt.Println("got to Option data by delta")
	c.mu.RLock()
	defer c.mu.RUnlock()

	// find exp date
	exp := dt.DTEToDate(dte)
	atm, err := c.getUnderlyingPrice(underlying)
	if err != nil {
		return nil, fmt.Errorf("OptionDataByDelta: unable to get underlying price for '%s'\n", underlying)
	}
	s := roundNearest(atm, round)
	opt := options.OptionSymbol{
		Underlying: underlying,
		Date:       exp,
		Strike:     s,
		OptionType: optType,
	}
	fmt.Printf("OptionDataByDelta: option after rounding: %s\n", opt.DxLinkString())
	delta, err := c.getOptDelta(opt.DxLinkString())
	if err != nil {
		return nil, fmt.Errorf("OptionDataByDelta: unable to find INITIAL option in subscription data: %s, %w", opt.DxLinkString(), err)
	}

	dist := math.Abs(delta - targetDelta)
	fmt.Printf("OptionDataByDelta: Initial values, option: %s, delta: %.6f, distance: %.6f\n", opt.DxLinkString(), delta, dist)
	// start a loop and increment to closest delta or delta = 0
	if delta > targetDelta {
		for s += float64(round); s < s*1.1; s += float64(round) {
			opt.IncrementStrike(float64(round))
			delta, err := c.getOptDelta(opt.DxLinkString())
			if err != nil {
				return nil, fmt.Errorf("OptionDataByDelta: unable to find option in subscription data: %s, %w", opt.DxLinkString(), err)
			}
			if math.Abs(delta-targetDelta) > dist {
				opt.DecrementStrike(float64(round))
				optDataPtr, err := c.GetOptData(opt.DxLinkString())
				if err != nil {
					return nil, err
				}
				fmt.Printf("Opt Data Ptr: %+v\n", *optDataPtr)
				return optDataPtr, nil
			}
			dist = math.Abs(delta - targetDelta)

			fmt.Printf("OptionDataByDelta: option: %s, delta: %.6f, distance: %.6f\n", opt.DxLinkString(), delta, dist)
		}
		return nil, fmt.Errorf("no option found incrementing")
	} else {
		for s -= float64(round); s > s*0.9; s -= float64(round) {
			opt.DecrementStrike(float64(round))
			delta, err := c.getOptDelta(opt.DxLinkString())
			if err != nil {
				return nil, fmt.Errorf("OptionDataByDelta: unable to find option in subscription data: %s, %w", opt.DxLinkString(), err)
			}
			if math.Abs(delta-targetDelta) > dist {
				opt.IncrementStrike(float64(round))
				optDataPtr, err := c.GetOptData(opt.DxLinkString())
				if err != nil {
					return nil, err
				}
				fmt.Printf("Opt Data Ptr: %+v\n", optDataPtr)
				return optDataPtr, nil
			}
			dist = math.Abs(delta - targetDelta)

			fmt.Printf("OptionDataByDelta:  option: %s, delta: %.6f, distance: %.6f\n", opt.DxLinkString(), delta, dist)
		}
		return nil, fmt.Errorf("no option found decrementing")
	}
}

func (c *DxLinkClient) getUnderlyingData(sym string) (*UnderlyingData, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	delay := c.delay
	for i := 0; i < c.retries; i++ {
		if underlyingPtr, ok := c.underlyingSubs[sym]; ok {
			if underlyingPtr != nil {
				return underlyingPtr, nil
			}
		}
		fmt.Printf("Retrying getUnderlyingData, attempt %d, delay: %s\n", i+1, delay.String())
		time.Sleep(delay)
		if c.expBackoff {
			delay *= 2
		}
	}
	return nil, fmt.Errorf("unable to find underlying in subscription data or is nil: %s", sym)
}
func (c *DxLinkClient) getUnderlyingPrice(sym string) (float64, error) {
	data, err := c.getUnderlyingData(sym)
	if err != nil {
		return 0.0, err
	}
	if data.Trade.Price == nil {
		return 0, fmt.Errorf("underlyingData.Trade.Price is a nil ptr")
	}
	price := *data.Trade.Price
	return price, nil

}

func (c *DxLinkClient) GetOptData(opt string) (*OptionData, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	delay := c.delay
	retry := func(i int, delay time.Duration) time.Duration {
		fmt.Printf("Retrying getOptData, attempt %d, delay: %s\n", i+1, delay.String())
		time.Sleep(delay)
		if c.expBackoff {
			delay *= 2
		}
		return delay
	}
	for i := 0; i < c.retries; i++ {
		if optionDataPtr, ok := c.optionSubs[opt]; !ok {
			delay = retry(i, delay)
		} else if optionDataPtr == nil {
			delay = retry(i, delay)
		} else if optionDataPtr.Greek.Delta == nil ||
			optionDataPtr.Quote.AskPrice == nil ||
			*optionDataPtr.Quote.AskPrice == 0.0 ||
			optionDataPtr.Quote.BidPrice == nil ||
			*optionDataPtr.Quote.BidPrice == 0.0 {
			delay = retry(i, delay)
		} else {
			return optionDataPtr, nil
		}
	}
	return nil, fmt.Errorf("unable to find option in subscription data or is nil: %s", opt)
}

func (c *DxLinkClient) getOptDelta(opt string) (float64, error) {
	optionDataPtr, err := c.GetOptData(opt)
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
