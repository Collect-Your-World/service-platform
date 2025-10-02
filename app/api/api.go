package api

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/http2"

	"backend/service-platform/app/api/controller"
	"backend/service-platform/app/api/middleware"
	"backend/service-platform/app/api/router"
	"backend/service-platform/app/database/repository"
	"backend/service-platform/app/internal/runtime"
	"backend/service-platform/app/internal/validator"
	"backend/service-platform/app/manager"
	ctxutil "backend/service-platform/app/pkg/util/context"
)

type Server runtime.Resource

func (s *Server) Start(ctx context.Context) {
	res := runtime.Resource(*s)

	// Initialize application components
	appComponents := s.initializeComponents(res)

	// Start HTTP server
	serverChannel := s.startHTTPServer(*appComponents.router)

	// Wait for shutdown signal or server error
	s.waitForShutdownSignal(serverChannel)

	// Perform graceful shutdown
	s.performGracefulShutdown(*appComponents.router)
}

type AppComponents struct {
	repositories *repository.Repositories
	managers     *manager.Managers
	controllers  *controller.Controllers
	validators   *validator.Validators
	middleware   *middleware.Middleware
	router       *router.Router
}

func (s *Server) initializeComponents(res runtime.Resource) *AppComponents {
	s.Logger.Info("Initializing application components")

	// Initialize repositories first
	repositories := repository.NewRepositories(res)

	// Initialize managers
	managers := manager.NewManagers(res, nil, repositories)
	controllers := controller.NewControllers(managers, res)

	validators := validator.NewValidators(res)
	if err := validators.Setup(); err != nil {
		s.Logger.Error("Failed to setup validators", zap.Error(err))
		panic(err)
	}

	newMiddleware := middleware.NewMiddleware(res)
	newRouter := router.NewRouter(res, validators, newMiddleware, controllers, repositories)

	return &AppComponents{
		repositories: repositories,
		managers:     managers,
		controllers:  controllers,
		validators:   validators,
		middleware:   newMiddleware,
		router:       newRouter,
	}
}

// Removed worker and SQS services

func (s *Server) startHTTPServer(router router.Router) chan error {
	channel := make(chan error, 1)

	// Serve HTTP until error
	go func() {
		address := fmt.Sprintf(":%d", s.Config.ServerConfig.Port)
		s.Logger.Info("Starting HTTP Server", zap.String("address", address))
		channel <- router.StartH2CServer(address, &http2.Server{})
	}()

	s.Logger.Info(
		"Serving until error or shutdown",
		zap.Int("port", s.Config.ServerConfig.Port),
		zap.String("env", string(ctxutil.GetAppModeFromEnv())),
	)

	return channel
}

func (s *Server) waitForShutdownSignal(serverChannel chan error) {
	sigChannel := shutdownSignals()
	defer close(sigChannel)

	// Wait for a shutdown signal or server error
	select {
	case sig := <-sigChannel:
		s.Logger.Info(
			"Received shutdown signal",
			zap.String("signal", sig.String()),
		)
	case err := <-serverChannel:
		s.Logger.Error(
			"Received error from server, initiating shutdown",
			zap.Error(err),
		)
	}
}

func (s *Server) performGracefulShutdown(router router.Router) {
	s.Logger.Info("Starting graceful shutdown")
	gracefulCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server last
	s.shutdownHTTPServer(gracefulCtx, router)

	s.Logger.Info("Shutdown complete")
}

// Removed worker and SQS shutdown

func (s *Server) shutdownHTTPServer(ctx context.Context, router router.Router) {
	s.Logger.Info("Shutting down HTTP server")
	if err := router.Shutdown(ctx); err != nil {
		s.Logger.Error(
			"Could not shutdown HTTP server gracefully",
			zap.Error(err),
		)
	} else {
		s.Logger.Info("HTTP Server shutdown gracefully")
	}
}

func shutdownSignals() chan os.Signal {
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, syscall.SIGINT, syscall.SIGTERM)

	return channel
}
