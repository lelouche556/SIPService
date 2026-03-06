package router

import "net/http"

func (r *MainRouter) registerFundRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/fund/funds", r.fundController.GetFunds)
	mux.HandleFunc("POST /api/v1/fund/funds", r.fundController.CreateFund)
}
