package service

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"SIP/internal/model"
	"SIP/internal/repository"
	"SIP/internal/strategy"
	"SIP/internal/util"
)

type ExecutionService struct {
	sipRepo         repository.SIPRepository
	installmentRepo repository.InstallmentRepository
	pricing         *PricingService
	payment         *PaymentService
	idGen           *util.IDGenerator
	loc             *time.Location
	now             func() time.Time
	mu              sync.Mutex
	inProgress      map[string]struct{}
}

var (
	executionServiceOnce sync.Once
	executionServiceInst *ExecutionService
)

func NewExecutionService(
	sipRepo repository.SIPRepository,
	installmentRepo repository.InstallmentRepository,
	pricingService *PricingService,
	paymentService *PaymentService,
	idGen *util.IDGenerator,
	loc *time.Location,
	now func() time.Time,
) *ExecutionService {
	executionServiceOnce.Do(func() {
		executionServiceInst = &ExecutionService{
			sipRepo:         sipRepo,
			installmentRepo: installmentRepo,
			pricing:         pricingService,
			payment:         paymentService,
			idGen:           idGen,
			loc:             loc,
			now:             now,
			inProgress:      map[string]struct{}{},
		}
	})
	return executionServiceInst
}

func (s *ExecutionService) RunDueExecutions(ctx context.Context, now time.Time) error {
	due, err := s.sipRepo.ListDue(ctx, now.In(s.loc))
	if err != nil {
		return err
	}
	for _, sip := range due {
		if _, err := s.ExecuteSIP(ctx, sip.SIPID, now); err != nil {
			if errors.Is(err, util.ErrExecutionInProgress) {
				continue
			}
		}
	}
	return nil
}

func (s *ExecutionService) ExecuteSIP(ctx context.Context, sipID string, now time.Time) (model.Installment, error) {
	sip, err := s.sipRepo.GetByID(ctx, sipID)
	if err != nil {
		return model.Installment{}, err
	}
	if sip.Status != model.SIPStatusActive {
		return model.Installment{}, fmt.Errorf("%w: sip %s is not active", util.ErrInvalidState, sipID)
	}
	if sip.NextRunAt.After(now.In(s.loc)) {
		return model.Installment{}, fmt.Errorf("%w: sip %s is not due", util.ErrValidation, sipID)
	}

	if !s.tryStartExecution(sipID) {
		return model.Installment{}, util.ErrExecutionInProgress
	}
	defer s.finishExecution(sipID)

	count, err := s.installmentRepo.CountBySIP(ctx, sipID)
	if err != nil {
		return model.Installment{}, err
	}
	sequence := count + 1
	amount := sip.BaseAmountPaise
	if sip.StepUpEnabled {
		amount = computeStepUpAmount(sip.BaseAmountPaise, sip.StepUpBps, sequence)
	}

	price, err := s.pricing.GetLatestNAV(ctx, sip.FundID)
	if err != nil {
		return model.Installment{}, err
	}

	nowInLoc := now.In(s.loc)
	inst := model.Installment{
		InstallmentID:    s.idGen.Next("inst"),
		SIPID:            sip.SIPID,
		SequenceNo:       sequence,
		ScheduledAt:      sip.NextRunAt,
		AmountPaise:      amount,
		NAVMic:           price.NAVMic,
		PaymentRequestID: s.idGen.Next("payreq"),
		PaymentStatus:    model.PaymentStatusPending,
		CreatedAt:        nowInLoc,
		UpdatedAt:        nowInLoc,
	}
	inst, err = s.installmentRepo.Create(ctx, inst)
	if err != nil {
		return model.Installment{}, err
	}

	_, err = s.payment.InitiatePayment(ctx, PaymentRequest{
		PaymentRequestID: inst.PaymentRequestID,
		UserID:           sip.UserID,
		SIPID:            sip.SIPID,
		InstallmentID:    inst.InstallmentID,
		AmountPaise:      inst.AmountPaise,
		RequestedAt:      nowInLoc,
	})
	if err != nil {
		return model.Installment{}, err
	}

	strat, err := strategy.ForMode(sip.Mode)
	if err != nil {
		return model.Installment{}, err
	}
	sip.NextRunAt = strat.NextRun(sip.NextRunAt, sip, s.loc)
	sip.UpdatedAt = nowInLoc
	if _, err := s.sipRepo.Update(ctx, sip); err != nil {
		return model.Installment{}, err
	}
	return inst, nil
}

