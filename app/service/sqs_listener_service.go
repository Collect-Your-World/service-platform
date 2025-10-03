package service

import (
	"backend/service-platform/app/internal/runtime"
	"backend/service-platform/app/pkg/sqs"
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

type SQSListenerService struct {
	listener *sqs.Listener
	sqsQueue *sqs.Queue
	logger   *zap.Logger
	config   sqs.Config
	running  bool
	mu       sync.RWMutex
}

type SQSListenerConfig struct {
	SQSConfig  sqs.Config
	JobManager interface{}
	QueueURLs  []string
}

func NewSQSListenerService(res runtime.Resource, config SQSListenerConfig) (*SQSListenerService, error) {
	logger := res.Logger.With(zap.String("component", "sqs_listener_service"))

	sqsQueue, err := sqs.NewSQSQueue(context.Background(), config.SQSConfig, logger)
	if err != nil {
		return nil, err
	}

	bridge := sqs.NewJobManagerBridge(config.JobManager, logger)

	queueURLs := config.QueueURLs
	if len(queueURLs) == 0 {
		queueURLs = []string{
			config.SQSConfig.QueueURLs.SqsScheduledJobQueue,
		}
	}

	var validURLs []string
	for _, url := range queueURLs {
		if url != "" {
			validURLs = append(validURLs, url)
		}
	}

	listener, err := sqs.NewListener(
		context.Background(),
		config.SQSConfig,
		sqs.ListenerConfig{
			QueueURLs:   validURLs,
			WorkerCount: 1,
			Handler:     bridge,
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	return &SQSListenerService{
		listener: listener,
		sqsQueue: sqsQueue,
		logger:   logger,
		config:   config.SQSConfig,
	}, nil
}

func (s *SQSListenerService) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("SQS listener service is already running")
	}
	s.running = true
	s.mu.Unlock()

	s.logger.Info("Starting SQS listener service")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.listener.Start(ctx); err != nil {
			s.logger.Error("SQS listener failed", zap.Error(err))
		}
	}()

	s.logger.Info("SQS listener service started successfully")

	<-ctx.Done()
	s.logger.Info("Shutting down SQS listener service")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.listener.Stop(shutdownCtx); err != nil {
		s.logger.Error("Failed to stop SQS listener gracefully", zap.Error(err))
	}

	wg.Wait()

	s.mu.Lock()
	s.running = false
	s.mu.Unlock()

	s.logger.Info("SQS listener service stopped")
	return nil
}

func (s *SQSListenerService) Stop(ctx context.Context) error {
	s.mu.RLock()
	running := s.running
	s.mu.RUnlock()

	if !running {
		return nil
	}

	return s.listener.Stop(ctx)
}

func (s *SQSListenerService) GetStats() SQSListenerStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	listenerStats := s.listener.GetStats()

	return SQSListenerStats{
		Running:       s.running,
		ListenerStats: listenerStats,
	}
}

func (s *SQSListenerService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *SQSListenerService) GetQueue() *sqs.Queue {
	return s.sqsQueue
}

type SQSListenerStats struct {
	Running       bool              `json:"running"`
	ListenerStats sqs.ListenerStats `json:"listener_stats"`
}
