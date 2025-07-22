package dxlink

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
	// increment up in strike from ATM
	if (string(leg.OptType) == "P" && leg.StrikeMethVal > 50) ||
		(string(leg.OptType) == "C" && leg.StrikeMethVal < 50) {
		for s := math.Floor(currentPrice); s < currentPrice
		atmOption := options.OptionSymbol{
			Underlying: underlying,
			Date:       exp,
			Strike:     currentPrice,
			OptionType: options.OptionType(leg.OptType),
		}.DxLinkString()
	}
	if (string(leg.OptType) == "P" && leg.StrikeMethVal < 50) ||
		(string(leg.OptType) == "C" && leg.StrikeMethVal > 50) {
		incr = incr * -1
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	data, ok := c.optionSubs[atmOption]
	if !ok {
		return "", fmt.Errorf("unable to find ATM option in subscription data: %s", atmOption)
	}
}
