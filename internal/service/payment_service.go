package service

import (
	"context"
	"sync"
	"time"
)

type PaymentRequest struct {
	PaymentRequestID string
	UserID           string
	SIPID            string
	InstallmentID    string
	AmountPaise      int64
	RequestedAt      time.Time
}

type PaymentAck struct {
	PaymentRequestID string    `json:"paymentRequestId"`
	AcceptedAt       time.Time `json:"acceptedAt"`
}

type PaymentService struct {
	mu   sync.Mutex
	seen map[string]PaymentRequest
	now  func() time.Time
}

var (
	paymentServiceOnce sync.Once
	paymentServiceInst *PaymentService
)

func NewPaymentService(now func() time.Time) *PaymentService {
	paymentServiceOnce.Do(func() {
		paymentServiceInst = &PaymentService{seen: map[string]PaymentRequest{}, now: now}
	})
	return paymentServiceInst
}

func (s *PaymentService) InitiatePayment(_ context.Context, req PaymentRequest) (PaymentAck, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.seen[req.PaymentRequestID]; !ok {
		s.seen[req.PaymentRequestID] = req
	}
	return PaymentAck{PaymentRequestID: req.PaymentRequestID, AcceptedAt: s.now()}, nil
}
