package service

import (
	"context"
	"sync"

	"SIP/internal/model"
	"SIP/internal/repository"
)

type PortfolioService struct {
	sipRepo         repository.SIPRepository
	installmentRepo repository.InstallmentRepository
}

var (
	portfolioServiceOnce sync.Once
	portfolioServiceInst *PortfolioService
)

type PortfolioView struct {
	SIPs         []model.SIP `json:"sips"`
	ActiveCount  int         `json:"activeCount"`
	PausedCount  int         `json:"pausedCount"`
	StoppedCount int         `json:"stoppedCount"`
}

func NewPortfolioService(sipRepo repository.SIPRepository, installmentRepo repository.InstallmentRepository) *PortfolioService {
	portfolioServiceOnce.Do(func() {
		portfolioServiceInst = &PortfolioService{sipRepo: sipRepo, installmentRepo: installmentRepo}
	})
	return portfolioServiceInst
}

func (s *PortfolioService) GetPortfolio(ctx context.Context, userID string) (PortfolioView, error) {
	sips, err := s.sipRepo.ListByUser(ctx, userID)
	if err != nil {
		return PortfolioView{}, err
	}
	view := PortfolioView{SIPs: sips}
	for _, sip := range sips {
		switch sip.Status {
		case model.SIPStatusActive:
			view.ActiveCount++
		case model.SIPStatusPaused:
			view.PausedCount++
		case model.SIPStatusStopped:
			view.StoppedCount++
		}
	}
	return view, nil
}

func (s *PortfolioService) GetSIPDetail(ctx context.Context, sipID, userID string) (model.SIP, []model.Installment, error) {
	sip, err := s.sipRepo.GetByID(ctx, sipID)
	if err != nil {
		return model.SIP{}, nil, err
	}
	if sip.UserID != userID {
		return model.SIP{}, nil, ErrUnauthorized
	}
	insts, err := s.installmentRepo.ListBySIP(ctx, sipID)
	if err != nil {
		return model.SIP{}, nil, err
	}
	return sip, insts, nil
}
