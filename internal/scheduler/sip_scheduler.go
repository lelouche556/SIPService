package scheduler

import (
	"context"
	"log"
	"sync"
	"time"
)

type Runner interface {
	RunDueExecutions(ctx context.Context, now time.Time) error
}

type SIPScheduler struct {
	runner   Runner
	interval time.Duration
	now      func() time.Time
}

var (
	sipSchedulerOnce sync.Once
	sipSchedulerInst *SIPScheduler
)

func NewSIPScheduler(runner Runner, interval time.Duration, now func() time.Time) *SIPScheduler {
	sipSchedulerOnce.Do(func() {
		sipSchedulerInst = &SIPScheduler{runner: runner, interval: interval, now: now}
	})
	return sipSchedulerInst
}

func (s *SIPScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.runner.RunDueExecutions(ctx, s.now()); err != nil {
				log.Printf("scheduler run failed: %v", err)
			}
		}
	}
}
