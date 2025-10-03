package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"

	"backend/service-platform/app/api/middleware"
	"backend/service-platform/app/internal/config"
	"backend/service-platform/app/internal/runtime"
	jwtPkg "backend/service-platform/app/pkg/jwt"

	"go.uber.org/zap"
)

type JwtAuthenticationSuite struct {
	suite.Suite
	jwtAuth     middleware.JwtAuthentication
	echo        *echo.Echo
	req         *http.Request
	rec         *httptest.ResponseRecorder
	ctx         echo.Context
	res         runtime.Resource
	testUserID  uuid.UUID
	jwtInstance jwtPkg.Jwt
	secretKey   string
}

func TestJwtAuthenticationSuite(t *testing.T) {
	suite.Run(t, new(JwtAuthenticationSuite))
}

func (s *JwtAuthenticationSuite) SetupSuite() {
	s.testUserID = uuid.New()
	s.secretKey = "test-secret-key-for-jwt-testing-12345"

	logger, _ := zap.NewDevelopment()
	s.res = runtime.Resource{
		Config: config.ApplicationConfig{
			JwtConfig: config.JwtConfig{
				SecretKey:        s.secretKey,
				AccessExpiration: 1 * time.Hour,
			},
		},
		Logger: logger,
	}

	s.jwtInstance = jwtPkg.NewJwt(s.res.Config.JwtConfig)
}

func (s *JwtAuthenticationSuite) SetupTest() {
	s.echo = echo.New()
	s.req = httptest.NewRequest(http.MethodGet, "/", nil)
	s.rec = httptest.NewRecorder()
	s.ctx = s.echo.NewContext(s.req, s.rec)

	s.jwtAuth = middleware.NewJwtAuthentication(s.res)
}

// Helper function to create a valid JWT token
func (s *JwtAuthenticationSuite) createValidToken(userID *uuid.UUID, username *string, email *string, role *string) string {
	phoneNumber := "+1234567890"
	emailVerified := true
	phoneVerified := false
	lastLoginAt := time.Now()

	claims := &jwtPkg.Claims{
		UserID:        userID,
		Username:      username,
		Email:         email,
		PhoneNumber:   &phoneNumber,
		Role:          role,
		EmailVerified: &emailVerified,
		PhoneVerified: &phoneVerified,
		LastLoginAt:   &lastLoginAt,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(s.secretKey))
	return tokenString
}

// Helper function to create an expired JWT token
func (s *JwtAuthenticationSuite) createExpiredToken() string {
	userID := s.testUserID
	username := "testuser"
	email := "test@example.com"
	phoneNumber := "+1234567890"
	roleStr := "USER"
	emailVerified := true
	phoneVerified := false
	lastLoginAt := time.Now()

	claims := &jwtPkg.Claims{
		UserID:        &userID,
		Username:      &username,
		Email:         &email,
		PhoneNumber:   &phoneNumber,
		Role:          &roleStr,
		EmailVerified: &emailVerified,
		PhoneVerified: &phoneVerified,
		LastLoginAt:   &lastLoginAt,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(s.secretKey))
	return tokenString
}

// CanHandle Tests

func (s *JwtAuthenticationSuite) TestCanHandle_ValidBearerToken() {
	// Arrange
	s.req.Header.Set("Authorization", "Bearer valid_token_123")

	// Act
	result := s.jwtAuth.CanHandle(s.ctx)

	// Assert
	s.True(result)
}

func (s *JwtAuthenticationSuite) TestCanHandle_ValidBearerTokenLowercase() {
	// Arrange
	s.req.Header.Set("Authorization", "bearer valid_token_123")

	// Act
	result := s.jwtAuth.CanHandle(s.ctx)

	// Assert
	s.True(result)
}

func (s *JwtAuthenticationSuite) TestCanHandle_NoAuthHeader() {
	// Arrange - no authorization header

	// Act
	result := s.jwtAuth.CanHandle(s.ctx)

	// Assert
	s.False(result)
}

