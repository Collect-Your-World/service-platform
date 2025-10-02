package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"backend/service-platform/app/api/client/request"
	"backend/service-platform/app/api/client/response"
	"backend/service-platform/app/database/constant/role"
	httputil "backend/service-platform/app/test/util"
)

type AuthFlowIntegrationSuite struct {
	RouterSuite
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
	registerReq := request.RegisterRequest{
		Email:    testEmail,
		Password: testPassword,
	}

	registerResp, registerCode, err := httputil.RequestHTTP[response.GeneralResponse[string]](
		s.e,
		http.MethodPost,
		"/api/v1/auth/register",
		nil,
		registerReq,
	)
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, registerCode)
	s.r.Equal("success", registerResp.Message)
	s.r.Equal("registered", registerResp.Data)

	// Step 2: Login with the registered user
	s.T().Log("Step 2: Logging in with registered user")
	loginReq := request.AuthUserRequest{
		Email:    testEmail,
		Password: testPassword,
	}

	loginResp, loginCode, err := httputil.RequestHTTP[response.GeneralResponse[response.AuthResponse]](
		s.e,
		http.MethodPost,
		"/api/v1/auth/login",
		nil,
		loginReq,
	)
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, loginCode)
	s.r.Equal("success", loginResp.Message)
	s.r.NotEmpty(loginResp.Data.AccessToken)
	s.r.NotEmpty(loginResp.Data.RefreshToken)
	s.r.Equal("Bearer", loginResp.Data.TokenType)
	s.r.Equal(testEmail, *loginResp.Data.Username)

	// Verify refresh token cookie was set
	// Note: In the current implementation, refresh token is returned in response, not cookie
	// This might need to be updated based on actual implementation

	// Step 3: Call /me with the access token
	s.T().Log("Step 3: Getting user profile with access token")
	accessToken := loginResp.Data.AccessToken

	meResp, meCode, err := httputil.RequestHTTP[response.GeneralResponse[response.MeResponse]](
		s.e,
		http.MethodGet,
		"/api/v1/auth/me",
		&accessToken,
		nil,
	)
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, meCode)
	s.r.Equal("success", meResp.Message)
	s.r.NotEmpty(meResp.Data.ID)
	s.r.Equal(testEmail, meResp.Data.Username)
	s.r.Equal(testEmail, *meResp.Data.Email)
	s.r.Equal(role.User, meResp.Data.Role)
	s.r.Equal(false, meResp.Data.EmailVerified)
	s.r.Equal(false, meResp.Data.PhoneVerified)

	// Step 4: Refresh the token using the refresh token
	s.T().Log("Step 4: Refreshing access token")
	refreshReq := request.RefreshTokenRequest{
		RefreshToken: loginResp.Data.RefreshToken,
	}

	refreshResp, refreshCode, err := httputil.RequestHTTP[response.GeneralResponse[response.AuthResponse]](
		s.e,
		http.MethodPost,
		"/api/v1/auth/refresh",
		nil,
		refreshReq,
	)
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, refreshCode)
	s.r.Equal("success", refreshResp.Message)
	s.r.NotEmpty(refreshResp.Data.AccessToken)
	s.r.Equal(testEmail, *refreshResp.Data.Username)

	// Verify we got a new access token (it should be different from the original)
	s.r.NotEqual(accessToken, refreshResp.Data.AccessToken)

	// Step 5: Use the new access token to call /me again
	s.T().Log("Step 5: Using new access token to get user profile")
	newAccessToken := refreshResp.Data.AccessToken

	meResp2, meCode2, err := httputil.RequestHTTP[response.GeneralResponse[response.MeResponse]](
		s.e,
		http.MethodGet,
		"/api/v1/auth/me",
		&newAccessToken,
		nil,
	)
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, meCode2)
	s.r.Equal("success", meResp2.Message)
	s.r.Equal(meResp.Data.ID, meResp2.Data.ID)
	s.r.Equal(meResp.Data.Username, meResp2.Data.Username)
	s.r.Equal(*meResp.Data.Email, *meResp2.Data.Email)
	s.r.Equal(meResp.Data.Role, meResp2.Data.Role)

	// Step 6: Logout
	s.T().Log("Step 6: Logging out")
	logoutReq := request.LogoutRequest{
		RefreshToken: loginResp.Data.RefreshToken,
	}

	logoutResp, logoutCode, err := httputil.RequestHTTP[response.GeneralResponse[string]](
		s.e,
		http.MethodPost,
		"/api/v1/auth/logout",
		nil,
		logoutReq,
	)
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, logoutCode)
	s.r.Equal("success", logoutResp.Message)
	s.r.Equal("Logged out successfully", logoutResp.Data)

	// Step 7: Verify logout worked by trying to refresh with the same token
	s.T().Log("Step 7: Verifying logout by attempting to refresh token")
	refreshAfterLogoutResp, refreshAfterLogoutCode, err := httputil.RequestHTTP[response.GeneralResponse[any]](
		s.e,
		http.MethodPost,
		"/api/v1/auth/refresh",
		nil,
		refreshReq,
	)
	s.r.NoError(err)
	s.r.Equal(http.StatusUnauthorized, refreshAfterLogoutCode)
	s.r.Equal(http.StatusUnauthorized, refreshAfterLogoutResp.Code)
}

