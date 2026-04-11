package controllers_test

import (
	integrationtest "backend/service-platform/app/internal/integrationtest"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"backend/service-platform/app/internal/auth/constants/role"
	authmanagers "backend/service-platform/app/internal/auth/managers"
	mocks "backend/service-platform/app/internal/auth/managers/mocks"
	authmodels "backend/service-platform/app/internal/auth/models"
	commonmodels "backend/service-platform/app/internal/common/models"
	testutil "backend/service-platform/app/internal/testutil"

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
	integrationtest.RouterSuite
}

func TestAuthControllerSuite(t *testing.T) {
	suite.Run(t, new(AuthControllerSuite))
}

// Register Tests

func (s *AuthControllerSuite) TestRegister_Success() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.Managers.AuthManager = m

	req := authmodels.RegisterRequest{
		Email:    "newuser@example.com",
		Password: "password123",
	}

	m.EXPECT().Register(mock.Anything, req).Return(nil)

	// Act
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[string]](
		s.Echo,
		http.MethodPost,
		RegisterEndpoint,
		nil,
		req,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, code)
	s.R.Equal("success", resp.Message)
	s.R.Equal("registered", resp.Data)
}

func (s *AuthControllerSuite) TestRegister_EmailAlreadyExists() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.Managers.AuthManager = m

	req := authmodels.RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
	}

	m.EXPECT().Register(mock.Anything, req).Return(authmanagers.ErrEmailAlreadyExists)

	// Act
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodPost,
		RegisterEndpoint,
		nil,
		req,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusConflict, code)
	s.R.Equal(http.StatusConflict, resp.Code)
	s.R.Equal("email already exists", resp.Message)
}

func (s *AuthControllerSuite) TestRegister_InvalidEmail() {
	// Arrange
	req := authmodels.RegisterRequest{
		Email:    "invalid-email",
		Password: "password123",
	}

	// Act
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodPost,
		RegisterEndpoint,
		nil,
		req,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusBadRequest, code)
	s.R.Equal(http.StatusBadRequest, resp.Code)
	s.R.Equal("Invalid data", resp.Message)
}

func (s *AuthControllerSuite) TestRegister_ShortPassword() {
	// Arrange
	req := authmodels.RegisterRequest{
		Email:    "user@example.com",
		Password: "short",
	}

	// Act
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodPost,
		RegisterEndpoint,
		nil,
		req,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusBadRequest, code)
	s.R.Equal(http.StatusBadRequest, resp.Code)
	s.R.Equal("Invalid data", resp.Message)
}

// Login Tests

func (s *AuthControllerSuite) TestLogin_Success() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.Managers.AuthManager = m

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
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[authmodels.AuthResponse]](
		s.Echo,
		http.MethodPost,
		LoginEndpoint,
		nil,
		req,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, code)
	s.R.Equal("success", resp.Message)
	s.R.Equal("test@example.com", *resp.Data.Username)
	s.R.Equal("access_token_123", resp.Data.AccessToken)
	s.R.Equal("", resp.Data.RefreshToken)
	// Cookie set
	c := testutil.GetCookie("refresh_token")
	s.Require().NotNil(c)
	s.R.Equal("refresh_token_456", c.Value)
	s.R.True(c.Expires.After(time.Now()))
	s.R.Greater(c.MaxAge, 0)
	s.R.Equal(int64(3600), resp.Data.ExpiresIn)
	s.R.Equal("Bearer", resp.Data.TokenType)
}

func (s *AuthControllerSuite) TestLogin_InvalidCredentials() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.Managers.AuthManager = m

	req := authmodels.AuthUserRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	expectedError := authmanagers.ErrInvalidCredentials
	m.EXPECT().Login(mock.Anything, req).Return(nil, expectedError)

	// Act
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodPost,
		LoginEndpoint,
		nil,
		req,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusUnauthorized, code)
	s.R.Equal(http.StatusUnauthorized, resp.Code)
	s.R.Equal("Invalid credentials", resp.Message)
}

func (s *AuthControllerSuite) TestLogin_InvalidRequestBody() {
	// Arrange
	invalidReq := map[string]interface{}{
		"invalid_field": "value",
	}

	// Act
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodPost,
		LoginEndpoint,
		nil,
		invalidReq,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusBadRequest, code)
	s.R.Equal(http.StatusBadRequest, resp.Code)
	s.R.Equal("Invalid request data", resp.Message)
}

func (s *AuthControllerSuite) TestLogin_EmptyEmail() {
	// Arrange
	req := authmodels.AuthUserRequest{
		Email:    "",
		Password: "password123",
	}

	// Act
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodPost,
		LoginEndpoint,
		nil,
		req,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusBadRequest, code)
	s.R.Equal(http.StatusBadRequest, resp.Code)
	s.R.Equal("Invalid request data", resp.Message)
}

