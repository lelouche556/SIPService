package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"SIP/internal/dto"
	"SIP/internal/model"
	"SIP/internal/service"
	"SIP/internal/util"
)

type SIPController struct {
	sipService       *service.SIPService
	executionService *service.ExecutionService
	portfolioService *service.PortfolioService
	loc              *time.Location
}

var (
	sipControllerOnce sync.Once
	sipControllerInst *SIPController
)

func NewSIPController(sipService *service.SIPService, executionService *service.ExecutionService, portfolioService *service.PortfolioService, loc *time.Location) *SIPController {
	sipControllerOnce.Do(func() {
		sipControllerInst = &SIPController{
			sipService:       sipService,
			executionService: executionService,
			portfolioService: portfolioService,
			loc:              loc,
		}
	})
	return sipControllerInst
}

func (c *SIPController) CreateSIP(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSIPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteErr(w, http.StatusBadRequest, err)
		return
	}
	startAt, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		util.WriteErr(w, http.StatusBadRequest, err)
		return
	}
	if req.StepUpEnabled && (req.StepUpBps < 0 || req.StepUpBps > 10000) {
		verr := fmt.Errorf("%w: stepUpBps must be in [0,10000]", util.ErrValidation)
		util.WriteErr(w, util.HTTPStatus(verr), verr)
		return
	}
	now := time.Now().In(c.loc)
	if startAt.In(c.loc).Before(now) {
		verr := fmt.Errorf("%w: startAt must be >= now", util.ErrValidation)
		util.WriteErr(w, util.HTTPStatus(verr), verr)
		return
	}
	userID := userIDFromRequest(r)
	if req.UserID != "" {
		userID = req.UserID
	}
	sip, err := c.sipService.CreateSIP(r.Context(), service.CreateSIPInput{
		UserID:          userID,
		FundID:          req.FundID,
		Mode:            model.SIPMode(strings.ToUpper(req.Mode)),
		StartAt:         startAt.In(c.loc),
		BaseAmountPaise: req.BaseAmountPaise,
		StepUpEnabled:   req.StepUpEnabled,
		StepUpBps:       req.StepUpBps,
	})
	if err != nil {
		status := util.HTTPStatus(err)
		if errors.Is(err, service.ErrUnauthorized) {
			status = http.StatusForbidden
		}
		util.WriteErr(w, status, err)
		return
	}
	util.WriteJSON(w, http.StatusCreated, sip)
}

func (c *SIPController) GetSIPByID(w http.ResponseWriter, r *http.Request) {
	sipID := r.PathValue("id")
	userID := userIDFromRequest(r)
	sip, insts, err := c.portfolioService.GetSIPDetail(r.Context(), sipID, userID)
	if err != nil {
		status := util.HTTPStatus(err)
		if errors.Is(err, service.ErrUnauthorized) {
			status = http.StatusForbidden
		}
		util.WriteErr(w, status, err)
		return
	}
	util.WriteJSON(w, http.StatusOK, map[string]any{"sip": sip, "installments": insts})
}

func (c *SIPController) PauseSIP(w http.ResponseWriter, r *http.Request) {
	sipID := r.PathValue("id")
	userID := userIDFromRequest(r)
	sip, err := c.sipService.PauseSIP(r.Context(), sipID, userID)
	if err != nil {
		status := util.HTTPStatus(err)
		if errors.Is(err, service.ErrUnauthorized) {
			status = http.StatusForbidden
		}
		util.WriteErr(w, status, err)
		return
	}
	util.WriteJSON(w, http.StatusOK, sip)
}

func (c *SIPController) UnpauseSIP(w http.ResponseWriter, r *http.Request) {
	sipID := r.PathValue("id")
	userID := userIDFromRequest(r)
	sip, err := c.sipService.UnpauseSIP(r.Context(), sipID, userID)
	if err != nil {
		status := util.HTTPStatus(err)
		if errors.Is(err, service.ErrUnauthorized) {
			status = http.StatusForbidden
		}
		util.WriteErr(w, status, err)
		return
	}
	util.WriteJSON(w, http.StatusOK, sip)
}

func (c *SIPController) StopSIP(w http.ResponseWriter, r *http.Request) {
	sipID := r.PathValue("id")
	userID := userIDFromRequest(r)
	sip, err := c.sipService.StopSIP(r.Context(), sipID, userID)
	if err != nil {
		status := util.HTTPStatus(err)
		if errors.Is(err, service.ErrUnauthorized) {
			status = http.StatusForbidden
		}
		util.WriteErr(w, status, err)
		return
	}
	util.WriteJSON(w, http.StatusOK, sip)
}

func (c *SIPController) CatchUpSIP(w http.ResponseWriter, r *http.Request) {
	sipID := r.PathValue("id")
	userID := userIDFromRequest(r)
	var req dto.LumpSumRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteErr(w, http.StatusBadRequest, err)
		return
	}
	inst, err := c.executionService.CatchUp(r.Context(), sipID, userID, req.NumInstallments, time.Now().In(c.loc))
	if err != nil {
		util.WriteErr(w, http.StatusBadRequest, err)
		return
	}
	util.WriteJSON(w, http.StatusOK, inst)
}

func (c *SIPController) PaymentCallback(w http.ResponseWriter, r *http.Request) {
	var req dto.PaymentCallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteErr(w, http.StatusBadRequest, err)
		return
	}
	inst, err := c.executionService.HandlePaymentCallback(
		r.Context(), req.PaymentRequestID, model.PaymentStatus(strings.ToUpper(req.Status)), req.FailureReason,
	)
	if err != nil {
		util.WriteErr(w, util.HTTPStatus(err), err)
		return
	}
	util.WriteJSON(w, http.StatusOK, inst)
}

func userIDFromRequest(r *http.Request) string {
	if userID := strings.TrimSpace(r.Header.Get("X-User-Id")); userID != "" {
		return userID
	}
	if userID := strings.TrimSpace(r.URL.Query().Get("userId")); userID != "" {
		return userID
	}
	return "user-1"
}