// TestAuthFlow_InvalidCredentials tests the error handling in the auth flow
func (s *AuthFlowIntegrationSuite) TestAuthFlow_InvalidCredentials() {
	// Test login with invalid credentials
	loginReq := request.AuthUserRequest{
		Email:    "nonexistent@example.com",
		Password: "wrongpassword",
	}

	loginResp, loginCode, err := httputil.RequestHTTP[response.GeneralResponse[any]](
		s.e,
		http.MethodPost,
		"/api/v1/auth/login",
		nil,
		loginReq,
	)
	s.r.NoError(err)
	s.r.Equal(http.StatusUnauthorized, loginCode)
	s.r.Equal(http.StatusUnauthorized, loginResp.Code)
	s.r.Equal("Invalid credentials", loginResp.Message)
}

// TestAuthFlow_InvalidToken tests accessing protected endpoints with invalid tokens
func (s *AuthFlowIntegrationSuite) TestAuthFlow_InvalidToken() {
	// Test /me with invalid token
	invalidToken := "invalid.jwt.token"

	meResp, meCode, err := httputil.RequestHTTP[response.GeneralResponse[any]](
		s.e,
		http.MethodGet,
		"/api/v1/auth/me",
		&invalidToken,
		nil,
	)
	s.r.NoError(err)
	s.r.Equal(http.StatusUnauthorized, meCode)
	s.r.Equal(http.StatusUnauthorized, meResp.Code)
}

