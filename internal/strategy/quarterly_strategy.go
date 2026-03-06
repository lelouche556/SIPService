package strategy

import (
	"time"

	"SIP/internal/model"
	"SIP/internal/util"
)

type QuarterlyStrategy struct{}

func (QuarterlyStrategy) NextRun(prev time.Time, sip model.SIP, loc *time.Location) time.Time {
	p := prev.In(loc)
	year, month, _ := p.Date()
	h, m, s := p.Clock()
	ns := p.Nanosecond()

	targetMonth := int(month) + 3
	for targetMonth > 12 {
		targetMonth -= 12
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
