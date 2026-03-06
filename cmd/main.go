package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"SIP/internal/controller"
	"SIP/internal/repository"
	"SIP/internal/router"
	"SIP/internal/scheduler"
	"SIP/internal/service"
	"SIP/internal/util"
)

func main() {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Fatalf("timezone load failed: %v", err)
	}
	nowFn := time.Now

	idGen := util.NewIDGenerator(1000)

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

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	sipScheduler := scheduler.NewSIPScheduler(executionService, 5*time.Second, nowFn)
	go sipScheduler.Start(ctx)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mainRouter.Handler(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, done := context.WithTimeout(context.Background(), 5*time.Second)
		defer done()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Printf("SIP system listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