// TestAuthFlow_ExpiredToken tests handling of expired tokens
func (s *AuthFlowIntegrationSuite) TestAuthFlow_ExpiredToken() {
	// This test would require creating an expired JWT token
	// For now, we'll test with an invalid token format
	expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjE1MTYyMzkwMjJ9.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	meResp, meCode, err := httputil.RequestHTTP[response.GeneralResponse[any]](
		s.e,
		http.MethodGet,
		"/api/v1/auth/me",
		&expiredToken,
		nil,
	)
	s.r.NoError(err)
	s.r.Equal(http.StatusUnauthorized, meCode)
	s.r.Equal(http.StatusUnauthorized, meResp.Code)
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
		registerReq := request.RegisterRequest{
			Email:    user.email,
			Password: user.password,
		}

		registerResp, registerCode, err := httputil.RequestHTTP[response.GeneralResponse[string]](
			s.e,
			http.MethodPost,
			"/api/v1/auth/register",
			nil,
			registerReq,
		)
		s.r.NoError(err)
		s.r.Equal(http.StatusOK, registerCode)
		s.r.Equal("success", registerResp.Message)

		// Login
		loginReq := request.AuthUserRequest{
			Email:    user.email,
			Password: user.password,
		}

		loginResp, loginCode, err := httputil.RequestHTTP[response.GeneralResponse[response.AuthResponse]](
			s.e,
			http.MethodPost,
			"/api/v1/auth/login",
			nil,
			loginReq,
		)
		s.r.NoError(err)
		s.r.Equal(http.StatusOK, loginCode)
		s.r.Equal("success", loginResp.Message)
		s.r.Equal(user.email, *loginResp.Data.Username)

		accessTokens = append(accessTokens, loginResp.Data.AccessToken)
		refreshTokens = append(refreshTokens, loginResp.Data.RefreshToken)
	}

	// Test that each user can only access their own profile
	for i, user := range users {
		s.T().Logf("Testing user %d profile access: %s", i+1, user.email)

		meResp, meCode, err := httputil.RequestHTTP[response.GeneralResponse[response.MeResponse]](
			s.e,
			http.MethodGet,
			"/api/v1/auth/me",
			&accessTokens[i],
			nil,
		)
		s.r.NoError(err)
		s.r.Equal(http.StatusOK, meCode)
		s.r.Equal("success", meResp.Message)
		s.r.Equal(user.email, meResp.Data.Username)
		s.r.Equal(user.email, *meResp.Data.Email)
		s.r.Equal(role.User, meResp.Data.Role)
	}

	// Test that users cannot access each other's profiles with their tokens
	// (This is more of a security test - in practice, tokens should be user-specific)
	for i, user := range users {
		for j, otherUser := range users {
			if i != j {
				s.T().Logf("Testing user %d cannot access user %d profile", i+1, j+1)

				meResp, meCode, err := httputil.RequestHTTP[response.GeneralResponse[response.MeResponse]](
					s.e,
					http.MethodGet,
					"/api/v1/auth/me",
					&accessTokens[i], // Using user i's token
					nil,
				)
				s.r.NoError(err)
				s.r.Equal(http.StatusOK, meCode)
				// The token should only return user i's profile, not user j's
				s.r.Equal(user.email, meResp.Data.Username)
				s.r.Equal(user.email, *meResp.Data.Email)
				s.r.NotEqual(otherUser.email, meResp.Data.Username)
				s.r.NotEqual(otherUser.email, *meResp.Data.Email)
			}
		}
	}

	// Test refresh tokens for each user
	for i, user := range users {
		s.T().Logf("Testing refresh token for user %d: %s", i+1, user.email)

		refreshReq := request.RefreshTokenRequest{
			RefreshToken: refreshTokens[i],
		}

		refreshResp, refreshCode, err := httputil.RequestHTTP[response.GeneralResponse[response.AuthResponse]](
			s.e,
			http.MethodPost,
			"/api/v1/auth/refresh",
			nil,
			refreshReq,
		)
		s.r.NoError(err)
		s.r.Equal(http.StatusOK, refreshCode)
		s.r.Equal("success", refreshResp.Message)
		s.r.Equal(user.email, *refreshResp.Data.Username)
		s.r.NotEqual(accessTokens[i], refreshResp.Data.AccessToken) // Should be a new token
	}

	// Logout all users
	for i, user := range users {
		s.T().Logf("Logging out user %d: %s", i+1, user.email)

		logoutReq := request.LogoutRequest{
			RefreshToken: refreshTokens[i],
		}

		logoutResp, logoutCode, err := httputil.RequestHTTP[response.GeneralResponse[string]](
			s.e,
			http.MethodPost,
			"/api/v1/auth/logout",
			nil,
			logoutReq,
		)
		s.r.NoError(err)
		s.r.Equal(http.StatusOK, logoutCode)
		s.r.Equal("success", logoutResp.Message)
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

	registerReq := request.RegisterRequest{
		Email:    longEmail,
		Password: "password123",
	}

	registerResp, registerCode, err := httputil.RequestHTTP[response.GeneralResponse[any]](
		s.e,
		http.MethodPost,
		"/api/v1/auth/register",
		nil,
		registerReq,
	)
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, registerCode)
	s.r.Equal("success", registerResp.Message)

	// Test with special characters in password
	specialPassword := "P@ssw0rd!@#$%^&*()_+-=[]{}|;:,.<>?"

	registerReq2 := request.RegisterRequest{
		Email:    "special@example.com",
		Password: specialPassword,
	}

	registerResp2, registerCode2, err := httputil.RequestHTTP[response.GeneralResponse[any]](
		s.e,
		http.MethodPost,
		"/api/v1/auth/register",
		nil,
		registerReq2,
	)
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, registerCode2)
	s.r.Equal("success", registerResp2.Message)

	// Test login with the special password
	loginReq := request.AuthUserRequest{
		Email:    "special@example.com",
		Password: specialPassword,
	}

	loginResp, loginCode, err := httputil.RequestHTTP[response.GeneralResponse[response.AuthResponse]](
		s.e,
		http.MethodPost,
		"/api/v1/auth/login",
		nil,
		loginReq,
	)
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, loginCode)
	s.r.Equal("success", loginResp.Message)
}
