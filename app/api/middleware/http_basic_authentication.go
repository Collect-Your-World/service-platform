package middleware

import (
	"backend/service-platform/app/api/client/response"
	"backend/service-platform/app/internal/runtime"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type HttpBasicAuthentication struct {
	res              runtime.Resource
	validCredentials map[string]string // username -> password
}

func NewHttpBasicAuthentication(res runtime.Resource) HttpBasicAuthentication {
	// Initialize with some example credentials (in production, use proper user management)
	validCredentials := map[string]string{
		"admin": "admin-password", // In production, use hashed passwords
		"api":   "api-password",
	}

	return HttpBasicAuthentication{
		res:              res,
		validCredentials: validCredentials,
	}
}

func (hb *HttpBasicAuthentication) GetName() string {
	return authMethodBasic
}

func (hb *HttpBasicAuthentication) CanHandle(c echo.Context) bool {
	authHeader := c.Request().Header.Get(authHeaderName)
	if authHeader == "" {
		return false
	}

	parts := strings.SplitN(authHeader, " ", tokenParts)
	if len(parts) != tokenParts {
		return false
	}

	return strings.ToLower(parts[0]) == basicPrefix
}

func (hb *HttpBasicAuthentication) GetAuthenticationMethod(c echo.Context) (string, error) {
	method := c.Get(contextAuthMethod)
	if method == nil {
		return "", fmt.Errorf("authentication method not found in context")
	}

	methodStr, ok := method.(string)
	if !ok {
		return "", fmt.Errorf("invalid authentication method type in context")
	}

	return methodStr, nil
}

func (hb *HttpBasicAuthentication) RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if this handler can handle the request
			if !hb.CanHandle(c) {
				return hb.CreateErrorResponse(http.StatusUnauthorized, errMsgAuthRequired)
			}

			result, err := hb.Authenticate(c)
			if err != nil {
				hb.res.Logger.Debug("Basic Authentication failed",
					zap.String("handler", hb.GetName()),
					zap.String("error", err.Error()))

				// For Basic auth, we should return the WWW-Authenticate header
				httpErr := hb.CreateErrorResponse(http.StatusUnauthorized, errMsgInvalidBasicAuth)
				httpErr.SetInternal(&echo.HTTPError{
					Code:    http.StatusUnauthorized,
					Message: errMsgInvalidBasicAuth,
				})

				// Add WWW-Authenticate header for a basic auth challenge
				c.Response().Header().Set("WWW-Authenticate", "Basic realm=\"Restricted\"")
				return httpErr
			}

			if !result.Success {
				c.Response().Header().Set("WWW-Authenticate", "Basic realm=\"Restricted\"")
				return hb.CreateErrorResponse(http.StatusUnauthorized, errMsgInvalidBasicAuth)
			}

			// Set user context from authentication result
			hb.SetUserContext(c, result)
			return next(c)
		}
	}
}

func (hb *HttpBasicAuthentication) RequireRole(requiredRole string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get user role from context
			roleInterface := c.Get(contextRole)
			if roleInterface == nil {
				return hb.CreateErrorResponse(http.StatusForbidden, "Access denied: no role found")
			}

			userRole, ok := roleInterface.(string)
			if !ok {
				return hb.CreateErrorResponse(http.StatusForbidden, "Access denied: invalid role")
			}

			// Check if a user has the required role
			if !hb.HasRequiredRole(userRole, requiredRole) {
				return hb.CreateErrorResponse(http.StatusForbidden, "Access denied: insufficient permissions")
			}

			return next(c)
		}
	}
}

func (hb *HttpBasicAuthentication) Authenticate(c echo.Context) (*AuthenticationResult, error) {
	username, password, err := hb.extractCredentials(c)
	if err != nil {
		return nil, err
	}

	validPassword, exists := hb.validCredentials[username]
	if !exists {
		return nil, fmt.Errorf(errMsgInvalidBasicAuth)
	}

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(password), []byte(validPassword)) != 1 {
		return nil, fmt.Errorf(errMsgInvalidBasicAuth)
	}

	// Generate a UUID for a basic auth user (in production, get from a database)
	userUUID := uuid.New()
	usernamePtr := &username

	adminRole := "ADMIN"
	return &AuthenticationResult{
		Success:  true,
		UserID:   &userUUID,
		Username: usernamePtr,
		Role:     &adminRole, // Basic auth users get admin role
		Method:   authMethodBasic,
	}, nil
}

func (hb *HttpBasicAuthentication) SetUserContext(c echo.Context, result *AuthenticationResult) {
	if result.UserID != nil {
		c.Set(contextUserID, *result.UserID)
		c.Set(contextUserUUID, *result.UserID)
	}
	if result.Username != nil {
		c.Set(contextUsername, *result.Username)
	}
	if result.Email != nil {
		c.Set(contextEmail, *result.Email)
	}
	if result.PhoneNumber != nil {
		c.Set(contextPhoneNumber, *result.PhoneNumber)
	}
	if result.Role != nil {
		c.Set(contextRole, *result.Role)
	}
	if result.EmailVerified != nil {
		c.Set(contextEmailVerified, *result.EmailVerified)
	}
	if result.PhoneVerified != nil {
		c.Set(contextPhoneVerified, *result.PhoneVerified)
	}
	if result.LastLoginAt != nil {
		c.Set(contextLastLoginAt, *result.LastLoginAt)
	}

	// Set an authentication method and service name for tracking
	c.Set(contextAuthMethod, result.Method)
	if result.ServiceName != nil {
		c.Set(contextServiceName, *result.ServiceName)
	}
}

func (hb *HttpBasicAuthentication) CreateErrorResponse(statusCode int, message string) *echo.HTTPError {
	return echo.NewHTTPError(statusCode, response.ToErrorResponse(statusCode, message))

}

func (hb *HttpBasicAuthentication) HasRequiredRole(userRole string, requiredRole string) bool {
	return userRole == requiredRole
}

func (hb *HttpBasicAuthentication) extractCredentials(c echo.Context) (string, string, error) {
	authHeader := c.Request().Header.Get(authHeaderName)
	if authHeader == "" {
		return "", "", fmt.Errorf(errMsgHeaderMissing)
	}

	parts := strings.SplitN(authHeader, " ", tokenParts)
	if len(parts) != tokenParts || strings.ToLower(parts[0]) != basicPrefix {
		return "", "", fmt.Errorf(errMsgInvalidHeaderFormat)
	}

	payload, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", fmt.Errorf(errMsgInvalidBasicAuth)
	}

	credentials := strings.SplitN(string(payload), ":", 2)
	if len(credentials) != 2 {
		return "", "", fmt.Errorf(errMsgInvalidBasicAuth)
	}

	return credentials[0], credentials[1], nil
}

// IsBasicAuthenticated checks if the current request used basic authentication
func (hb *HttpBasicAuthentication) IsBasicAuthenticated(c echo.Context) bool {
	method, err := hb.GetAuthenticationMethod(c)
	return err == nil && method == authMethodBasic
}
