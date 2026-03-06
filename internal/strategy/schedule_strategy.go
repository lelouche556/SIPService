package strategy

import (
	"fmt"
	"time"

	"SIP/internal/model"
)

type ScheduleStrategy interface {
	NextRun(prev time.Time, sip model.SIP, loc *time.Location) time.Time
}

func ForMode(mode model.SIPMode) (ScheduleStrategy, error) {
	switch mode {
	case model.SIPModeWeekly:
		return WeeklyStrategy{}, nil
	case model.SIPModeMonthly:
		return MonthlyStrategy{}, nil
	case model.SIPModeQuarterly:
		return QuarterlyStrategy{}, nil
	default:
		return nil, fmt.Errorf("unsupported mode: %s", mode)
	}
}
