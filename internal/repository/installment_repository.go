package repository

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"SIP/internal/model"
	"SIP/internal/util"
)

type InstallmentRepository interface {
	Create(ctx context.Context, inst model.Installment) (model.Installment, error)
	Update(ctx context.Context, inst model.Installment) (model.Installment, error)
	ListBySIP(ctx context.Context, sipID string) ([]model.Installment, error)
	GetByPaymentRequestID(ctx context.Context, paymentRequestID string) (model.Installment, error)
	CountBySIP(ctx context.Context, sipID string) (int64, error)
}

type InMemoryInstallmentRepository struct {
	mu sync.RWMutex

	installments map[string]model.Installment
	bySIP        map[string][]string
	byPaymentReq map[string]string
}

var (
	installmentRepositoryOnce sync.Once
	installmentRepositoryInst *InMemoryInstallmentRepository
)

func NewInMemoryInstallmentRepository() *InMemoryInstallmentRepository {
	installmentRepositoryOnce.Do(func() {
		installmentRepositoryInst = &InMemoryInstallmentRepository{
			installments: map[string]model.Installment{},
			bySIP:        map[string][]string{},
			byPaymentReq: map[string]string{},
		}
	})
	return installmentRepositoryInst
}

func (r *InMemoryInstallmentRepository) Create(_ context.Context, inst model.Installment) (model.Installment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.installments[inst.InstallmentID]; ok {
		return model.Installment{}, fmt.Errorf("%w: installment %s already exists", util.ErrConflict, inst.InstallmentID)
	}
	r.installments[inst.InstallmentID] = inst
	r.bySIP[inst.SIPID] = append(r.bySIP[inst.SIPID], inst.InstallmentID)
	if inst.PaymentRequestID != "" {
		r.byPaymentReq[inst.PaymentRequestID] = inst.InstallmentID
	}
	return inst, nil
}

func (r *InMemoryInstallmentRepository) Update(_ context.Context, inst model.Installment) (model.Installment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.installments[inst.InstallmentID]; !ok {
		return model.Installment{}, fmt.Errorf("%w: installment %s", util.ErrNotFound, inst.InstallmentID)
	}
	r.installments[inst.InstallmentID] = inst
	if inst.PaymentRequestID != "" {
		r.byPaymentReq[inst.PaymentRequestID] = inst.InstallmentID
	}
	return inst, nil
}

func (r *InMemoryInstallmentRepository) ListBySIP(_ context.Context, sipID string) ([]model.Installment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := r.bySIP[sipID]
	out := make([]model.Installment, 0, len(ids))
	for _, id := range ids {
		out = append(out, r.installments[id])
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SequenceNo < out[j].SequenceNo })
	return out, nil
}

func (r *InMemoryInstallmentRepository) GetByPaymentRequestID(_ context.Context, paymentRequestID string) (model.Installment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byPaymentReq[paymentRequestID]
	if !ok {
		return model.Installment{}, fmt.Errorf("%w: payment request %s", util.ErrNotFound, paymentRequestID)
	}
	inst, ok := r.installments[id]
	if !ok {
		return model.Installment{}, fmt.Errorf("%w: installment %s", util.ErrNotFound, id)
	}
	return inst, nil
}

func (r *InMemoryInstallmentRepository) CountBySIP(_ context.Context, sipID string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return int64(len(r.bySIP[sipID])), nil
}
