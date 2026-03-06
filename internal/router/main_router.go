package router

import (
	"net/http"
	"sync"

	"SIP/internal/controller"
)

type MainRouter struct {
	fundController      *controller.FundController
	sipController       *controller.SIPController
	portfolioController *controller.PortfolioController
	userController      *controller.UserController
}

var (
	mainRouterOnce sync.Once
	mainRouterInst *MainRouter
)

func NewMainRouter(
	fundController *controller.FundController,
	sipController *controller.SIPController,
	portfolioController *controller.PortfolioController,
	userController *controller.UserController,
) *MainRouter {
	mainRouterOnce.Do(func() {
		mainRouterInst = &MainRouter{
			fundController:      fundController,
			sipController:       sipController,
			portfolioController: portfolioController,
			userController:      userController,
		}
	})
	return mainRouterInst
}

func (r *MainRouter) Handler() http.Handler {
	mux := http.NewServeMux()
	r.registerFundRoutes(mux)
	r.registerSIPRoutes(mux)
	r.registerPortfolioRoutes(mux)
	r.registerPaymentRoutes(mux)
	r.registerUserRoutes(mux)
	return mux
}