func (s *JwtAuthenticationSuite) TestCanHandle_EmptyAuthHeader() {
	// Arrange
	s.req.Header.Set("Authorization", "")

	// Act
	result := s.jwtAuth.CanHandle(s.ctx)

	// Assert
	s.False(result)
}

func (s *JwtAuthenticationSuite) TestCanHandle_InvalidTokenFormat() {
	// Arrange
	s.req.Header.Set("Authorization", "Invalid token_123")

	// Act
	result := s.jwtAuth.CanHandle(s.ctx)

	// Assert
	s.False(result)
}

func (s *JwtAuthenticationSuite) TestCanHandle_OnlyBearer() {
	// Arrange
	s.req.Header.Set("Authorization", "Bearer")

	// Act
	result := s.jwtAuth.CanHandle(s.ctx)

	// Assert
	s.False(result)
}

func (s *JwtAuthenticationSuite) TestCanHandle_BasicAuth() {
	// Arrange
	s.req.Header.Set("Authorization", "Basic dGVzdDp0ZXN0")

	// Act
	result := s.jwtAuth.CanHandle(s.ctx)

	// Assert
	s.False(result)
}

// Authenticate Tests

func (s *JwtAuthenticationSuite) TestAuthenticate_Success() {
	// Arrange
	username := "testuser"
	email := "test@example.com"
	userRole := "USER"
	validToken := s.createValidToken(&s.testUserID, &username, &email, &userRole)
	s.req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", validToken))

	// Act
	result, err := s.jwtAuth.Authenticate(s.ctx)

	// Assert
	s.NoError(err)
	s.NotNil(result)
	s.True(result.Success)
	s.Equal(s.testUserID, *result.UserID)
	s.Equal("testuser", *result.Username)
	s.Equal("test@example.com", *result.Email)
	s.Equal("USER", *result.Role)
	s.Equal("jwt", result.Method)
}

func (s *JwtAuthenticationSuite) TestAuthenticate_InvalidToken() {
	// Arrange
	s.req.Header.Set("Authorization", "Bearer invalid_token_signature")

	// Act
	result, err := s.jwtAuth.Authenticate(s.ctx)

	// Assert
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "Invalid credentials")
}

func (s *JwtAuthenticationSuite) TestAuthenticate_ExpiredToken() {
	// Arrange
	expiredToken := s.createExpiredToken()
	s.req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", expiredToken))

	// Act
	result, err := s.jwtAuth.Authenticate(s.ctx)

	// Assert
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "Invalid credentials")
}

func (s *JwtAuthenticationSuite) TestAuthenticate_NoAuthHeader() {
	// Arrange - no authorization header

	// Act
	result, err := s.jwtAuth.Authenticate(s.ctx)

	// Assert
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "Authorization header missing")
}

func (s *JwtAuthenticationSuite) TestAuthenticate_InvalidHeaderFormat() {
	// Arrange
	s.req.Header.Set("Authorization", "InvalidFormat")

	// Act
	result, err := s.jwtAuth.Authenticate(s.ctx)

	// Assert
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "Invalid authorization header format")
}

func (s *JwtAuthenticationSuite) TestAuthenticate_EmptyToken() {
	// Arrange
	s.req.Header.Set("Authorization", "Bearer ")

	// Act
	result, err := s.jwtAuth.Authenticate(s.ctx)

	// Assert
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "Invalid authorization header format")
}

func (s *JwtAuthenticationSuite) TestAuthenticate_WithoutRoles() {
	// Arrange
	username := "testuser"
	email := "test@example.com"
	validToken := s.createValidToken(&s.testUserID, &username, &email, nil)
	s.req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", validToken))

	// Act
	result, err := s.jwtAuth.Authenticate(s.ctx)

	// Assert
	s.NoError(err)
	s.NotNil(result)
	s.True(result.Success)
	s.Nil(result.Role)
}

