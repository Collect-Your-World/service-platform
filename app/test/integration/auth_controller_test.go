package integration

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"backend/service-platform/app/internal/auth/constants/role"
	authmanagers "backend/service-platform/app/internal/auth/managers"
	authmodels "backend/service-platform/app/internal/auth/models"
	commonmodels "backend/service-platform/app/internal/common/models"
	mocks "backend/service-platform/app/test/mocks/managers"
	httputil "backend/service-platform/app/test/util"

	"backend/service-platform/app/pkg/jwt"

	"github.com/google/uuid"
)

const (
	LoginEndpoint        = "/api/v1/auth/login"
	RegisterEndpoint     = "/api/v1/auth/register"
	LogoutEndpoint       = "/api/v1/auth/logout"
	RefreshTokenEndpoint = "/api/v1/auth/refresh-token"
	MeEndpoint           = "/api/v1/auth/me"
)

type AuthControllerSuite struct {
	RouterSuite
}

func TestAuthControllerSuite(t *testing.T) {
	suite.Run(t, new(AuthControllerSuite))
}

// Register Tests

func (s *AuthControllerSuite) TestRegister_Success() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.managers.AuthManager = m

	req := authmodels.RegisterRequest{
		Email:    "newuser@example.com",
		Password: "password123",
	}

	m.EXPECT().Register(mock.Anything, req).Return(nil)

	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[string]](
		s.e,
		http.MethodPost,
		RegisterEndpoint,
		nil,
		req,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, code)
	s.r.Equal("success", resp.Message)
	s.r.Equal("registered", resp.Data)
}

func (s *AuthControllerSuite) TestRegister_EmailAlreadyExists() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.managers.AuthManager = m

	req := authmodels.RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
	}

	m.EXPECT().Register(mock.Anything, req).Return(authmanagers.ErrEmailAlreadyExists)

	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.e,
		http.MethodPost,
		RegisterEndpoint,
		nil,
		req,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusConflict, code)
	s.r.Equal(http.StatusConflict, resp.Code)
	s.r.Equal("email already exists", resp.Message)
}

func (s *AuthControllerSuite) TestRegister_InvalidEmail() {
	// Arrange
	req := authmodels.RegisterRequest{
		Email:    "invalid-email",
		Password: "password123",
	}

	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.e,
		http.MethodPost,
		RegisterEndpoint,
		nil,
		req,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusBadRequest, code)
	s.r.Equal(http.StatusBadRequest, resp.Code)
	s.r.Equal("Invalid data", resp.Message)
}

func (s *AuthControllerSuite) TestRegister_ShortPassword() {
	// Arrange
	req := authmodels.RegisterRequest{
		Email:    "user@example.com",
		Password: "short",
	}

	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.e,
		http.MethodPost,
		RegisterEndpoint,
		nil,
		req,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusBadRequest, code)
	s.r.Equal(http.StatusBadRequest, resp.Code)
	s.r.Equal("Invalid data", resp.Message)
}

// Login Tests

func (s *AuthControllerSuite) TestLogin_Success() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.managers.AuthManager = m

	req := authmodels.AuthUserRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	username := "test@example.com"
	expectedResponse := &authmodels.AuthResponse{
		Username:     &username,
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_456",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	}

	m.EXPECT().Login(mock.Anything, req).Return(expectedResponse, nil)

	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[authmodels.AuthResponse]](
		s.e,
		http.MethodPost,
		LoginEndpoint,
		nil,
		req,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, code)
	s.r.Equal("success", resp.Message)
	s.r.Equal("test@example.com", *resp.Data.Username)
	s.r.Equal("access_token_123", resp.Data.AccessToken)
	s.r.Equal("", resp.Data.RefreshToken)
	// Cookie set
	c := httputil.GetCookie("refresh_token")
	s.Require().NotNil(c)
	s.r.Equal("refresh_token_456", c.Value)
	s.r.True(c.Expires.After(time.Now()))
	s.r.Greater(c.MaxAge, 0)
	s.r.Equal(int64(3600), resp.Data.ExpiresIn)
	s.r.Equal("Bearer", resp.Data.TokenType)
}

