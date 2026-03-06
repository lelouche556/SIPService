package model

import (
	"fmt"
	"time"
)

type SIP struct {
	SIPID  string `json:"sipId"`
	UserID string `json:"userId"`
	FundID string `json:"fundId"`

	Mode      SIPMode   `json:"mode"`
	StartAt   time.Time `json:"startAt"`
	NextRunAt time.Time `json:"nextRunAt"`

	AnchorWeekday    time.Weekday `json:"anchorWeekday"`
	AnchorDayOfMonth int          `json:"anchorDayOfMonth"`
	AnchorHour       int          `json:"anchorHour"`
	AnchorMinute     int          `json:"anchorMinute"`
	AnchorSecond     int          `json:"anchorSecond"`
	AnchorNanosecond int          `json:"anchorNanosecond"`

	BaseAmountPaise int64 `json:"baseAmountPaise"`
	StepUpEnabled   bool  `json:"stepUpEnabled"`
	StepUpBps       int32 `json:"stepUpBps"`

	Status SIPStatus `json:"status"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Version   int64     `json:"version"`
}

func (s SIP) Validate() error {
	if s.SIPID == "" || s.UserID == "" || s.FundID == "" {
		return fmt.Errorf("sipId, userId and fundId are required")
	}
	if err := s.Mode.Validate(); err != nil {
		return err
	}
	if s.BaseAmountPaise <= 0 {
		return fmt.Errorf("base amount must be > 0")
	}
	if s.StartAt.IsZero() || s.NextRunAt.IsZero() {
		return fmt.Errorf("startAt and nextRunAt are required")
	}
	if s.StepUpEnabled && (s.StepUpBps < 0 || s.StepUpBps > 10000) {
		return fmt.Errorf("stepUpBps should be in [0,10000]")
	}
	return nil
}

func (s *SIP) Pause(now time.Time) error {
	if s.Status == SIPStatusStopped {
		return fmt.Errorf("stopped sip cannot be paused")
	}
	s.Status = SIPStatusPaused
	s.UpdatedAt = now
	return nil
}

func (s *SIP) Unpause(now time.Time) error {
	if s.Status == SIPStatusStopped {
		return fmt.Errorf("stopped sip cannot be unpaused")
	}
	s.Status = SIPStatusActive
	s.UpdatedAt = now
	return nil
}

func (s *SIP) Stop(now time.Time) {
	s.Status = SIPStatusStopped
	s.UpdatedAt = now
}
