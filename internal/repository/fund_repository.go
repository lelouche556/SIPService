package repository

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"SIP/internal/model"
	"SIP/internal/util"
)

type FundFilters struct {
	Query    string
	Category string
	AMC      string
	RiskTag  string
}

type FundRepository interface {
	Create(ctx context.Context, fund model.Fund) (model.Fund, error)
	List(ctx context.Context) ([]model.Fund, error)
	Search(ctx context.Context, filters FundFilters) ([]model.Fund, error)
	GetByID(ctx context.Context, fundID string) (model.Fund, error)
}

type InMemoryFundRepository struct {
	mu    sync.RWMutex
	funds map[string]model.Fund
}

var (
	fundRepositoryOnce sync.Once
	fundRepositoryInst *InMemoryFundRepository
)

func NewInMemoryFundRepository(seed []model.Fund) *InMemoryFundRepository {
	fundRepositoryOnce.Do(func() {
		m := make(map[string]model.Fund, len(seed))
		for _, f := range seed {
			m[f.FundID] = f
		}
		fundRepositoryInst = &InMemoryFundRepository{funds: m}
	})
	return fundRepositoryInst
}

func (r *InMemoryFundRepository) List(_ context.Context) ([]model.Fund, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]model.Fund, 0, len(r.funds))
	for _, f := range r.funds {
		out = append(out, f)
	}
	return out, nil
}

func (r *InMemoryFundRepository) Create(_ context.Context, fund model.Fund) (model.Fund, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.funds[fund.FundID]; ok {
		return model.Fund{}, fmt.Errorf("%w: fund %s already exists", util.ErrConflict, fund.FundID)
	}
	r.funds[fund.FundID] = fund
	return fund, nil
}

func (r *InMemoryFundRepository) Search(_ context.Context, filters FundFilters) ([]model.Fund, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	q := strings.ToLower(strings.TrimSpace(filters.Query))
	out := make([]model.Fund, 0)
	for _, f := range r.funds {
		if q != "" && !strings.Contains(strings.ToLower(f.Name), q) {
			continue
		}
		if filters.Category != "" && !strings.EqualFold(filters.Category, f.Category) {
			continue
		}
		if filters.AMC != "" && !strings.EqualFold(filters.AMC, f.AMC) {
			continue
		}
		if filters.RiskTag != "" && !strings.EqualFold(filters.RiskTag, f.RiskTag) {
			continue
		}
		out = append(out, f)
	}
	return out, nil
}

func (r *InMemoryFundRepository) GetByID(_ context.Context, fundID string) (model.Fund, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.funds[fundID]
	if !ok {
		return model.Fund{}, fmt.Errorf("%w: fund %s", util.ErrNotFound, fundID)
	}
	return f, nil
}