func (s *AuthControllerSuite) TestLogin_InvalidCredentials() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.managers.AuthManager = m

	req := authmodels.AuthUserRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	expectedError := authmanagers.ErrInvalidCredentials
	m.EXPECT().Login(mock.Anything, req).Return(nil, expectedError)

	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.e,
		http.MethodPost,
		LoginEndpoint,
		nil,
		req,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusUnauthorized, code)
	s.r.Equal(http.StatusUnauthorized, resp.Code)
	s.r.Equal("Invalid credentials", resp.Message)
}

func (s *AuthControllerSuite) TestLogin_InvalidRequestBody() {
	// Arrange
	invalidReq := map[string]interface{}{
		"invalid_field": "value",
	}

	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.e,
		http.MethodPost,
		LoginEndpoint,
		nil,
		invalidReq,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusBadRequest, code)
	s.r.Equal(http.StatusBadRequest, resp.Code)
	s.r.Equal("Invalid request data", resp.Message)
}

func (s *AuthControllerSuite) TestLogin_EmptyEmail() {
	// Arrange
	req := authmodels.AuthUserRequest{
		Email:    "",
		Password: "password123",
	}

	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.e,
		http.MethodPost,
		LoginEndpoint,
		nil,
		req,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusBadRequest, code)
	s.r.Equal(http.StatusBadRequest, resp.Code)
	s.r.Equal("Invalid request data", resp.Message)
}

func (s *AuthControllerSuite) TestLogin_EmptyPassword() {
	// Arrange
	req := authmodels.AuthUserRequest{
		Email:    "test@example.com",
		Password: "",
	}

	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.e,
		http.MethodPost,
		LoginEndpoint,
		nil,
		req,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusBadRequest, code)
	s.r.Equal(http.StatusBadRequest, resp.Code)
	s.r.Equal("Invalid request data", resp.Message)
}

func (s *AuthControllerSuite) TestLogin_DatabaseError() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.managers.AuthManager = m

	req := authmodels.AuthUserRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	expectedError := errors.New("database connection failed")
	m.EXPECT().Login(mock.Anything, req).Return(nil, expectedError)

	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.e,
		http.MethodPost,
		LoginEndpoint,
		nil,
		req,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusInternalServerError, code)
	s.r.Equal(http.StatusInternalServerError, resp.Code)
	s.r.Equal("Internal server error", resp.Message)
}

// RefreshToken Tests

func (s *AuthControllerSuite) TestRefreshToken_Success() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.managers.AuthManager = m

	req := authmodels.RefreshTokenRequest{
		RefreshToken: "valid_refresh_token",
	}

	username := "test@example.com"
	expectedResponse := &authmodels.AuthResponse{
		Username:     &username,
		AccessToken:  "new_access_token_123",
		RefreshToken: "new_refresh_token_456",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	}

	m.EXPECT().RefreshToken(mock.Anything, req).Return(expectedResponse, nil)

	// Seed cookie as client would send
	httputil.SetCookie("refresh_token", req.RefreshToken, 3600)
	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[authmodels.AuthResponse]](
		s.e,
		http.MethodPost,
		RefreshTokenEndpoint,
		nil,
		req,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, code)
	s.r.Equal("success", resp.Message)
	s.r.Equal("test@example.com", *resp.Data.Username)
	s.r.Equal("new_access_token_123", resp.Data.AccessToken)
	s.r.Equal("", resp.Data.RefreshToken)
	// Rotated cookie
	c := httputil.GetCookie("refresh_token")
	s.Require().NotNil(c)
	s.r.Equal("new_refresh_token_456", c.Value)
	s.r.True(c.Expires.After(time.Now()))
	s.r.Greater(c.MaxAge, 0)
}

func (s *AuthControllerSuite) TestRefreshToken_InvalidToken() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.managers.AuthManager = m

	req := authmodels.RefreshTokenRequest{
		RefreshToken: "invalid_refresh_token",
	}

	expectedError := authmanagers.ErrInvalidRefreshToken
	m.EXPECT().RefreshToken(mock.Anything, req).Return(nil, expectedError)

	// Seed cookie
	httputil.SetCookie("refresh_token", req.RefreshToken, 3600)
	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.e,
		http.MethodPost,
		RefreshTokenEndpoint,
		nil,
		req,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusUnauthorized, code)
	s.r.Equal(http.StatusUnauthorized, resp.Code)
	s.r.Equal(authmanagers.ErrInvalidRefreshToken.Error(), resp.Message)
}

