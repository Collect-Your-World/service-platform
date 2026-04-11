package platform

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/http2"

	appmanagers "backend/service-platform/app/internal/managers"
	platctrl "backend/service-platform/app/internal/platform/controllers"
	echomw "backend/service-platform/app/internal/platform/middleware"
	"backend/service-platform/app/internal/platform/router"
	"backend/service-platform/app/internal/platform/runtime"
	"backend/service-platform/app/internal/platform/validator"
	"backend/service-platform/app/internal/repository"
	ctxutil "backend/service-platform/app/pkg/util/context"
)

type Server runtime.Resource

func (s *Server) Start(ctx context.Context) {
	res := runtime.Resource(*s)

	appComponents := s.initializeComponents(res)

	serverChannel := s.startHTTPServer(*appComponents.router)

	s.waitForShutdownSignal(serverChannel)

	s.performGracefulShutdown(*appComponents.router)
}

type AppComponents struct {
	repositories *repository.Repositories
	managers     *appmanagers.Managers
	controllers  *platctrl.Controllers
	validators   *validator.Validators
	middleware   *echomw.Middleware
	router       *router.Router
}

func (s *Server) initializeComponents(res runtime.Resource) *AppComponents {
	s.Logger.Info("Initializing application components")

	repositories := repository.NewRepositories(res)

	mgrs := appmanagers.NewManagers(res, nil, repositories)
	controllers := platctrl.NewControllers(mgrs, res)

	validators := validator.NewValidators(res)
	if err := validators.Setup(); err != nil {
		s.Logger.Error("Failed to setup validators", zap.Error(err))
		panic(err)
	}

	newMiddleware := echomw.NewMiddleware(res)
	newRouter := router.NewRouter(res, validators, newMiddleware, controllers, repositories)

	return &AppComponents{
		repositories: repositories,
		managers:     mgrs,
		controllers:  controllers,
		validators:   validators,
		middleware:   newMiddleware,
		router:       newRouter,
	}
}

func (s *Server) startHTTPServer(rtr router.Router) chan error {
	channel := make(chan error, 1)

	go func() {
		address := fmt.Sprintf(":%d", s.Config.ServerConfig.Port)
		s.Logger.Info("Starting HTTP Server", zap.String("address", address))
		channel <- rtr.StartH2CServer(address, &http2.Server{})
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

func (s *Server) performGracefulShutdown(rtr router.Router) {
	s.Logger.Info("Starting graceful shutdown")
	gracefulCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.shutdownHTTPServer(gracefulCtx, rtr)

	s.Logger.Info("Shutdown complete")
}

func (s *Server) shutdownHTTPServer(ctx context.Context, rtr router.Router) {
	s.Logger.Info("Shutting down HTTP server")
	if err := rtr.Shutdown(ctx); err != nil {
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