// RequireAuth Tests

func (s *JwtAuthenticationSuite) TestRequireAuth_Success() {
	// Arrange
	username := "testuser"
	email := "test@example.com"
	userRole := "USER"
	validToken := s.createValidToken(&s.testUserID, &username, &email, &userRole)
	s.req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", validToken))

	nextCalled := false
	next := func(c echo.Context) error {
		nextCalled = true
		return nil
	}

	middleware := s.jwtAuth.RequireAuth()

	// Act
	err := middleware(next)(s.ctx)

	// Assert
	s.NoError(err)
	s.True(nextCalled)
	s.Equal(s.testUserID.String(), s.ctx.Get("user_id"))
	s.Equal(s.testUserID, s.ctx.Get("user_uuid"))
	s.Equal("testuser", s.ctx.Get("username"))
	s.Equal("test@example.com", s.ctx.Get("email"))
	s.Equal("USER", s.ctx.Get("role"))
	s.Equal("jwt", s.ctx.Get("auth_method"))
}

func (s *JwtAuthenticationSuite) TestRequireAuth_Unauthorized() {
	// Arrange - no authorization header
	nextCalled := false
	next := func(c echo.Context) error {
		nextCalled = true
		return nil
	}

	middleware := s.jwtAuth.RequireAuth()

	// Act
	err := middleware(next)(s.ctx)

	// Assert
	s.Error(err)
	s.False(nextCalled)

	httpErr, ok := err.(*echo.HTTPError)
	s.True(ok)
	s.Equal(http.StatusUnauthorized, httpErr.Code)
}

func (s *JwtAuthenticationSuite) TestRequireAuth_InvalidToken() {
	// Arrange
	s.req.Header.Set("Authorization", "Bearer invalid_token")

	nextCalled := false
	next := func(c echo.Context) error {
		nextCalled = true
		return nil
	}

	middleware := s.jwtAuth.RequireAuth()

	// Act
	err := middleware(next)(s.ctx)

	// Assert
	s.Error(err)
	s.False(nextCalled)

	httpErr, ok := err.(*echo.HTTPError)
	s.True(ok)
	s.Equal(http.StatusUnauthorized, httpErr.Code)
}

// RequireRole Tests

func (s *JwtAuthenticationSuite) TestRequireRole_Success() {
	// Arrange
	username := "testuser"
	email := "test@example.com"
	roles := "ADMIN"
	validToken := s.createValidToken(&s.testUserID, &username, &email, &roles)
	s.req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", validToken))

	nextCalled := false
	next := func(c echo.Context) error {
		nextCalled = true
		return nil
	}

	middleware := s.jwtAuth.RequireRole("ADMIN")

	// Act
	err := middleware(next)(s.ctx)

	// Assert
	s.NoError(err)
	s.True(nextCalled)
}

func (s *JwtAuthenticationSuite) TestRequireRole_InsufficientPermissions() {
	// Arrange
	username := "testuser"
	email := "test@example.com"
	roles := "USER"
	validToken := s.createValidToken(&s.testUserID, &username, &email, &roles)
	s.req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", validToken))

	nextCalled := false
	next := func(c echo.Context) error {
		nextCalled = true
		return nil
	}

	middleware := s.jwtAuth.RequireRole("ADMIN")

	// Act
	err := middleware(next)(s.ctx)

	// Assert
	s.Error(err)
	s.False(nextCalled)

	httpErr, ok := err.(*echo.HTTPError)
	s.True(ok)
	s.Equal(http.StatusForbidden, httpErr.Code)
}

func (s *JwtAuthenticationSuite) TestRequireRole_NoRolesInToken() {
	// Arrange
	username := "testuser"
	email := "test@example.com"
	validToken := s.createValidToken(&s.testUserID, &username, &email, nil)
	s.req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", validToken))

	nextCalled := false
	next := func(c echo.Context) error {
		nextCalled = true
		return nil
	}

	middleware := s.jwtAuth.RequireRole("USER")

	// Act
	err := middleware(next)(s.ctx)

	// Assert
	s.Error(err)
	s.False(nextCalled)

	httpErr, ok := err.(*echo.HTTPError)
	s.True(ok)
	s.Equal(http.StatusUnauthorized, httpErr.Code)
}

