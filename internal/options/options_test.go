package options

import (
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
)

func TestOCCSymbolParse(t *testing.T) {
	occ := "AAPL  230818C00197500"
	option, err := ParseOCCOption(occ)
	assert.Equal(t, err, nil)

	assert.Equal(t, option.Underlying, "AAPL")
	expected_date := time.Date(2023, 8, 18, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, option.Date, expected_date)
	assert.Equal(t, option.OptionType, CallOption)
	assert.Equal(t, option.Strike, 197.5)
}

func TestOCCSymbolParseBuild(t *testing.T) {
	occ := "AAPL  230818C00197500"
	option, err := ParseOCCOption(occ)
	assert.Equal(t, err, nil)

	occAfter := option.OCCString()
	assert.Equal(t, occ, occAfter)
}