func (s *AuthControllerSuite) TestLogin_EmptyPassword() {
	// Arrange
	req := authmodels.AuthUserRequest{
		Email:    "test@example.com",
		Password: "",
	}

	// Act
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodPost,
		LoginEndpoint,
		nil,
		req,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusBadRequest, code)
	s.R.Equal(http.StatusBadRequest, resp.Code)
	s.R.Equal("Invalid request data", resp.Message)
}

func (s *AuthControllerSuite) TestLogin_DatabaseError() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.Managers.AuthManager = m

	req := authmodels.AuthUserRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	expectedError := errors.New("database connection failed")
	m.EXPECT().Login(mock.Anything, req).Return(nil, expectedError)

	// Act
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodPost,
		LoginEndpoint,
		nil,
		req,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusInternalServerError, code)
	s.R.Equal(http.StatusInternalServerError, resp.Code)
	s.R.Equal("Internal server error", resp.Message)
}

// RefreshToken Tests

func (s *AuthControllerSuite) TestRefreshToken_Success() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.Managers.AuthManager = m

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
	testutil.SetCookie("refresh_token", req.RefreshToken, 3600)
	// Act
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[authmodels.AuthResponse]](
		s.Echo,
		http.MethodPost,
		RefreshTokenEndpoint,
		nil,
		req,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, code)
	s.R.Equal("success", resp.Message)
	s.R.Equal("test@example.com", *resp.Data.Username)
	s.R.Equal("new_access_token_123", resp.Data.AccessToken)
	s.R.Equal("", resp.Data.RefreshToken)
	// Rotated cookie
	c := testutil.GetCookie("refresh_token")
	s.Require().NotNil(c)
	s.R.Equal("new_refresh_token_456", c.Value)
	s.R.True(c.Expires.After(time.Now()))
	s.R.Greater(c.MaxAge, 0)
}

func (s *AuthControllerSuite) TestRefreshToken_InvalidToken() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.Managers.AuthManager = m

	req := authmodels.RefreshTokenRequest{
		RefreshToken: "invalid_refresh_token",
	}

	expectedError := authmanagers.ErrInvalidRefreshToken
	m.EXPECT().RefreshToken(mock.Anything, req).Return(nil, expectedError)

	// Seed cookie
	testutil.SetCookie("refresh_token", req.RefreshToken, 3600)
	// Act
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodPost,
		RefreshTokenEndpoint,
		nil,
		req,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusUnauthorized, code)
	s.R.Equal(http.StatusUnauthorized, resp.Code)
	s.R.Equal(authmanagers.ErrInvalidRefreshToken.Error(), resp.Message)
}

func (s *AuthControllerSuite) TestRefreshToken_RevokedToken() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.Managers.AuthManager = m

	req := authmodels.RefreshTokenRequest{
		RefreshToken: "revoked_refresh_token",
	}

	expectedError := authmanagers.ErrRefreshTokenRevoked
	m.EXPECT().RefreshToken(mock.Anything, req).Return(nil, expectedError)

	// Seed cookie
	testutil.SetCookie("refresh_token", req.RefreshToken, 3600)
	// Act
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodPost,
		RefreshTokenEndpoint,
		nil,
		req,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusUnauthorized, code)
	s.R.Equal(http.StatusUnauthorized, resp.Code)
	s.R.Equal(authmanagers.ErrRefreshTokenRevoked.Error(), resp.Message)
}

func (s *AuthControllerSuite) TestRefreshToken_ExpiredToken() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.Managers.AuthManager = m

	req := authmodels.RefreshTokenRequest{
		RefreshToken: "expired_refresh_token",
	}

	expectedError := authmanagers.ErrRefreshTokenExpired
	m.EXPECT().RefreshToken(mock.Anything, req).Return(nil, expectedError)

	// Seed cookie
	testutil.SetCookie("refresh_token", req.RefreshToken, 1)
	// Act
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodPost,
		RefreshTokenEndpoint,
		nil,
		req,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusUnauthorized, code)
	s.R.Equal(http.StatusUnauthorized, resp.Code)
	s.R.Equal(authmanagers.ErrRefreshTokenExpired.Error(), resp.Message)
}

func (s *AuthControllerSuite) TestRefreshToken_EmptyToken() {
	// Arrange
	req := authmodels.RefreshTokenRequest{
		RefreshToken: "",
	}

	// Ensure no cookie
	testutil.ClearCookies()
	// Act
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodPost,
		RefreshTokenEndpoint,
		nil,
		req,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusUnauthorized, code)
	s.R.Equal(http.StatusUnauthorized, resp.Code)
	s.R.Equal("Missing refresh token", resp.Message)
}

func (s *AuthControllerSuite) TestRefreshToken_DatabaseError() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.Managers.AuthManager = m

	req := authmodels.RefreshTokenRequest{
		RefreshToken: "valid_refresh_token",
	}

	expectedError := errors.New("database error")
	m.EXPECT().RefreshToken(mock.Anything, req).Return(nil, expectedError)

	// Seed cookie
	testutil.SetCookie("refresh_token", req.RefreshToken, 3600)
	// Act
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodPost,
		RefreshTokenEndpoint,
		nil,
		req,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusInternalServerError, code)
	s.R.Equal(http.StatusInternalServerError, resp.Code)
	s.R.Equal("Internal server error", resp.Message)
}

