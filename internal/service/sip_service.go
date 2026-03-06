package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"SIP/internal/model"
	"SIP/internal/repository"
	"SIP/internal/util"
)

var ErrUnauthorized = errors.New("sip does not belong to user")

type SIPService struct {
	sipRepo  repository.SIPRepository
	fundRepo repository.FundRepository
	userRepo repository.UserRepository
	idGen    *util.IDGenerator
	loc      *time.Location
	now      func() time.Time
}

var (
	sipServiceOnce sync.Once
	sipServiceInst *SIPService
)

type CreateSIPInput struct {
	UserID          string
	FundID          string
	Mode            model.SIPMode
	StartAt         time.Time
	BaseAmountPaise int64
	StepUpEnabled   bool
	StepUpBps       int32
}

func NewSIPService(
	sipRepo repository.SIPRepository,
	fundRepo repository.FundRepository,
	userRepo repository.UserRepository,
	idGen *util.IDGenerator,
	loc *time.Location,
	now func() time.Time,
) *SIPService {
	sipServiceOnce.Do(func() {
		sipServiceInst = &SIPService{sipRepo: sipRepo, fundRepo: fundRepo, userRepo: userRepo, idGen: idGen, loc: loc, now: now}
	})
	return sipServiceInst
}

func (s *SIPService) CreateSIP(ctx context.Context, in CreateSIPInput) (model.SIP, error) {
	if _, err := s.userRepo.GetByID(ctx, in.UserID); err != nil {
		return model.SIP{}, err
	}
	fund, err := s.fundRepo.GetByID(ctx, in.FundID)
	if err != nil {
		return model.SIP{}, err
	}
	if !fund.IsActive {
		return model.SIP{}, fmt.Errorf("%w: fund %s is inactive", util.ErrValidation, fund.FundID)
	}
	if err := util.ValidateCreateSIP(in.Mode, in.StartAt, in.BaseAmountPaise, in.StepUpEnabled, in.StepUpBps); err != nil {
		return model.SIP{}, err
	}

	now := s.now().In(s.loc)
	start := in.StartAt.In(s.loc)
	sip := model.SIP{
		SIPID:            s.idGen.Next("sip"),
		UserID:           in.UserID,
		FundID:           in.FundID,
		Mode:             in.Mode,
		StartAt:          start,
		NextRunAt:        start,
		AnchorWeekday:    start.Weekday(),
		AnchorDayOfMonth: start.Day(),
		AnchorHour:       start.Hour(),
		AnchorMinute:     start.Minute(),
		AnchorSecond:     start.Second(),
		AnchorNanosecond: start.Nanosecond(),
		BaseAmountPaise:  in.BaseAmountPaise,
		StepUpEnabled:    in.StepUpEnabled,
		StepUpBps:        in.StepUpBps,
		Status:           model.SIPStatusActive,
		CreatedAt:        now,
		UpdatedAt:        now,
		Version:          1,
	}
	if err := sip.Validate(); err != nil {
		return model.SIP{}, err
	}
	return s.sipRepo.Create(ctx, sip)
}

func (s *SIPService) PauseSIP(ctx context.Context, sipID, userID string) (model.SIP, error) {
	sip, err := s.sipRepo.GetByID(ctx, sipID)
	if err != nil {
		return model.SIP{}, err
	}
	if sip.UserID != userID {
		return model.SIP{}, ErrUnauthorized
	}
	if err := sip.Pause(s.now().In(s.loc)); err != nil {
		return model.SIP{}, err
	}
	return s.sipRepo.Update(ctx, sip)
}

func (s *SIPService) UnpauseSIP(ctx context.Context, sipID, userID string) (model.SIP, error) {
	sip, err := s.sipRepo.GetByID(ctx, sipID)
	if err != nil {
		return model.SIP{}, err
	}
	if sip.UserID != userID {
		return model.SIP{}, ErrUnauthorized
	}
	if err := sip.Unpause(s.now().In(s.loc)); err != nil {
		return model.SIP{}, err
	}
	return s.sipRepo.Update(ctx, sip)
}

func (s *SIPService) StopSIP(ctx context.Context, sipID, userID string) (model.SIP, error) {
	sip, err := s.sipRepo.GetByID(ctx, sipID)
	if err != nil {
		return model.SIP{}, err
	}
	if sip.UserID != userID {
		return model.SIP{}, ErrUnauthorized
	}
	sip.Stop(s.now().In(s.loc))
	return s.sipRepo.Update(ctx, sip)
}