func (s *ExecutionService) HandlePaymentCallback(ctx context.Context, paymentRequestID string, status model.PaymentStatus, failureReason string) (model.Installment, error) {
	inst, err := s.installmentRepo.GetByPaymentRequestID(ctx, paymentRequestID)
	if err != nil {
		return model.Installment{}, err
	}
	if inst.PaymentStatus.IsTerminal() {
		return inst, nil
	}

	inst.UpdatedAt = s.now().In(s.loc)
	switch status {
	case model.PaymentStatusSuccess:
		inst.PaymentStatus = model.PaymentStatusSuccess
		inst.ExecutedAt = s.now().In(s.loc)
		inst.UnitsMic = unitsMic(inst.AmountPaise, inst.NAVMic)
		inst.FailureReason = ""
	case model.PaymentStatusFailed:
		inst.PaymentStatus = model.PaymentStatusFailed
		inst.FailureReason = failureReason
	default:
		return model.Installment{}, fmt.Errorf("%w: unsupported payment status %s", util.ErrValidation, status)
	}

	return s.installmentRepo.Update(ctx, inst)
}

func (s *ExecutionService) CatchUp(ctx context.Context, sipID, userID string, numInstallments int64, now time.Time) (model.Installment, error) {
	if numInstallments <= 0 {
		return model.Installment{}, fmt.Errorf("%w: numInstallments must be > 0", util.ErrValidation)
	}
	sip, err := s.sipRepo.GetByID(ctx, sipID)
	if err != nil {
		return model.Installment{}, err
	}
	if sip.UserID != userID {
		return model.Installment{}, ErrUnauthorized
	}
	if sip.Status == model.SIPStatusStopped {
		return model.Installment{}, fmt.Errorf("%w: stopped sip cannot catch up", util.ErrInvalidState)
	}

	count, err := s.installmentRepo.CountBySIP(ctx, sipID)
	if err != nil {
		return model.Installment{}, err
	}
	total := int64(0)
	for i := int64(1); i <= numInstallments; i++ {
		seq := count + i
		amount := sip.BaseAmountPaise
		if sip.StepUpEnabled {
			amount = computeStepUpAmount(sip.BaseAmountPaise, sip.StepUpBps, seq)
		}
		total += amount
	}

	price, err := s.pricing.GetLatestNAV(ctx, sip.FundID)
	if err != nil {
		return model.Installment{}, err
	}
	nowInLoc := now.In(s.loc)
	inst := model.Installment{
		InstallmentID:    s.idGen.Next("inst"),
		SIPID:            sip.SIPID,
		SequenceNo:       count + 1,
		ScheduledAt:      nowInLoc,
		AmountPaise:      total,
		NAVMic:           price.NAVMic,
		PaymentRequestID: s.idGen.Next("payreq"),
		PaymentStatus:    model.PaymentStatusPending,
		FailureReason:    fmt.Sprintf("catch-up for %d installments", numInstallments),
		CreatedAt:        nowInLoc,
		UpdatedAt:        nowInLoc,
	}
	inst, err = s.installmentRepo.Create(ctx, inst)
	if err != nil {
		return model.Installment{}, err
	}

	if _, err := s.payment.InitiatePayment(ctx, PaymentRequest{
		PaymentRequestID: inst.PaymentRequestID,
		UserID:           sip.UserID,
		SIPID:            sip.SIPID,
		InstallmentID:    inst.InstallmentID,
		AmountPaise:      inst.AmountPaise,
		RequestedAt:      nowInLoc,
	}); err != nil {
		return model.Installment{}, err
	}

	if sip.Status == model.SIPStatusPaused {
		_ = sip.Unpause(nowInLoc)
		_, _ = s.sipRepo.Update(ctx, sip)
	}
	return inst, nil
}

func (s *ExecutionService) tryStartExecution(sipID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.inProgress[sipID]; exists {
		return false
	}
	s.inProgress[sipID] = struct{}{}
	return true
}

func (s *ExecutionService) finishExecution(sipID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.inProgress, sipID)
}

func computeStepUpAmount(basePaise int64, stepUpBps int32, sequence int64) int64 {
	if sequence <= 1 || stepUpBps == 0 {
		return basePaise
	}
	result := big.NewInt(basePaise)
	mul := big.NewInt(int64(10000 + stepUpBps))
	div := big.NewInt(10000)
	for i := int64(1); i < sequence; i++ {
		result.Mul(result, mul)
		result.Add(result, big.NewInt(5000))
		result.Div(result, div)
	}
	if !result.IsInt64() {
		return basePaise
	}
	return result.Int64()
}

func unitsMic(amountPaise, navMic int64) int64 {
	if amountPaise <= 0 || navMic <= 0 {
		return 0
	}
	n := big.NewInt(amountPaise)
	n.Mul(n, big.NewInt(10000000000))
	n.Add(n, big.NewInt(navMic*50))
	d := big.NewInt(navMic * 100)
	n.Div(n, d)
	if !n.IsInt64() {
		return 0
	}
	return n.Int64()
}
