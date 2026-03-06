package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"SIP/internal/model"
	"SIP/internal/repository"
	"SIP/internal/util"
)

type FundService struct {
	fundRepo repository.FundRepository
	pricing  *PricingService
	idGen    *util.IDGenerator
	now      func() time.Time
}

var (
	fundServiceOnce sync.Once
	fundServiceInst *FundService
)

func NewFundService(fundRepo repository.FundRepository, pricing *PricingService, idGen *util.IDGenerator, now func() time.Time) *FundService {
	fundServiceOnce.Do(func() {
		fundServiceInst = &FundService{fundRepo: fundRepo, pricing: pricing, idGen: idGen, now: now}
	})
	return fundServiceInst
}

func (s *FundService) BrowseFunds(ctx context.Context, filters repository.FundFilters) ([]model.Fund, error) {
	if filters.Query == "" && filters.Category == "" && filters.AMC == "" && filters.RiskTag == "" {
		return s.fundRepo.List(ctx)
	}
	return s.fundRepo.Search(ctx, filters)
}

type FundWithPrice struct {
	Fund model.Fund        `json:"fund"`
	NAV  model.MarketPrice `json:"nav"`
}

func (s *FundService) ListFundsWithPrice(ctx context.Context, filters repository.FundFilters) ([]FundWithPrice, error) {
	funds, err := s.BrowseFunds(ctx, filters)
	if err != nil {
		return nil, err
	}
	out := make([]FundWithPrice, 0, len(funds))
	for _, f := range funds {
		price, err := s.pricing.GetLatestNAV(ctx, f.FundID)
		if err != nil {
			return nil, err
		}
		out = append(out, FundWithPrice{Fund: f, NAV: price})
	}
	return out, nil
}

func (s *FundService) CreateFund(ctx context.Context, fundID, name, amc, category, riskTag string, isActive *bool, navMic int64) (model.Fund, error) {
	if name == "" || amc == "" || category == "" || riskTag == "" {
		return model.Fund{}, fmt.Errorf("%w: name, amc, category, riskTag are required", util.ErrValidation)
	}
	if navMic <= 0 {
		return model.Fund{}, fmt.Errorf("%w: navMic must be > 0", util.ErrValidation)
	}
	if fundID == "" {
		fundID = s.idGen.Next("fund")
	}
	active := true
	if isActive != nil {
		active = *isActive
	}
	fund := model.Fund{
		FundID:   fundID,
		Name:     name,
		AMC:      amc,
		Category: category,
		RiskTag:  riskTag,
		IsActive: active,
	}
	created, err := s.fundRepo.Create(ctx, fund)
	if err != nil {
		return model.Fund{}, err
	}
	s.pricing.UpdateNAV(created.FundID, navMic, s.now())
	return created, nil
}
