package dxlink

import (
	"fmt"
	"strconv"
	"time"
	"unicode"
)

type OptionSymbol struct {
	Underlying string
	Date       time.Time
	Type       OptionType
	Strike     float64
}

func (o OptionSymbol) String() string {
	return fmt.Sprintf(".%s%s%s%.0f", o.Underlying, o.Date.Format("060102"), o.Type, o.Strike)
}

func ParseOption(option string) (*OptionSymbol, error) {
	var split int
	if string(option[0]) != "." {
		return nil, fmt.Errorf("unrecognized option format, no '.' prefix: %s", option)
	}
	for i, r := range option {
		if string(r) == "." || unicode.IsLetter(r) {
			continue
		}
		if unicode.IsDigit(r) {
			split = i
			break
		}
		return nil, fmt.Errorf("unrecognized option format, no number found: %s", option)
	}
	opt_date := option[split : split+6]
	date, err := time.Parse("060102", opt_date)
	if err != nil {
		return nil, fmt.Errorf("unable to parse date: %s, err: %w", opt_date, err)
	}
	opt_type := OptionType(option[split+6 : split+7])
	strike, err := strconv.ParseFloat(option[split+7:], 64)
	if err != nil {
		return nil, fmt.Errorf("unable to parse strike as float: %s, err: %w", option[split+7:], err)
	}

	res := OptionSymbol{
		Underlying: option[1:split],
		Date:       date,
		Type:       opt_type,
		Strike:     strike,
	}
	return &res, nil
}
