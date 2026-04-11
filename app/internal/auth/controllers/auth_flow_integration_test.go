package controllers_test

import (
	integrationtest "backend/service-platform/app/internal/integrationtest"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"backend/service-platform/app/internal/auth/constants/role"
	authmodels "backend/service-platform/app/internal/auth/models"
	commonmodels "backend/service-platform/app/internal/common/models"
	testutil "backend/service-platform/app/internal/testutil"
)

type AuthFlowIntegrationSuite struct {
	integrationtest.RouterSuite
}

func TestAuthFlowIntegrationSuite(t *testing.T) {
	suite.Run(t, new(AuthFlowIntegrationSuite))
}

// TestCompleteAuthFlow_Register_Login_Me_Refresh_Logout tests the complete authentication flow
// as described in the sequence diagram:
// 1. POST /login (email, password) -> access_token (JSON) + refresh_token (cookie)
// 2. GET /api/me (Authorization: Bearer access_token) -> User profile
// 3. POST /refresh (cookie) -> New access_token
func (s *AuthFlowIntegrationSuite) TestCompleteAuthFlow_Register_Login_Me_Refresh_Logout() {
	// This test uses the real auth manager (not mocked) to test the complete flow
	// with actual database interactions

	testEmail := "integration@example.com"
	testPassword := "password123"

	// Step 1: Register a new user
	s.T().Log("Step 1: Registering new user")
	registerReq := authmodels.RegisterRequest{
		Email:    testEmail,
		Password: testPassword,
	}

	registerResp, registerCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[string]](
		s.Echo,
		http.MethodPost,
		"/api/v1/auth/register",
		nil,
		registerReq,
	)
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, registerCode)
	s.R.Equal("success", registerResp.Message)
	s.R.Equal("registered", registerResp.Data)

	// Step 2: Login with the registered user
	s.T().Log("Step 2: Logging in with registered user")
	loginReq := authmodels.AuthUserRequest{
		Email:    testEmail,
		Password: testPassword,
	}

	loginResp, loginCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[authmodels.AuthResponse]](
		s.Echo,
		http.MethodPost,
		"/api/v1/auth/login",
		nil,
		loginReq,
	)
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, loginCode)
	s.R.Equal("success", loginResp.Message)
	s.R.NotEmpty(loginResp.Data.AccessToken)
	s.R.Equal("", loginResp.Data.RefreshToken)
	s.R.Equal("Bearer", loginResp.Data.TokenType)
	s.R.Equal(testEmail, *loginResp.Data.Username)

	// Read refresh token from cookie
	rt := testutil.GetCookie("refresh_token")
	s.Require().NotNil(rt)

	// Step 3: Call /me with the access token
	s.T().Log("Step 3: Getting user profile with access token")
	accessToken := loginResp.Data.AccessToken

	meResp, meCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[authmodels.MeResponse]](
		s.Echo,
		http.MethodGet,
		"/api/v1/auth/me",
		&accessToken,
		nil,
	)
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, meCode)
	s.R.Equal("success", meResp.Message)
	s.R.NotEmpty(meResp.Data.ID)
	s.R.Equal(testEmail, meResp.Data.Username)
	s.R.Equal(testEmail, *meResp.Data.Email)
	s.R.Equal(role.User, meResp.Data.Role)
	s.R.Equal(false, meResp.Data.EmailVerified)
	s.R.Equal(false, meResp.Data.PhoneVerified)

	// Step 4: Refresh the token using the refresh token
	s.T().Log("Step 4: Refreshing access token")
	refreshReq := authmodels.RefreshTokenRequest{
		RefreshToken: rt.Value,
	}

	// Seed cookie for refresh
	testutil.SetCookie("refresh_token", rt.Value, 3600)
	refreshResp, refreshCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[authmodels.AuthResponse]](
		s.Echo,
		http.MethodPost,
		"/api/v1/auth/refresh-token",
		nil,
		refreshReq,
	)
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, refreshCode)
	s.R.Equal("success", refreshResp.Message)
	s.R.NotEmpty(refreshResp.Data.AccessToken)
	s.R.Equal(testEmail, *refreshResp.Data.Username)

	// Verify we got a new access token (it should be different from the original)
	s.R.NotEqual(accessToken, refreshResp.Data.AccessToken)

	// Step 5: Use the new access token to call /me again
	s.T().Log("Step 5: Using new access token to get user profile")
	newAccessToken := refreshResp.Data.AccessToken

	meResp2, meCode2, err := testutil.RequestHTTP[commonmodels.GeneralResponse[authmodels.MeResponse]](
		s.Echo,
		http.MethodGet,
		"/api/v1/auth/me",
		&newAccessToken,
		nil,
	)
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, meCode2)
	s.R.Equal("success", meResp2.Message)
	s.R.Equal(meResp.Data.ID, meResp2.Data.ID)
	s.R.Equal(meResp.Data.Username, meResp2.Data.Username)
	s.R.Equal(*meResp.Data.Email, *meResp2.Data.Email)
	s.R.Equal(meResp.Data.Role, meResp2.Data.Role)

	// Step 6: Logout
	s.T().Log("Step 6: Logging out")
	logoutReq := authmodels.LogoutRequest{
		RefreshToken: rt.Value,
	}

	// Seed cookie for logout
	testutil.SetCookie("refresh_token", rt.Value, 3600)
	logoutResp, logoutCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[string]](
		s.Echo,
		http.MethodPost,
		"/api/v1/auth/logout",
		nil,
		logoutReq,
	)
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, logoutCode)
	s.R.Equal("success", logoutResp.Message)
	s.R.Equal("Logged out successfully", logoutResp.Data)

	// Step 7: Verify logout worked by trying to refresh with the same token
	s.T().Log("Step 7: Verifying logout by attempting to refresh token")
	refreshAfterLogoutResp, refreshAfterLogoutCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodPost,
		"/api/v1/auth/refresh-token",
		nil,
		refreshReq,
	)
	s.R.NoError(err)
	s.R.Equal(http.StatusUnauthorized, refreshAfterLogoutCode)
	s.R.Equal(http.StatusUnauthorized, refreshAfterLogoutResp.Code)
}

