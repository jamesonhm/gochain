package dt

import (
	"math"
	"time"
)

func TZNY() *time.Location {
	nytz, _ := time.LoadLocation("America/New_York")
	return nytz
}

func Midnight(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, TZNY())
}

func EndOfDay(date time.Time) (*time.Time, error) {
	end := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, TZNY())
	return &end, nil
}

func WeekdaysBetween(start, end time.Time) int {
	offset := int(end.Weekday()) - int(start.Weekday())
	if end.Weekday() == time.Sunday {
		offset++
	}
	start = start.AddDate(0, 0, -int(start.Weekday()))
	end = end.AddDate(0, 0, -int(end.Weekday()))
	diff := end.Sub(start).Truncate(time.Hour * 24)
	weeks := float64((diff.Hours() / 24) / 7)
	return int(math.Round(weeks)*5) + offset
}

func PreviousWeekday(d time.Time) time.Time {
	if d.Weekday() == 1 {
		return d.AddDate(0, 0, -3)
	} else if d.Weekday() == 0 {
		return d.AddDate(0, 0, -2)
	}
	return d.AddDate(0, 0, -1)
}

func NextWeekday(d time.Time) time.Time {
	if d.Weekday() == 5 {
		return d.AddDate(0, 0, 3)
	} else if d.Weekday() == 6 {
		return d.AddDate(0, 0, 2)
	}
	return d.AddDate(0, 0, 1)
}

func DTEToDate(dte int) time.Time {
	exp := time.Now().In(TZNY()).AddDate(0, 0, dte)
	if exp.Weekday() < 1 || exp.Weekday() > 5 {
		exp = NextWeekday(exp)
	}
	return Midnight(exp)
}

func DTEToDateHolidays(start time.Time, dte int, holidays []time.Time) time.Time {
	exp := start.In(TZNY()).AddDate(0, 0, dte)
	for {
		if exp.Weekday() == 0 || exp.Weekday() == 6 || inls(exp, holidays) {
			exp = NextWeekday(exp)
		} else {
			return Midnight(exp)
		}
	}
}

func inls(v time.Time, ls []time.Time) bool {
	for _, h := range ls {
		if YMDEqual(v, h) {
			return true
		}
	}
	return false
}

func YMDEqual(d1 time.Time, d2 time.Time) bool {
	return d1.Year() == d2.Year() && d1.Month() == d2.Month() && d1.Day() == d2.Day()
}
