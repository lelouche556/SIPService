package router

import "net/http"

func (r *MainRouter) registerPaymentRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/sip/payments/callback", r.sipController.PaymentCallback)
}
