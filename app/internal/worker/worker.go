package worker

import (
	"backend/service-platform/app/internal/platform/runtime"
	"context"

	"go.uber.org/zap"
)

type Server runtime.Resource

func (s *Server) Start(ctx context.Context) {
	res := runtime.Resource(*s)

	workerConfig := res.Config.WorkerConfig

	services := NewServices(res, workerConfig)

	s.Logger.Info("Starting dedicated worker server")

	if err := services.WorkerService.Start(ctx); err != nil {
		s.Logger.Error("Worker service failed", zap.Error(err))
	}
}
