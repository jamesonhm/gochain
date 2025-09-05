package dt

import (
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
)

func holidays() []time.Time {
	var holidays = [...]string{"2022-01-17", "2022-02-21", "2022-04-15",
		"2022-05-30", "2022-06-20", "2022-07-04", "2022-09-05", "2022-11-24", "2022-12-26",
		"2023-01-02", "2023-01-16", "2023-02-20", "2023-04-07", "2023-05-29", "2023-06-19",
		"2023-07-04", "2023-09-04", "2023-11-23", "2023-12-25", "2024-01-01", "2024-01-15",
		"2024-02-19", "2024-03-29", "2024-05-27", "2024-06-19", "2024-07-04", "2024-09-02",
		"2024-11-28", "2024-12-25", "2025-01-01", "2025-01-09", "2025-01-20", "2025-02-17",
		"2025-04-18", "2025-05-26", "2025-06-19", "2025-07-04", "2025-09-01", "2025-11-27",
		"2025-12-25", "2026-01-01", "2026-01-19", "2026-02-16", "2026-04-03", "2026-05-25",
		"2026-06-19", "2026-07-03", "2026-09-07", "2026-11-26", "2026-12-25"}
	var holidaydt []time.Time
	for _, h := range holidays {
		dt, _ := time.Parse(time.DateOnly, h)
		holidaydt = append(holidaydt, dt)
	}
	return holidaydt
}

func TestDTEToDateHolidays(t *testing.T) {
	var starts = []time.Time{
		time.Date(2022, 1, 10, 8, 30, 40, 0, TZNY()),
		time.Date(2022, 2, 14, 9, 35, 10, 0, TZNY()),
		time.Date(2022, 2, 14, 9, 35, 10, 0, TZNY()),
		time.Date(2022, 4, 14, 9, 35, 10, 0, TZNY()),
	}
	var dtes = []int{
		5,
		7,
		1,
		1,
	}
	var expecteds = []time.Time{
		time.Date(2022, 1, 18, 0, 0, 0, 0, TZNY()),
		time.Date(2022, 2, 22, 0, 0, 0, 0, TZNY()),
		time.Date(2022, 2, 15, 0, 0, 0, 0, TZNY()),
		time.Date(2022, 4, 18, 0, 0, 0, 0, TZNY()),
	}
	for i, start := range starts {
		dt := DTEToDateHolidays(start, dtes[i], holidays())
		assert.Equal(t, dt, expecteds[i])
	}

}
