package util

import (
	"fmt"
	"time"

	"SIP/internal/model"
)

func ValidateCreateSIP(mode model.SIPMode, startAt time.Time, amountPaise int64, stepUpEnabled bool, stepUpBps int32) error {
	if err := mode.Validate(); err != nil {
		return fmt.Errorf("%w: %s", ErrValidation, err.Error())
	}
	if startAt.IsZero() {
		return fmt.Errorf("%w: startAt is required", ErrValidation)
	}
	if amountPaise <= 0 {
		return fmt.Errorf("%w: baseAmountPaise must be > 0", ErrValidation)
	}
	if stepUpEnabled && (stepUpBps < 0 || stepUpBps > 10000) {
		return fmt.Errorf("%w: stepUpBps should be in [0,10000]", ErrValidation)
	}
	return nil
}