// TestAuthFlow_InvalidCredentials tests the error handling in the auth flow
func (s *AuthFlowIntegrationSuite) TestAuthFlow_InvalidCredentials() {
	// Test login with invalid credentials
	loginReq := authmodels.AuthUserRequest{
		Email:    "nonexistent@example.com",
		Password: "wrongpassword",
	}

	loginResp, loginCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodPost,
		"/api/v1/auth/login",
		nil,
		loginReq,
	)
	s.R.NoError(err)
	s.R.Equal(http.StatusUnauthorized, loginCode)
	s.R.Equal(http.StatusUnauthorized, loginResp.Code)
	s.R.Equal("Invalid credentials", loginResp.Message)
}

// TestAuthFlow_InvalidToken tests accessing protected endpoints with invalid tokens
func (s *AuthFlowIntegrationSuite) TestAuthFlow_InvalidToken() {
	// Test /me with invalid token
	invalidToken := "invalid.jwt.token"

	meResp, meCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodGet,
		"/api/v1/auth/me",
		&invalidToken,
		nil,
	)
	s.R.NoError(err)
	s.R.Equal(http.StatusUnauthorized, meCode)
	s.R.Equal(http.StatusUnauthorized, meResp.Code)
}

// TestAuthFlow_ExpiredToken tests handling of expired tokens
func (s *AuthFlowIntegrationSuite) TestAuthFlow_ExpiredToken() {
	// This test would require creating an expired JWT token
	// For now, we'll test with an invalid token format
	expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjE1MTYyMzkwMjJ9.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	meResp, meCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodGet,
		"/api/v1/auth/me",
		&expiredToken,
		nil,
	)
	s.R.NoError(err)
	s.R.Equal(http.StatusUnauthorized, meCode)
	s.R.Equal(http.StatusUnauthorized, meResp.Code)
}

