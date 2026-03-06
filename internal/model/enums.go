package model

import "fmt"

type SIPMode string

const (
	SIPModeWeekly    SIPMode = "WEEKLY"
	SIPModeMonthly   SIPMode = "MONTHLY"
	SIPModeQuarterly SIPMode = "QUARTERLY"
)

func (m SIPMode) Validate() error {
	switch m {
	case SIPModeWeekly, SIPModeMonthly, SIPModeQuarterly:
		return nil
	default:
		return fmt.Errorf("invalid SIP mode: %s", m)
	}
}

type SIPStatus string

const (
	SIPStatusActive  SIPStatus = "ACTIVE"
	SIPStatusPaused  SIPStatus = "PAUSED"
	SIPStatusStopped SIPStatus = "STOPPED"
)

type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "PENDING"
	PaymentStatusSuccess PaymentStatus = "SUCCESS"
	PaymentStatusFailed  PaymentStatus = "FAILED"
)

func (s PaymentStatus) IsTerminal() bool {
	return s == PaymentStatusSuccess || s == PaymentStatusFailed
}
