package dt

import (
	"math"
	"time"
)

func Midnight(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}

func EndOfDay(date time.Time) (*time.Time, error) {
	nytz, err := time.LoadLocation("America/New_York")
	if err != nil {
		return nil, err
	}
	end := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, nytz)
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
