package middleware

import (
	"backend/service-platform/app/api/client/response"
	"backend/service-platform/app/internal/runtime"
	"backend/service-platform/app/pkg/jwt"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type JwtAuthentication struct {
	jwt jwt.Jwt
	res runtime.Resource
}

func NewJwtAuthentication(res runtime.Resource) JwtAuthentication {
	newJwt := jwt.NewJwt(res.Config.JwtConfig)
	return JwtAuthentication{
		jwt: newJwt,
		res: res,
	}
}

func (j JwtAuthentication) GetName() string {
	return authMethodJWT
}

func (j JwtAuthentication) CanHandle(ec echo.Context) bool {
	authHeader := ec.Request().Header.Get(authHeaderName)
	if authHeader == "" {
		return false
	}

	parts := strings.SplitN(authHeader, " ", tokenParts)
	if len(parts) != tokenParts {
		return false
	}

	return parts[0] == jwt.TokenTypeBearer || strings.ToLower(parts[0]) == strings.ToLower(jwt.TokenTypeBearer)

}

func (j JwtAuthentication) GetAuthenticationMethod(c echo.Context) (string, error) {
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

func (j JwtAuthentication) RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Use this JWT handler for authentication
			if !j.CanHandle(c) {
				return j.CreateErrorResponse(http.StatusUnauthorized, errMsgAuthRequired)
			}

			result, err := j.Authenticate(c)
			if err != nil {
				j.res.Logger.Debug("JWT Authentication failed",
					zap.String("handler", j.GetName()),
					zap.String("error", err.Error()))
				return j.CreateErrorResponse(http.StatusUnauthorized, errMsgInvalidCredentials)
			}

			if !result.Success {
				return j.CreateErrorResponse(http.StatusUnauthorized, errMsgInvalidCredentials)
			}

			// Set user context from authentication result
			j.SetUserContext(c, result)
			return next(c)
		}
	}
}

func (j JwtAuthentication) RequireRole(requiredRole string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract Authorization header
			authHeader := c.Request().Header.Get(authHeaderName)
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return j.CreateErrorResponse(http.StatusUnauthorized, "Missing or invalid Authorization header")
			}

			// Extract the token string
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			// Parse and validate JWT
			token, err := j.jwt.ValidateToken(tokenString)
			if err != nil {
				return j.CreateErrorResponse(http.StatusUnauthorized, "Invalid token")
			}

			// Check if user role is present in the token
			if token.Role == nil || *token.Role == "" {
				return j.CreateErrorResponse(http.StatusUnauthorized, "User role not found in token")
			}

			// Check if a user has the required role
			if !j.HasRequiredRole(*token.Role, requiredRole) {
				return j.CreateErrorResponse(http.StatusForbidden, "Access denied: insufficient permissions")
			}

			return next(c)
		}
	}
}

func (j JwtAuthentication) Authenticate(ec echo.Context) (*AuthenticationResult, error) {
	token, err := j.extractToken(ec)
	if err != nil {
		return nil, err
	}

	claims, err := j.jwt.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf(errMsgInvalidCredentials)
	}

	return &AuthenticationResult{
		Success:       true,
		UserID:        claims.UserID,
		Username:      claims.Username,
		Email:         claims.Email,
		PhoneNumber:   claims.PhoneNumber,
		Role:          claims.Role,
		EmailVerified: claims.EmailVerified,
		PhoneVerified: claims.PhoneVerified,
		LastLoginAt:   claims.LastLoginAt,
		Method:        authMethodJWT,
	}, nil
}

func (j JwtAuthentication) SetUserContext(c echo.Context, result *AuthenticationResult) {
	if result.UserID != nil {
		c.Set(contextUserID, result.UserID.String())
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

func (j JwtAuthentication) CreateErrorResponse(statusCode int, message string) *echo.HTTPError {
	return echo.NewHTTPError(statusCode, response.ToErrorResponse(statusCode, message))
}

func (j JwtAuthentication) HasRequiredRole(userRole string, requiredRole string) bool {
	return userRole == requiredRole
}

func (j *JwtAuthentication) extractToken(ec echo.Context) (string, error) {
	authHeader := ec.Request().Header.Get(authHeaderName)
	if authHeader == "" {
		return "", fmt.Errorf(errMsgHeaderMissing)
	}

	parts := strings.SplitN(authHeader, " ", tokenParts)
	if len(parts) != tokenParts {
		return "", fmt.Errorf(errMsgInvalidHeaderFormat)
	}

	if parts[0] != jwt.TokenTypeBearer && strings.ToLower(parts[0]) != strings.ToLower(jwt.TokenTypeBearer) {
		return "", fmt.Errorf(errMsgInvalidHeaderFormat)
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", fmt.Errorf(errMsgInvalidHeaderFormat)
	}

	return token, nil
}

// IsJWTAuthenticated checks if the current request used JWT authentication
func (j JwtAuthentication) IsJWTAuthenticated(c echo.Context) bool {
	method, err := j.GetAuthenticationMethod(c)
	return err == nil && method == authMethodJWT
}