func (s *JwtAuthenticationSuite) TestRequireRole_EmptyRolesInToken() {
	// Arrange
	username := "testuser"
	email := "test@example.com"
	emptyRoles := ""
	validToken := s.createValidToken(&s.testUserID, &username, &email, &emptyRoles)
	s.req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", validToken))

	nextCalled := false
	next := func(c echo.Context) error {
		nextCalled = true
		return nil
	}

	middleware := s.jwtAuth.RequireRole("USER")

	// Act
	err := middleware(next)(s.ctx)

	// Assert
	s.Error(err)
	s.False(nextCalled)

	httpErr, ok := err.(*echo.HTTPError)
	s.True(ok)
	s.Equal(http.StatusUnauthorized, httpErr.Code)
}

func (s *JwtAuthenticationSuite) TestRequireRole_MultipleRequiredRoles() {
	// Arrange
	username := "testuser"
	email := "test@example.com"
	roles := "USER"
	validToken := s.createValidToken(&s.testUserID, &username, &email, &roles)
	s.req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", validToken))

	nextCalled := false
	next := func(c echo.Context) error {
		nextCalled = true
		return nil
	}

	middleware := s.jwtAuth.RequireRole("ADMIN")

	// Act
	err := middleware(next)(s.ctx)

	// Assert
	s.Error(err)
	s.False(nextCalled)

	// Verify it's a 403 Forbidden error
	httpErr, ok := err.(*echo.HTTPError)
	s.True(ok)
	s.Equal(http.StatusForbidden, httpErr.Code)
}

func (s *JwtAuthenticationSuite) TestRequireRole_InvalidToken() {
	// Arrange
	s.req.Header.Set("Authorization", "Bearer invalid_token")

	nextCalled := false
	next := func(c echo.Context) error {
		nextCalled = true
		return nil
	}

	middleware := s.jwtAuth.RequireRole("USER")

	// Act
	err := middleware(next)(s.ctx)

	// Assert
	s.Error(err)
	s.False(nextCalled)

	httpErr, ok := err.(*echo.HTTPError)
	s.True(ok)
	s.Equal(http.StatusUnauthorized, httpErr.Code)
}

func (s *JwtAuthenticationSuite) TestRequireRole_NoAuthHeader() {
	// Arrange - no authorization header
	nextCalled := false
	next := func(c echo.Context) error {
		nextCalled = true
		return nil
	}

	middleware := s.jwtAuth.RequireRole("USER")

	// Act
	err := middleware(next)(s.ctx)

	// Assert
	s.Error(err)
	s.False(nextCalled)

	httpErr, ok := err.(*echo.HTTPError)
	s.True(ok)
	s.Equal(http.StatusUnauthorized, httpErr.Code)
}

// HasRequiredRole Tests

func (s *JwtAuthenticationSuite) TestHasRequiredRole_SingleRoleMatch() {
	// Arrange
	userRole := "ADMIN"
	requiredRole := "ADMIN"

	// Act
	result := s.jwtAuth.HasRequiredRole(userRole, requiredRole)

	// Assert
	s.True(result)
}

func (s *JwtAuthenticationSuite) TestHasRequiredRole_NoMatch() {
	// Arrange
	userRole := "USER"
	requiredRole := "ADMIN"

	// Act
	result := s.jwtAuth.HasRequiredRole(userRole, requiredRole)

	// Assert
	s.False(result)
}

func (s *JwtAuthenticationSuite) TestHasRequiredRole_EmptyUserRoles() {
	// Arrange
	userRole := ""
	requiredRole := "USER"

	// Act
	result := s.jwtAuth.HasRequiredRole(userRole, requiredRole)

	// Assert
	s.False(result)
}

