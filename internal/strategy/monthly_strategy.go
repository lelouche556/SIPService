package strategy

import (
	"time"

	"SIP/internal/model"
	"SIP/internal/util"
)

type MonthlyStrategy struct{}

func (MonthlyStrategy) NextRun(prev time.Time, sip model.SIP, loc *time.Location) time.Time {
	p := prev.In(loc)
	year, month, _ := p.Date()
	h, m, s := p.Clock()
	ns := p.Nanosecond()

	targetMonth := int(month) + 1
	if targetMonth > 12 {
		targetMonth = 1
		year++
	}
	day := sip.AnchorDayOfMonth
	if day < 1 {
		day = 1
	}
	lastDay := util.LastDayOfMonth(year, time.Month(targetMonth), loc)
	if day > lastDay {
		day = lastDay
	}
	return time.Date(year, time.Month(targetMonth), day, h, m, s, ns, loc)
}
