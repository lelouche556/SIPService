package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"SIP/internal/model"
	"SIP/internal/util"
)

type PricingService struct {
	mu     sync.RWMutex
	latest map[string]model.MarketPrice
}

var (
	pricingServiceOnce sync.Once
	pricingServiceInst *PricingService
)

func NewPricingService(seed []model.MarketPrice) *PricingService {
	pricingServiceOnce.Do(func() {
		m := make(map[string]model.MarketPrice, len(seed))
		for _, p := range seed {
			m[p.FundID] = p
		}
		pricingServiceInst = &PricingService{latest: m}
	})
	return pricingServiceInst
}

func (s *PricingService) GetLatestNAV(_ context.Context, fundID string) (model.MarketPrice, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.latest[fundID]
	if !ok {
		return model.MarketPrice{}, fmt.Errorf("%w: nav for fund %s", util.ErrNotFound, fundID)
	}
	return p, nil
}

func (s *PricingService) UpdateNAV(fundID string, navMic int64, asOf time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.latest[fundID] = model.MarketPrice{FundID: fundID, NAVMic: navMic, AsOf: asOf}
}
