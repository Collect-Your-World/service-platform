package middleware

import (
	"backend/service-platform/app/internal/runtime"

	"github.com/labstack/echo/v4"
)

type Middleware struct {
	JwtAuthentication       JwtAuthentication
	ApiKeyAuthentication    ApiKeyAuthentication
	HttpBasicAuthentication HttpBasicAuthentication
}

func NewMiddleware(res runtime.Resource) *Middleware {
	return &Middleware{
		JwtAuthentication:       NewJwtAuthentication(res),
		ApiKeyAuthentication:    NewApiKeyAuthentication(res),
		HttpBasicAuthentication: NewHttpBasicAuthentication(res),
	}
}

func (m *Middleware) RequireAuth() echo.MiddlewareFunc {
	return m.JwtAuthentication.RequireAuth()
}

func (m *Middleware) RequireRole(requiredRole string) echo.MiddlewareFunc {
	return m.JwtAuthentication.RequireRole(requiredRole)
}
