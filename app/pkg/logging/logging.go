package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	ctxutil "backend/service-platform/app/pkg/util/context"
)

type LogConfig struct {
	ServiceName string
	Env         ctxutil.AppMode
}

func NewLogConfig(serviceName string, appMode ctxutil.AppMode) *LogConfig {
	return &LogConfig{
		ServiceName: serviceName,
		Env:         appMode,
	}
}

func (cfg *LogConfig) NewLogging() (*zap.Logger, error) {
	logLevel := getLogLevel(cfg.Env)
	zapConfig := zap.NewProductionConfig()
	encoderLevel := zapcore.LowercaseLevelEncoder

	if cfg.Env != ctxutil.AppModeProd {
		zapConfig = zap.NewDevelopmentConfig()
	}

	zapConfig.Level = zap.NewAtomicLevelAt(logLevel)
	zapConfig.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    encoderLevel,
		EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// set logger for db using env
	//cfg.setDbLogger(cfg.Env == ctxutil.AppModeLocal)

	if cfg.Env == ctxutil.AppModeLocal { // early return
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		return zapConfig.Build()
	}

	zapConfig.Level = zap.NewAtomicLevelAt(logLevel)
	jsonEncoder := zapcore.NewJSONEncoder(zapConfig.EncoderConfig)
	// Note: maybe we also need to write file in parallel in future so using Tee
	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, zapcore.AddSync(os.Stdout), getLogLevel(cfg.Env)),
	)
	return zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	), nil
}

func getLogLevel(appMode ctxutil.AppMode) zapcore.Level {
	switch appMode {
	case ctxutil.AppModeProd, ctxutil.AppModeTest:
		return zapcore.WarnLevel
	case ctxutil.AppModeDev:
		return zapcore.InfoLevel
	default:
		return zapcore.InfoLevel
	}
}