func (s *AuthControllerSuite) TestRefreshToken_RevokedToken() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.managers.AuthManager = m

	req := authmodels.RefreshTokenRequest{
		RefreshToken: "revoked_refresh_token",
	}

	expectedError := authmanagers.ErrRefreshTokenRevoked
	m.EXPECT().RefreshToken(mock.Anything, req).Return(nil, expectedError)

	// Seed cookie
	httputil.SetCookie("refresh_token", req.RefreshToken, 3600)
	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.e,
		http.MethodPost,
		RefreshTokenEndpoint,
		nil,
		req,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusUnauthorized, code)
	s.r.Equal(http.StatusUnauthorized, resp.Code)
	s.r.Equal(authmanagers.ErrRefreshTokenRevoked.Error(), resp.Message)
}

func (s *AuthControllerSuite) TestRefreshToken_ExpiredToken() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.managers.AuthManager = m

	req := authmodels.RefreshTokenRequest{
		RefreshToken: "expired_refresh_token",
	}

	expectedError := authmanagers.ErrRefreshTokenExpired
	m.EXPECT().RefreshToken(mock.Anything, req).Return(nil, expectedError)

	// Seed cookie
	httputil.SetCookie("refresh_token", req.RefreshToken, 1)
	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.e,
		http.MethodPost,
		RefreshTokenEndpoint,
		nil,
		req,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusUnauthorized, code)
	s.r.Equal(http.StatusUnauthorized, resp.Code)
	s.r.Equal(authmanagers.ErrRefreshTokenExpired.Error(), resp.Message)
}

func (s *AuthControllerSuite) TestRefreshToken_EmptyToken() {
	// Arrange
	req := authmodels.RefreshTokenRequest{
		RefreshToken: "",
	}

	// Ensure no cookie
	httputil.ClearCookies()
	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.e,
		http.MethodPost,
		RefreshTokenEndpoint,
		nil,
		req,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusUnauthorized, code)
	s.r.Equal(http.StatusUnauthorized, resp.Code)
	s.r.Equal("Missing refresh token", resp.Message)
}

func (s *AuthControllerSuite) TestRefreshToken_DatabaseError() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.managers.AuthManager = m

	req := authmodels.RefreshTokenRequest{
		RefreshToken: "valid_refresh_token",
	}

	expectedError := errors.New("database error")
	m.EXPECT().RefreshToken(mock.Anything, req).Return(nil, expectedError)

	// Seed cookie
	httputil.SetCookie("refresh_token", req.RefreshToken, 3600)
	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.e,
		http.MethodPost,
		RefreshTokenEndpoint,
		nil,
		req,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusInternalServerError, code)
	s.r.Equal(http.StatusInternalServerError, resp.Code)
	s.r.Equal("Internal server error", resp.Message)
}

// Logout Tests

func (s *AuthControllerSuite) TestLogout_Success() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.managers.AuthManager = m

	req := authmodels.LogoutRequest{
		RefreshToken: "valid_refresh_token",
	}

	m.EXPECT().Logout(mock.Anything, req).Return(nil)

	// Seed cookie
	httputil.SetCookie("refresh_token", req.RefreshToken, 3600)
	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[string]](
		s.e,
		http.MethodPost,
		LogoutEndpoint,
		nil,
		req,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, code)
	s.r.Equal("success", resp.Message)
	s.r.Equal("Logged out successfully", resp.Data)
}

func (s *AuthControllerSuite) TestLogout_InvalidToken() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.managers.AuthManager = m

	req := authmodels.LogoutRequest{
		RefreshToken: "invalid_refresh_token",
	}

	expectedError := errors.New("invalid token")
	m.EXPECT().Logout(mock.Anything, req).Return(expectedError)

	// Seed cookie
	httputil.SetCookie("refresh_token", req.RefreshToken, 3600)
	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.e,
		http.MethodPost,
		LogoutEndpoint,
		nil,
		req,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusInternalServerError, code)
	s.r.Equal(http.StatusInternalServerError, resp.Code)
	s.r.Equal("Internal server error", resp.Message)
}

func (s *AuthControllerSuite) TestLogout_EmptyToken() {
	// Arrange
	req := authmodels.LogoutRequest{
		RefreshToken: "",
	}

	// Ensure no cookie
	httputil.ClearCookies()
	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.e,
		http.MethodPost,
		LogoutEndpoint,
		nil,
		req,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusUnauthorized, code)
	s.r.Equal(http.StatusUnauthorized, resp.Code)
	s.r.Equal("Missing refresh token", resp.Message)
}

