package worker

import (
	"github.com/go-co-op/gocron/v2"
	"go.uber.org/zap"
)

type WorkerLog struct {
	logger *zap.SugaredLogger
}

// Debug implements gocron.Logger.
func (w *WorkerLog) Debug(msg string, args ...any) {
	args = append([]any{msg}, args...)
	w.logger.Debug(args...)
}

// Error implements gocron.Logger.
func (w *WorkerLog) Error(msg string, args ...any) {
	args = append([]any{msg}, args...)
	w.logger.Error(args...)
}

// Info implements gocron.Logger.
func (w *WorkerLog) Info(msg string, args ...any) {
	args = append([]any{msg}, args...)
	w.logger.Info(args...)
}

// Warn implements gocron.Logger.
func (w *WorkerLog) Warn(msg string, args ...any) {
	args = append([]any{msg}, args...)
	w.logger.Warn(args...)
}

func NewWorkerLog(logger *zap.SugaredLogger) gocron.Logger {
	return &WorkerLog{logger}
}
