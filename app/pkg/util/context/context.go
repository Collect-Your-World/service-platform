package ctxutil

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

type AppMode string

const (
	AppModeLocal AppMode = "local"
	AppModeTest  AppMode = "test"
	AppModeDev   AppMode = "dev"
	AppModeProd  AppMode = "production"
)

func SetAppMode(ctx context.Context, appMode AppMode) context.Context {
	return context.WithValue(ctx, "app_mode", appMode)
}

func GetAppModeFromEnv() AppMode {
	env := strings.ToLower(os.Getenv("APP_ENV"))
	switch env {
	case string(AppModeLocal):
		return AppModeLocal
	case string(AppModeTest):
		return AppModeTest
	case string(AppModeDev):
		return AppModeDev
	default:
		return AppModeLocal
	}
}

func GetContextValue(ctx context.Context, key string) (string, error) {
	value := ctx.Value(key)
	if value == nil {
		return "", errors.New(fmt.Sprintf("key %s not exists", key))
	}

	str, ok := value.(string)
	if !ok {
		return "", errors.New(fmt.Sprintf("key %s not valid", key))
	}

	return str, nil
}

func GetContextValueAnyType(ctx context.Context, key string) (any, error) {
	value := ctx.Value(key)
	if value == nil {
		return "", errors.New(fmt.Sprintf("key %s not exists", key))
	}

	return value, nil
}
