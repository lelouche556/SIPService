package util

import "time"

func LastDayOfMonth(year int, month time.Month, loc *time.Location) int {
	firstOfNext := time.Date(year, month+1, 1, 0, 0, 0, 0, loc)
	last := firstOfNext.Add(-24 * time.Hour)
	return last.Day()
}
