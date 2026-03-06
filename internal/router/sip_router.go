package router

import "net/http"

func (r *MainRouter) registerSIPRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/sip/sips", r.sipController.CreateSIP)
	mux.HandleFunc("GET /api/v1/sip/sips/{id}", r.sipController.GetSIPByID)
	mux.HandleFunc("PATCH /api/v1/sip/sips/{id}/pause", r.sipController.PauseSIP)
	mux.HandleFunc("PATCH /api/v1/sip/sips/{id}/unpause", r.sipController.UnpauseSIP)
	mux.HandleFunc("PATCH /api/v1/sip/sips/{id}/stop", r.sipController.StopSIP)
	mux.HandleFunc("POST /api/v1/sip/sips/{id}/catchup", r.sipController.CatchUpSIP)
}
