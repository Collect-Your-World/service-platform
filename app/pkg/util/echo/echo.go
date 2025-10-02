package echoutil

import (
	"errors"
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"backend/service-platform/app/internal/runtime"
)

func SetupCORSMiddleware(res runtime.Resource) echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: strings.Split(res.Config.RouterConfig.AllowedOrigins, ","),
		AllowHeaders: []string{
			echo.HeaderAuthorization,
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderCookie,
			echo.HeaderSetCookie,
			echo.HeaderAccessControlAllowHeaders,
			"Cf-Turnstile-Token",
		},
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowCredentials: true,
	})
}

func SetupLoggerMiddleware(res runtime.Resource) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		HandleError:  true,
		LogLatency:   true,
		LogProtocol:  true,
		LogRemoteIP:  true,
		LogHost:      true,
		LogMethod:    true,
		LogURI:       true,
		LogURIPath:   true,
		LogRoutePath: true,
		LogRequestID: true,
		LogReferer:   true,
		LogUserAgent: true,
		LogStatus:    true,
		LogError:     true,
		Skipper: func(c echo.Context) bool {
			return strings.EqualFold(c.Request().RequestURI, "/inform") || strings.EqualFold(c.Request().RequestURI, "/favicon.ico")
		},
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			fields := []zap.Field{
				// Request context
				zap.String("request_id", v.RequestID),
				zap.String("remote_ip", v.RemoteIP),
				zap.String("host", v.Host),

				// Request details
				zap.String("method", v.Method),
				zap.String("uri", v.URI),
				zap.String("uri_path", v.URIPath),
				zap.String("route", v.RoutePath),
				zap.String("referer", v.Referer),
				zap.String("user_agent", v.UserAgent),
				zap.String("protocol", v.Protocol),

				// Response details
				zap.Int("status", v.Status),
				zap.Duration("latency", v.Latency),
			}
			if v.Error != nil {
				res.Logger.Error("request failed", append(fields, zap.Error(v.Error))...)
			} else {
				res.Logger.Info("request", fields...)
			}

			return nil
		},
	})
}

func BindAndValidate(c echo.Context, payload any) error {
	if err := c.Bind(payload); err != nil {
		return err
	}
	if err := c.Validate(payload); err != nil {
		return err
	}
	return nil
}

func GetAuthToken(c echo.Context) (string, error) {
	header := c.Request().Header
	authToken := header.Get("Authorization")
	if authToken == "" {
		return "", errors.New("token not found")
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(authToken, prefix) {
		return "", errors.New("invalid authorization header")
	}
	token := strings.TrimPrefix(authToken, prefix)
	if token == "" {
		return "", fmt.Errorf("empty auth token")
	}

	return token, nil
}