// Logout Tests

func (s *AuthControllerSuite) TestLogout_Success() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.Managers.AuthManager = m

	req := authmodels.LogoutRequest{
		RefreshToken: "valid_refresh_token",
	}

	m.EXPECT().Logout(mock.Anything, req).Return(nil)

	// Seed cookie
	testutil.SetCookie("refresh_token", req.RefreshToken, 3600)
	// Act
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[string]](
		s.Echo,
		http.MethodPost,
		LogoutEndpoint,
		nil,
		req,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, code)
	s.R.Equal("success", resp.Message)
	s.R.Equal("Logged out successfully", resp.Data)
}

func (s *AuthControllerSuite) TestLogout_InvalidToken() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.Managers.AuthManager = m

	req := authmodels.LogoutRequest{
		RefreshToken: "invalid_refresh_token",
	}

	expectedError := errors.New("invalid token")
	m.EXPECT().Logout(mock.Anything, req).Return(expectedError)

	// Seed cookie
	testutil.SetCookie("refresh_token", req.RefreshToken, 3600)
	// Act
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodPost,
		LogoutEndpoint,
		nil,
		req,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusInternalServerError, code)
	s.R.Equal(http.StatusInternalServerError, resp.Code)
	s.R.Equal("Internal server error", resp.Message)
}

func (s *AuthControllerSuite) TestLogout_EmptyToken() {
	// Arrange
	req := authmodels.LogoutRequest{
		RefreshToken: "",
	}

	// Ensure no cookie
	testutil.ClearCookies()
	// Act
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodPost,
		LogoutEndpoint,
		nil,
		req,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusUnauthorized, code)
	s.R.Equal(http.StatusUnauthorized, resp.Code)
	s.R.Equal("Missing refresh token", resp.Message)
}

func (s *AuthControllerSuite) TestLogout_DatabaseError() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.Managers.AuthManager = m

	req := authmodels.LogoutRequest{
		RefreshToken: "valid_refresh_token",
	}

	expectedError := errors.New("database connection failed")
	m.EXPECT().Logout(mock.Anything, req).Return(expectedError)

	// Seed cookie
	testutil.SetCookie("refresh_token", req.RefreshToken, 3600)
	// Act
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodPost,
		LogoutEndpoint,
		nil,
		req,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusInternalServerError, code)
	s.R.Equal(http.StatusInternalServerError, resp.Code)
	s.R.Equal("Internal server error", resp.Message)
}

// Me Tests

func (s *AuthControllerSuite) TestMe_Success() {
	// Arrange
	cfg := s.Resource.Config
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
	s.R.NoError(err)

	// Act
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[authmodels.MeResponse]](
		s.Echo,
		http.MethodGet,
		MeEndpoint,
		&accessToken.Token,
		nil,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, code)
	s.R.Equal("success", resp.Message)
	s.R.Equal(userID, resp.Data.ID)
	s.R.Equal(username, resp.Data.Username)
	s.R.Equal(email, *resp.Data.Email)
	s.R.Equal(phoneNumber, *resp.Data.PhoneNumber)
	s.R.Equal(role.User, resp.Data.Role)
	s.R.Equal(emailVerified, resp.Data.EmailVerified)
	s.R.Equal(phoneVerified, resp.Data.PhoneVerified)
}

func (s *AuthControllerSuite) TestMe_MissingAuthorization() {
	// Act
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodGet,
		MeEndpoint,
		nil,
		nil,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusUnauthorized, code)
	s.R.Equal(http.StatusUnauthorized, resp.Code)
}

func (s *AuthControllerSuite) TestMe_InvalidToken() {
	// Arrange
	invalidToken := "invalid.jwt.token"

	// Act
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodGet,
		MeEndpoint,
		&invalidToken,
		nil,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusUnauthorized, code)
	s.R.Equal(http.StatusUnauthorized, resp.Code)
}

// Edge Cases and Complex Scenarios

func (s *AuthControllerSuite) TestLogin_SpecialCharactersInPassword() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.Managers.AuthManager = m

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
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[authmodels.AuthResponse]](
		s.Echo,
		http.MethodPost,
		LoginEndpoint,
		nil,
		req,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, code)
	s.R.Equal("success", resp.Message)
	s.R.Equal("test@example.com", *resp.Data.Username)
}

func (s *AuthControllerSuite) TestRefreshToken_ValidRequestWithLongToken() {
	// Arrange
	m := mocks.NewMockAuthManager(s.T())
	s.Managers.AuthManager = m

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
	testutil.SetCookie("refresh_token", longToken, 3600)
	// Act
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[authmodels.AuthResponse]](
		s.Echo,
		http.MethodPost,
		RefreshTokenEndpoint,
		nil,
		req,
	)

	// Assert
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, code)
	s.R.Equal("success", resp.Message)
}
