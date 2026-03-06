package router

import "net/http"

func (r *MainRouter) registerUserRoutes(mux *http.ServeMux) {
	if r.userController == nil {
		return
	}
	mux.HandleFunc("POST /api/v1/user/users", r.userController.CreateUser)
	mux.HandleFunc("GET /api/v1/user/users", r.userController.ListUsers)
}