func (s *AuthControllerSuite) TestLogout_DatabaseError() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.managers.AuthManager = m

	req := authmodels.LogoutRequest{
		RefreshToken: "valid_refresh_token",
	}

	expectedError := errors.New("database connection failed")
	m.EXPECT().Logout(mock.Anything, req).Return(expectedError)

	// Seed cookie
	httputil.SetCookie("refresh_token", req.RefreshToken, 3600)
	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.e,
		http.MethodPost,
		LogoutEndpoint,
		nil,
		req,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusInternalServerError, code)
	s.r.Equal(http.StatusInternalServerError, resp.Code)
	s.r.Equal("Internal server error", resp.Message)
}

// Me Tests

func (s *AuthControllerSuite) TestMe_Success() {
	// Arrange
	cfg := s.resource.Config
	j := jwt.NewJwt(cfg.JwtConfig)
	userID := uuid.New()
	username := "test@example.com"
	email := "test@example.com"
	phoneNumber := "+1234567890"
	roleStr := "USER"
	emailVerified := true
	phoneVerified := false
	lastLoginAt := time.Now()

	accessToken, err := j.GenerateAccessToken(&userID, &username, &email, &phoneNumber, &roleStr, &emailVerified, &phoneVerified, &lastLoginAt)
	s.r.NoError(err)

	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[authmodels.MeResponse]](
		s.e,
		http.MethodGet,
		MeEndpoint,
		&accessToken.Token,
		nil,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, code)
	s.r.Equal("success", resp.Message)
	s.r.Equal(userID, resp.Data.ID)
	s.r.Equal(username, resp.Data.Username)
	s.r.Equal(email, *resp.Data.Email)
	s.r.Equal(phoneNumber, *resp.Data.PhoneNumber)
	s.r.Equal(role.User, resp.Data.Role)
	s.r.Equal(emailVerified, resp.Data.EmailVerified)
	s.r.Equal(phoneVerified, resp.Data.PhoneVerified)
}

func (s *AuthControllerSuite) TestMe_MissingAuthorization() {
	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.e,
		http.MethodGet,
		MeEndpoint,
		nil,
		nil,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusUnauthorized, code)
	s.r.Equal(http.StatusUnauthorized, resp.Code)
}

func (s *AuthControllerSuite) TestMe_InvalidToken() {
	// Arrange
	invalidToken := "invalid.jwt.token"

	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.e,
		http.MethodGet,
		MeEndpoint,
		&invalidToken,
		nil,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusUnauthorized, code)
	s.r.Equal(http.StatusUnauthorized, resp.Code)
}

// Edge Cases and Complex Scenarios

func (s *AuthControllerSuite) TestLogin_SpecialCharactersInPassword() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.managers.AuthManager = m

	req := authmodels.AuthUserRequest{
		Email:    "test@example.com",
		Password: "P@ssw0rd!@#$%^&*()",
	}

	username := "test@example.com"
	expectedResponse := &authmodels.AuthResponse{
		Username:     &username,
		AccessToken:  "access_token_special",
		RefreshToken: "refresh_token_special",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	}

	m.EXPECT().Login(mock.Anything, req).Return(expectedResponse, nil)

	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[authmodels.AuthResponse]](
		s.e,
		http.MethodPost,
		LoginEndpoint,
		nil,
		req,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, code)
	s.r.Equal("success", resp.Message)
	s.r.Equal("test@example.com", *resp.Data.Username)
}

func (s *AuthControllerSuite) TestRefreshToken_ValidRequestWithLongToken() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.managers.AuthManager = m

	longToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	req := authmodels.RefreshTokenRequest{
		RefreshToken: longToken,
	}

	username := "test@example.com"
	expectedResponse := &authmodels.AuthResponse{
		Username:     &username,
		AccessToken:  "new_access_token",
		RefreshToken: "new_refresh_token",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	}

	m.EXPECT().RefreshToken(mock.Anything, req).Return(expectedResponse, nil)

	// Seed cookie as client would send
	httputil.SetCookie("refresh_token", longToken, 3600)
	// Act
	resp, code, err := httputil.RequestHTTP[commonmodels.GeneralResponse[authmodels.AuthResponse]](
		s.e,
		http.MethodPost,
		RefreshTokenEndpoint,
		nil,
		req,
	)

	// Assert
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, code)
	s.r.Equal("success", resp.Message)
}
