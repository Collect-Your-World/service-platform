package worker

import (
	"backend/service-platform/app/internal/runtime"
	service "backend/service-platform/app/service"
	"context"

	"go.uber.org/zap"
)

type Server runtime.Resource

func (s *Server) Start(ctx context.Context) {
	res := runtime.Resource(*s)

	// Use worker config from resource
	workerConfig := res.Config.WorkerConfig

	// Initialize services to get WorkerService
	services := service.NewServices(res, workerConfig)

	s.Logger.Info("Starting dedicated worker server")

	// Start worker service (blocks until context is cancelled)
	if err := services.WorkerService.Start(ctx); err != nil {
		s.Logger.Error("Worker service failed", zap.Error(err))
	}
}
