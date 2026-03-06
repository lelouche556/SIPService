package router

import "net/http"

func (r *MainRouter) registerPortfolioRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/sip/portfolio", r.portfolioController.GetPortfolio)
}
