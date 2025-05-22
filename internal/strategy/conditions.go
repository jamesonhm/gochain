package strategy

import (
	"log/slog"
	"time"

	"github.com/jamesonhm/gochain/internal/dxlink"
	"github.com/jamesonhm/gochain/internal/tasty"
)

type EntryCondition func(marketData *dxlink.DxLinkClient, accountData *tasty.TastyAPI) bool
type ExitCondition func(marketData *dxlink.DxLinkClient, accountData *tasty.TastyAPI) bool
type StrikeCalc func(marketData *dxlink.DxLinkClient) float64

func EntryDayOfWeeek(allowedDays []time.Weekday) EntryCondition {
	return func(_ *dxlink.DxLinkClient, _ *tasty.TastyAPI) bool {
		today := time.Now().Weekday()
		for _, day := range allowedDays {
			if day == today {
				return true
			}
		}
		return false
	}
}

func VixONMoveMin(threshold float64) EntryCondition {
	return func(marketData *dxlink.DxLinkClient, _ *tasty.TastyAPI) bool {
		vixMove, err := marketData.VixONMove()
		if err != nil {
			slog.Error("VixONMoveMin Entry Condition", "error", err)
			return false
		}
		return vixMove >= threshold
	}
}

func VixONMoveMax(threshold float64) EntryCondition {
	return func(marketData *dxlink.DxLinkClient, _ *tasty.TastyAPI) bool {
		vixMove, err := marketData.VixONMove()
		if err != nil {
			slog.Error("VixONMoveMax Entry Condition", "error", err)
			return false
		}
		return vixMove <= threshold
	}
}
