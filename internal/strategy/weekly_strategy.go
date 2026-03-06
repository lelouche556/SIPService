package strategy

import (
	"time"

	"SIP/internal/model"
)

type WeeklyStrategy struct{}

func (WeeklyStrategy) NextRun(prev time.Time, _ model.SIP, loc *time.Location) time.Time {
	return prev.In(loc).AddDate(0, 0, 7)
}
