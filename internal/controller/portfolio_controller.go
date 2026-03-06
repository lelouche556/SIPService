package controller

import (
	"net/http"
	"sync"

	"SIP/internal/service"
	"SIP/internal/util"
)

type PortfolioController struct {
	portfolioService *service.PortfolioService
}

var (
	portfolioControllerOnce sync.Once
	portfolioControllerInst *PortfolioController
)

func NewPortfolioController(portfolioService *service.PortfolioService) *PortfolioController {
	portfolioControllerOnce.Do(func() {
		portfolioControllerInst = &PortfolioController{
			portfolioService: portfolioService,
		}
	})
	return portfolioControllerInst
}

func (c *PortfolioController) GetPortfolio(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		userID = "user-1"
	}
	offset, limit, err := util.ParsePagination(r.URL.Query().Get("offset"), r.URL.Query().Get("limit"))
	if err != nil {
		util.WriteErr(w, util.HTTPStatus(err), err)
		return
	}
	portfolio, err := c.portfolioService.GetPortfolio(r.Context(), userID)
	if err != nil {
		util.WriteErr(w, util.HTTPStatus(err), err)
		return
	}
	portfolio.SIPs = util.PaginateSlice(portfolio.SIPs, offset, limit)
	util.WriteJSON(w, http.StatusOK, portfolio)
}