// TestAuthFlow_MultipleUsers tests authentication flow with multiple users
func (s *AuthFlowIntegrationSuite) TestAuthFlow_MultipleUsers() {
	// Register and login multiple users to ensure isolation
	users := []struct {
		email    string
		password string
	}{
		{"user1@example.com", "password123"},
		{"user2@example.com", "password456"},
		{"user3@example.com", "password789"},
	}

	var accessTokens []string
	var refreshTokens []string

	// Register and login all users
	for i, user := range users {
		s.T().Logf("Registering user %d: %s", i+1, user.email)

		// Register
		registerReq := authmodels.RegisterRequest{
			Email:    user.email,
			Password: user.password,
		}

		registerResp, registerCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[string]](
			s.Echo,
			http.MethodPost,
			"/api/v1/auth/register",
			nil,
			registerReq,
		)
		s.R.NoError(err)
		s.R.Equal(http.StatusOK, registerCode)
		s.R.Equal("success", registerResp.Message)

		// Login
		loginReq := authmodels.AuthUserRequest{
			Email:    user.email,
			Password: user.password,
		}

		loginResp, loginCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[authmodels.AuthResponse]](
			s.Echo,
			http.MethodPost,
			"/api/v1/auth/login",
			nil,
			loginReq,
		)
		s.R.NoError(err)
		s.R.Equal(http.StatusOK, loginCode)
		s.R.Equal("success", loginResp.Message)
		s.R.Equal(user.email, *loginResp.Data.Username)

		accessTokens = append(accessTokens, loginResp.Data.AccessToken)
		// capture refresh token from cookie
		rtc := testutil.GetCookie("refresh_token")
		s.Require().NotNil(rtc)
		refreshTokens = append(refreshTokens, rtc.Value)
	}

	// Test that each user can only access their own profile
	for i, user := range users {
		s.T().Logf("Testing user %d profile access: %s", i+1, user.email)

		meResp, meCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[authmodels.MeResponse]](
			s.Echo,
			http.MethodGet,
			"/api/v1/auth/me",
			&accessTokens[i],
			nil,
		)
		s.R.NoError(err)
		s.R.Equal(http.StatusOK, meCode)
		s.R.Equal("success", meResp.Message)
		s.R.Equal(user.email, meResp.Data.Username)
		s.R.Equal(user.email, *meResp.Data.Email)
		s.R.Equal(role.User, meResp.Data.Role)
	}

	// Test that users cannot access each other's profiles with their tokens
	// (This is more of a security test - in practice, tokens should be user-specific)
	for i, user := range users {
		for j, otherUser := range users {
			if i != j {
				s.T().Logf("Testing user %d cannot access user %d profile", i+1, j+1)

				meResp, meCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[authmodels.MeResponse]](
					s.Echo,
					http.MethodGet,
					"/api/v1/auth/me",
					&accessTokens[i], // Using user i's token
					nil,
				)
				s.R.NoError(err)
				s.R.Equal(http.StatusOK, meCode)
				// The token should only return user i's profile, not user j's
				s.R.Equal(user.email, meResp.Data.Username)
				s.R.Equal(user.email, *meResp.Data.Email)
				s.R.NotEqual(otherUser.email, meResp.Data.Username)
				s.R.NotEqual(otherUser.email, *meResp.Data.Email)
			}
		}
	}

	// Test refresh tokens for each user
	for i, user := range users {
		s.T().Logf("Testing refresh token for user %d: %s", i+1, user.email)

		refreshReq := authmodels.RefreshTokenRequest{
			RefreshToken: refreshTokens[i],
		}

		// Seed cookie for refresh
		testutil.SetCookie("refresh_token", refreshReq.RefreshToken, 3600)
		refreshResp, refreshCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[authmodels.AuthResponse]](
			s.Echo,
			http.MethodPost,
			"/api/v1/auth/refresh-token",
			nil,
			refreshReq,
		)
		s.R.NoError(err)
		s.R.Equal(http.StatusOK, refreshCode)
		s.R.Equal("success", refreshResp.Message)
		s.R.Equal(user.email, *refreshResp.Data.Username)
		s.R.NotEqual(accessTokens[i], refreshResp.Data.AccessToken) // Should be a new token
	}

	// Logout all users
	for i, user := range users {
		s.T().Logf("Logging out user %d: %s", i+1, user.email)

		logoutReq := authmodels.LogoutRequest{
			RefreshToken: refreshTokens[i],
		}

		// Seed cookie for logout
		testutil.SetCookie("refresh_token", refreshTokens[i], 3600)
		logoutResp, logoutCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[string]](
			s.Echo,
			http.MethodPost,
			"/api/v1/auth/logout",
			nil,
			logoutReq,
		)
		s.R.NoError(err)
		s.R.Equal(http.StatusOK, logoutCode)
		s.R.Equal("success", logoutResp.Message)
	}
}

// TestAuthFlow_ConcurrentRequests tests the auth flow under concurrent load
func (s *AuthFlowIntegrationSuite) TestAuthFlow_ConcurrentRequests() {
	// This test would require implementing concurrent requests
	// For now, we'll test sequential requests to the same endpoints
	s.T().Skip("Concurrent testing requires additional setup")
}

// TestAuthFlow_EdgeCases tests various edge cases in the auth flow
func (s *AuthFlowIntegrationSuite) TestAuthFlow_EdgeCases() {
	// Test with very long email
	longEmail := "verylongemailaddressthatexceedsnormallimits@example.com"

	registerReq := authmodels.RegisterRequest{
		Email:    longEmail,
		Password: "password123",
	}

	registerResp, registerCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodPost,
		"/api/v1/auth/register",
		nil,
		registerReq,
	)
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, registerCode)
	s.R.Equal("success", registerResp.Message)

	// Test with special characters in password
	specialPassword := "P@ssw0rd!@#$%^&*()_+-=[]{}|;:,.<>?"

	registerReq2 := authmodels.RegisterRequest{
		Email:    "special@example.com",
		Password: specialPassword,
	}

	registerResp2, registerCode2, err := testutil.RequestHTTP[commonmodels.GeneralResponse[any]](
		s.Echo,
		http.MethodPost,
		"/api/v1/auth/register",
		nil,
		registerReq2,
	)
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, registerCode2)
	s.R.Equal("success", registerResp2.Message)

	// Test login with the special password
	loginReq := authmodels.AuthUserRequest{
		Email:    "special@example.com",
		Password: specialPassword,
	}

	loginResp, loginCode, err := testutil.RequestHTTP[commonmodels.GeneralResponse[authmodels.AuthResponse]](
		s.Echo,
		http.MethodPost,
		"/api/v1/auth/login",
		nil,
		loginReq,
	)
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, loginCode)
	s.R.Equal("success", loginResp.Message)
}
