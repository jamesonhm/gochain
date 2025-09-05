package tasty

import (
	"context"
	"net/http"
	"strings"
	"time"
)

const (
	MarketSessionHolidaysPath = "/market-time/equities/holidays"
)

func (c *TastyAPI) GetMarketHolidays(ctx context.Context) ([]HolidayDate, error) {
	res := &MarketHolidayResponse{}
	path := c.baseurl + MarketSessionHolidaysPath
	err := c.request(ctx, http.MethodGet, auth, path, nil, nil, res)
	if err != nil {
		return nil, err
	}
	holidays := res.Data.MarketHolidays
	half_days := res.Data.MarketHalfDays
	holidays = append(holidays, half_days...)
	return holidays, nil
}

func (c *TastyAPI) GetMarketHolidaysDT(ctx context.Context) ([]time.Time, error) {
	holidays, err := c.GetMarketHolidays(ctx)
	if err != nil {
		return nil, err
	}
	var ts []time.Time
	for _, h := range holidays {
		ts = append(ts, time.Time(h))
	}
	return ts, nil
}

type MarketHolidayResponse struct {
	Data struct {
		MarketHalfDays []HolidayDate `json:"market-half-days"`
		MarketHolidays []HolidayDate `json:"market-holidays"`
	} `json:"data"`
}

type HolidayDate time.Time

func (hd *HolidayDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	layout := "2006-01-02"
	loc, _ := time.LoadLocation("America/New_York")
	t, err := time.ParseInLocation(layout, s, loc)
	if err != nil {
		return err
	}
	*hd = HolidayDate(t)
	return nil
}

func (hd HolidayDate) String() string {
	return time.Time(hd).Format(time.DateOnly)
}