func (s *JwtAuthenticationSuite) TestHasRequiredRole_EmptyRequiredRoles() {
	// Arrange
	userRole := "USER"
	requiredRole := ""

	// Act
	result := s.jwtAuth.HasRequiredRole(userRole, requiredRole)

	// Assert
	s.False(result)
}

func (s *JwtAuthenticationSuite) TestHasRequiredRole_MultipleRequired_OneMatch() {
	// Arrange
	userRole := "USER"
	requiredRole := "USER"

	// Act
	result := s.jwtAuth.HasRequiredRole(userRole, requiredRole)

	// Assert
	s.True(result)
}

func (s *JwtAuthenticationSuite) TestHasRequiredRole_ManyRequiredRoles() {
	// Arrange
	userRole := "USER"
	requiredRole := "USER"

	// Act
	result := s.jwtAuth.HasRequiredRole(userRole, requiredRole)

	// Assert
	s.True(result)
}

// GetName Tests

func (s *JwtAuthenticationSuite) TestGetName() {
	// Act
	name := s.jwtAuth.GetName()

	// Assert
	s.Equal("jwt", name)
}

// IsJWTAuthenticated Tests

func (s *JwtAuthenticationSuite) TestIsJWTAuthenticated_True() {
	// Arrange
	s.ctx.Set("auth_method", "jwt")

	// Act
	result := s.jwtAuth.IsJWTAuthenticated(s.ctx)

	// Assert
	s.True(result)
}

func (s *JwtAuthenticationSuite) TestIsJWTAuthenticated_False() {
	// Arrange
	s.ctx.Set("auth_method", "api_key")

	// Act
	result := s.jwtAuth.IsJWTAuthenticated(s.ctx)

	// Assert
	s.False(result)
}

func (s *JwtAuthenticationSuite) TestIsJWTAuthenticated_NoMethod() {
	// Arrange - no auth method set

	// Act
	result := s.jwtAuth.IsJWTAuthenticated(s.ctx)

	// Assert
	s.False(result)
}

// CreateErrorResponse Tests

func (s *JwtAuthenticationSuite) TestCreateErrorResponse() {
	// Act
	httpErr := s.jwtAuth.CreateErrorResponse(http.StatusUnauthorized, "Test error message")

	// Assert
	s.NotNil(httpErr)
	s.Equal(http.StatusUnauthorized, httpErr.Code)

	// Check if the message is properly formatted - it should be a response object
	// The actual response structure depends on the CreateErrorResponse implementation
}

// Integration test combining multiple methods

func (s *JwtAuthenticationSuite) TestIntegration_FullAuthenticationFlow() {
	// Arrange
	username := "integrationuser"
	email := "integration@example.com"
	userRole := "USER"
	validToken := s.createValidToken(&s.testUserID, &username, &email, &userRole)

	s.req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", validToken))

	// Test CanHandle
	s.True(s.jwtAuth.CanHandle(s.ctx))

	// Test Authenticate
	result, err := s.jwtAuth.Authenticate(s.ctx)
	s.NoError(err)
	s.True(result.Success)

	// Test SetUserContext
	s.jwtAuth.SetUserContext(s.ctx, result)
	s.Equal(s.testUserID.String(), s.ctx.Get("user_id"))
	s.Equal("integrationuser", s.ctx.Get("username"))
	s.Equal("integration@example.com", s.ctx.Get("email"))
	s.Equal("USER", s.ctx.Get("role"))

	// Test IsJWTAuthenticated
	s.True(s.jwtAuth.IsJWTAuthenticated(s.ctx))

	// Test HasRequiredRole
	s.True(s.jwtAuth.HasRequiredRole(*result.Role, "USER"))
	s.False(s.jwtAuth.HasRequiredRole(*result.Role, "ADMIN"))
}
