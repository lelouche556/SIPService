package controller_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"SIP/internal/controller"
	"SIP/internal/model"
	"SIP/internal/repository"
	"SIP/internal/router"
	"SIP/internal/service"
	"SIP/internal/util"
)

type apiClient struct {
	handler http.Handler
	t       *testing.T
}

func (c *apiClient) do(method, path string, body any, out any) *http.Response {
	c.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			c.t.Fatalf("encode body: %v", err)
		}
	}
	req, err := http.NewRequest(method, path, &buf)
	if err != nil {
		c.t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	c.handler.ServeHTTP(rec, req)
	resp := rec.Result()
	if out != nil {
		if err := json.NewDecoder(rec.Body).Decode(out); err != nil {
			c.t.Fatalf("decode response: %v", err)
		}
	}
	return resp
}

func TestAPIEndpoints(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		t.Fatalf("timezone load failed: %v", err)
	}
	nowFn := time.Now

	idGen := util.NewIDGenerator(2000)
	fundRepo := repository.NewInMemoryFundRepository(nil)
	userRepo := repository.NewInMemoryUserRepository(nil)
	sipRepo := repository.NewInMemorySIPRepository()
	instRepo := repository.NewInMemoryInstallmentRepository()

	pricingService := service.NewPricingService(nil)
	paymentService := service.NewPaymentService(nowFn)
	userService := service.NewUserService(userRepo, idGen)
	fundService := service.NewFundService(fundRepo, pricingService, idGen, nowFn)
	sipService := service.NewSIPService(sipRepo, fundRepo, userRepo, idGen, loc, nowFn)
	executionService := service.NewExecutionService(sipRepo, instRepo, pricingService, paymentService, idGen, loc, nowFn)
	portfolioService := service.NewPortfolioService(sipRepo, instRepo)

	fundController := controller.NewFundController(fundService)
	sipController := controller.NewSIPController(sipService, executionService, portfolioService, loc)
	portfolioController := controller.NewPortfolioController(portfolioService)
	userController := controller.NewUserController(userService)
	mainRouter := router.NewMainRouter(fundController, sipController, portfolioController, userController)

	client := &apiClient{handler: mainRouter.Handler(), t: t}

	// POST /api/v1/user/users
	var user model.User
	resp := client.do(http.MethodPost, "/api/v1/user/users", map[string]any{"userId": "user-1", "name": "Test User"}, &user)
	if resp.StatusCode != http.StatusCreated || user.UserID == "" {
		t.Fatalf("create user failed status %d", resp.StatusCode)
	}

	// GET /api/v1/user/users
	var users []model.User
	resp = client.do(http.MethodGet, "/api/v1/user/users", nil, &users)
	if resp.StatusCode != http.StatusOK || len(users) == 0 {
		t.Fatalf("list users failed status %d", resp.StatusCode)
	}

	// POST /api/v1/fund/funds
	var fund model.Fund
	resp = client.do(http.MethodPost, "/api/v1/fund/funds", map[string]any{
		"fundId":   "fund-1",
		"name":     "Bluechip Equity Growth",
		"amc":      "Alpha AMC",
		"category": "Equity",
		"riskTag":  "High",
		"isActive": true,
		"navMic":   12543210,
	}, &fund)
	if resp.StatusCode != http.StatusCreated || fund.FundID == "" {
		t.Fatalf("create fund failed status %d", resp.StatusCode)
	}

	// GET /api/v1/fund/funds
	var funds []model.Fund
	resp = client.do(http.MethodGet, "/api/v1/fund/funds", nil, &funds)
	if resp.StatusCode != http.StatusOK || len(funds) == 0 {
		t.Fatalf("expected funds, status %d len %d", resp.StatusCode, len(funds))
	}

	// POST /api/v1/sip/sips
	createReq := map[string]any{
		"userId":          "user-1",
		"fundId":          "fund-1",
		"mode":            "MONTHLY",
		"startAt":         time.Now().In(loc).Add(time.Minute).Format(time.RFC3339),
		"baseAmountPaise": 100000,
		"stepUpEnabled":   true,
		"stepUpBps":       1000,
	}
	var created model.SIP
	resp = client.do(http.MethodPost, "/api/v1/sip/sips", createReq, &created)
	if resp.StatusCode != http.StatusCreated || created.SIPID == "" {
		t.Fatalf("create sip failed status %d", resp.StatusCode)
	}

	// GET /api/v1/sip/portfolio
	var portfolio map[string]any
	resp = client.do(http.MethodGet, "/api/v1/sip/portfolio?userId=user-1", nil, &portfolio)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("portfolio failed status %d", resp.StatusCode)
	}

	// GET /api/v1/sip/sips/{id}
	var sipDetail map[string]any
	resp = client.do(http.MethodGet, "/api/v1/sip/sips/"+created.SIPID+"?userId=user-1", nil, &sipDetail)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get sip failed status %d", resp.StatusCode)
	}

	// PATCH /api/v1/sip/sips/{id}/pause
	resp = client.do(http.MethodPatch, "/api/v1/sip/sips/"+created.SIPID+"/pause?userId=user-1", nil, &created)
	if resp.StatusCode != http.StatusOK || created.Status != model.SIPStatusPaused {
		t.Fatalf("pause failed status %d status %s", resp.StatusCode, created.Status)
	}

	// PATCH /api/v1/sip/sips/{id}/unpause
	resp = client.do(http.MethodPatch, "/api/v1/sip/sips/"+created.SIPID+"/unpause?userId=user-1", nil, &created)
	if resp.StatusCode != http.StatusOK || created.Status != model.SIPStatusActive {
		t.Fatalf("unpause failed status %d status %s", resp.StatusCode, created.Status)
	}

	// POST /api/v1/sip/sips/{id}/catchup
	var catchup model.Installment
	resp = client.do(http.MethodPost, "/api/v1/sip/sips/"+created.SIPID+"/catchup?userId=user-1", map[string]any{"numInstallments": 2}, &catchup)
	if resp.StatusCode != http.StatusOK || catchup.PaymentRequestID == "" {
		t.Fatalf("catchup failed status %d", resp.StatusCode)
	}

	// POST /api/v1/sip/payments/callback
	var updated model.Installment
	resp = client.do(http.MethodPost, "/api/v1/sip/payments/callback", map[string]any{
		"paymentRequestId": catchup.PaymentRequestID,
		"status":           "SUCCESS",
		"failureReason":    "",
	}, &updated)
	if resp.StatusCode != http.StatusOK || updated.PaymentStatus != model.PaymentStatusSuccess {
		t.Fatalf("callback failed status %d status %s", resp.StatusCode, updated.PaymentStatus)
	}

	// PATCH /api/v1/sip/sips/{id}/stop
	resp = client.do(http.MethodPatch, "/api/v1/sip/sips/"+created.SIPID+"/stop?userId=user-1", nil, &created)
	if resp.StatusCode != http.StatusOK || created.Status != model.SIPStatusStopped {
		t.Fatalf("stop failed status %d status %s", resp.StatusCode, created.Status)
	}
}
