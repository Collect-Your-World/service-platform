// Package middleware provides multi-handler authentication and authorization middleware for the HTTP API.
// This package supports multiple authentication mechanisms, including JWT, API keys, and Basic Auth.
//
// Supported Authentication Mechanisms:
// - JWT Bearer tokens (primary handler for user authentication)
// - API Key authentication (for service-to-service communication)
// - Basic Authentication (for legacy system integration)
//
// Key features:
// - Chain of responsibility pattern for trying multiple mechanisms
// - Optimized role checking with map-based lookup for large role sets
// - Centralized error response creation
// - Comprehensive helper functions for common operations
package middleware

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	// Context keys
	contextUserID        = "user_id"
	contextUserUUID      = "user_uuid"
	contextEmail         = "email"
	contextUsername      = "username"
	contextPhoneNumber   = "phone_number"
	contextRole          = "role"
	contextEmailVerified = "email_verified"
	contextPhoneVerified = "phone_verified"
	contextLastLoginAt   = "last_login_at"
	contextAuthMethod    = "auth_method"
	contextServiceName   = "service_name"

	// Authentication methods
	authMethodJWT    = "jwt"
	authMethodAPIKey = "api_key"
	authMethodBasic  = "basic_auth"

	// Token constants
	basicPrefix    = "basic"
	tokenParts     = 2
	authHeaderName = "Authorization"
	apiKeyHeader   = "X-API-Key"

	// Service authentication
	serviceUserID   = "00000000-0000-0000-0000-000000000000"
	serviceUsername = "service"

	// Error messages
	errMsgAuthRequired        = "Authentication required"
	errMsgInvalidCredentials  = "Invalid credentials"
	errMsgHeaderMissing       = "Authorization header missing"
	errMsgInvalidHeaderFormat = "Invalid authorization header format"
	errMsgInvalidAPIKey       = "Invalid API key"
	errMsgInvalidBasicAuth    = "Invalid basic authentication"
)

// AuthenticationResult represents the result of an authentication attempt
type AuthenticationResult struct {
	Success       bool       // Whether authentication was successful
	UserID        *uuid.UUID // Authenticated user ID
	Username      *string    // Authenticated username
	Email         *string    // Authenticated user email
	PhoneNumber   *string    // Authenticated user phone number
	Role          *string    // User role
	EmailVerified *bool      // Email verification status
	PhoneVerified *bool      // Phone verification status
	LastLoginAt   *time.Time // Last login timestamp
	Method        string     // Authentication method used
	ServiceName   *string    // Service name (for API key auth)
}

type Authentication interface {
	// GetName returns the name of this authentication handler
	GetName() string
	// CanHandle returns true if this handler can handle the request
	CanHandle(ec echo.Context) bool
	// GetAuthenticationMethod retrieves the authentication method used for the current request
	GetAuthenticationMethod(c echo.Context) (string, error)
	// RequireAuth validates credentials using any available authentication handler
	RequireAuth() echo.MiddlewareFunc
	// RequireRole validates that the user has the specified role
	// Note: This middleware includes authentication, so you don't need to use RequireAuthentication separately
	RequireRole(requiredRole string) echo.MiddlewareFunc
	// Authenticate performs authentication and returns the result
	Authenticate(ec echo.Context) (*AuthenticationResult, error)
	// SetUserContext efficiently sets all user-related context values from authentication result
	SetUserContext(c echo.Context, result *AuthenticationResult)
	// CreateErrorResponse creates a standardized error response
	CreateErrorResponse(statusCode int, message string) *echo.HTTPError
	// HasRequiredRole checks if a user has the required role
	HasRequiredRole(userRole string, requiredRole string) bool
}
