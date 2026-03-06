package controller

import (
	"encoding/json"
	"net/http"
	"sync"

	"SIP/internal/dto"
	"SIP/internal/repository"
	"SIP/internal/service"
	"SIP/internal/util"
)

type FundController struct {
	fundService *service.FundService
}

var (
	fundControllerOnce sync.Once
	fundControllerInst *FundController
)

func NewFundController(fundService *service.FundService) *FundController {
	fundControllerOnce.Do(func() {
		fundControllerInst = &FundController{
			fundService: fundService,
		}
	})
	return fundControllerInst
}

func (c *FundController) GetFunds(w http.ResponseWriter, r *http.Request) {
	filters := repository.FundFilters{
		Query:    r.URL.Query().Get("query"),
		Category: r.URL.Query().Get("category"),
		AMC:      r.URL.Query().Get("amc"),
		RiskTag:  r.URL.Query().Get("riskTag"),
	}
	offset, limit, err := util.ParsePagination(r.URL.Query().Get("offset"), r.URL.Query().Get("limit"))
	if err != nil {
		util.WriteErr(w, util.HTTPStatus(err), err)
		return
	}
	withPrice := r.URL.Query().Get("withPrice") == "true"
	if withPrice {
		funds, err := c.fundService.ListFundsWithPrice(r.Context(), filters)
		if err != nil {
			util.WriteErr(w, util.HTTPStatus(err), err)
			return
		}
		util.WriteJSON(w, http.StatusOK, util.PaginateSlice(funds, offset, limit))
		return
	}
	funds, err := c.fundService.BrowseFunds(r.Context(), filters)
	if err != nil {
		util.WriteErr(w, util.HTTPStatus(err), err)
		return
	}
	util.WriteJSON(w, http.StatusOK, util.PaginateSlice(funds, offset, limit))
}

func (c *FundController) CreateFund(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateFundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteErr(w, util.HTTPStatus(err), err)
		return
	}
	fund, err := c.fundService.CreateFund(r.Context(), req.FundID, req.Name, req.AMC, req.Category, req.RiskTag, req.IsActive, req.NAVMic)
	if err != nil {
		util.WriteErr(w, util.HTTPStatus(err), err)
		return
	}
	util.WriteJSON(w, http.StatusCreated, fund)
}
