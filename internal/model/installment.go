package model

import "time"

type Installment struct {
	InstallmentID string `json:"installmentId"`
	SIPID         string `json:"sipId"`
	SequenceNo    int64  `json:"sequenceNo"`

	ScheduledAt time.Time `json:"scheduledAt"`
	ExecutedAt  time.Time `json:"executedAt"`

	AmountPaise int64 `json:"amountPaise"`
	NAVMic      int64 `json:"navMic"`
	UnitsMic    int64 `json:"unitsMic"`

	PaymentRequestID string        `json:"paymentRequestId"`
	PaymentStatus    PaymentStatus `json:"paymentStatus"`
	FailureReason    string        `json:"failureReason,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
