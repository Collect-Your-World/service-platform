package jwt_test

import (
	"backend/service-platform/app/internal/config"
	jwtpkg "backend/service-platform/app/pkg/jwt"
	"encoding/base64"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestJwt() *jwtpkg.DefaultJwt {
	testConfig := config.JwtConfig{
		Issuer:            "test-issuer",
		SecretKey:         "test-secret-key-12345",
		AccessExpiration:  1 * time.Hour,
		RefreshExpiration: 2 * time.Hour,
	}
	return jwtpkg.NewJwt(testConfig).(*jwtpkg.DefaultJwt)
}

func TestDefaultJwt_GetExpirationTime(t *testing.T) {
	jwtService := createTestJwt()

	expectedSeconds := int64(3600) // 1 hour in seconds
	actualSeconds := jwtService.GetExpirationTime()

	assert.Equal(t, expectedSeconds, actualSeconds)
}

func TestDefaultJwt_ParseToken_ValidToken(t *testing.T) {
	jwtService := createTestJwt()

	// Create test claims
	claims := &jwtpkg.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "test-issuer",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}

	// Generate a valid token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret-key-12345"))
	require.NoError(t, err)

	// Parse the token
	parsedToken, err := jwtService.ParseToken(tokenString)

	assert.NoError(t, err)
	assert.NotNil(t, parsedToken)
	assert.True(t, parsedToken.Valid)
}

func TestDefaultJwt_ParseToken_InvalidToken(t *testing.T) {
	jwtService := createTestJwt()

	tests := []struct {
		name  string
		token string
	}{
		{"Empty token", ""},
		{"Invalid format", "invalid.token.format"},
		{"Wrong secret", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"},
		{"Malformed token", "not.a.jwt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedToken, err := jwtService.ParseToken(tt.token)

			assert.Error(t, err)
			// ParseToken may still return a token object even with errors
			if parsedToken != nil {
				assert.False(t, parsedToken.Valid)
			}
		})
	}
}

func TestDefaultJwt_ValidateToken_ValidToken(t *testing.T) {
	jwtService := createTestJwt()

	userID := uuid.New()
	username := "testuser"
	email := "test@example.com"
	phoneNumber := "+1234567890"
	roleStr := "USER"
	emailVerified := true
	phoneVerified := false
	lastLoginAt := time.Now()

	// Generate access token
	accessToken, err := jwtService.GenerateAccessToken(&userID, &username, &email, &phoneNumber, &roleStr, &emailVerified, &phoneVerified, &lastLoginAt)
	require.NoError(t, err)

	// Validate the token
	claims, err := jwtService.ValidateToken(accessToken.Token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, *claims.UserID)
	assert.Equal(t, username, *claims.Username)
	assert.Equal(t, email, *claims.Email)
	assert.Equal(t, phoneNumber, *claims.PhoneNumber)
	assert.Equal(t, roleStr, *claims.Role)
	assert.Equal(t, emailVerified, *claims.EmailVerified)
	assert.Equal(t, phoneVerified, *claims.PhoneVerified)
	assert.Nil(t, claims.RefreshTokenBase64)
}

func TestDefaultJwt_ValidateToken_InvalidToken(t *testing.T) {
	jwtService := createTestJwt()

	tests := []struct {
		name  string
		token string
	}{
		{"Empty token", ""},
		{"Invalid token", "invalid.token.string"},
		{"Expired token", generateExpiredToken(jwtService)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := jwtService.ValidateToken(tt.token)

			assert.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}

func TestDefaultJwt_GenerateAccessToken(t *testing.T) {
	jwtService := createTestJwt()

	userID := uuid.New()
	username := "testuser"
	email := "test@example.com"
	phoneNumber := "+1234567890"
	roleStr := "USER"
	emailVerified := true
	phoneVerified := false
	lastLoginAt := time.Now()

	accessToken, err := jwtService.GenerateAccessToken(&userID, &username, &email, &phoneNumber, &roleStr, &emailVerified, &phoneVerified, &lastLoginAt)

	assert.NoError(t, err)
	assert.NotNil(t, accessToken)
	assert.NotEmpty(t, accessToken.Token)
	assert.True(t, accessToken.ExpiredAt.After(time.Now()))

	// Validate that the generated token can be parsed
	claims, err := jwtService.ValidateToken(accessToken.Token)
	assert.NoError(t, err)
	assert.Equal(t, userID, *claims.UserID)
	assert.Equal(t, username, *claims.Username)
	assert.Equal(t, email, *claims.Email)
	assert.Equal(t, phoneNumber, *claims.PhoneNumber)
	assert.Equal(t, roleStr, *claims.Role)
	assert.Equal(t, emailVerified, *claims.EmailVerified)
	assert.Equal(t, phoneVerified, *claims.PhoneVerified)
	assert.Nil(t, claims.RefreshTokenBase64)
}

func TestDefaultJwt_GenerateAccessToken_NilValues(t *testing.T) {
	jwtService := createTestJwt()

	accessToken, err := jwtService.GenerateAccessToken(nil, nil, nil, nil, nil, nil, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, accessToken)
	assert.NotEmpty(t, accessToken.Token)

	// Validate that the generated token can be parsed
	claims, err := jwtService.ValidateToken(accessToken.Token)
	assert.NoError(t, err)
	assert.Nil(t, claims.UserID)
	assert.Nil(t, claims.Username)
	assert.Nil(t, claims.Email)
	assert.Nil(t, claims.PhoneNumber)
	assert.Nil(t, claims.Role)
	assert.Nil(t, claims.EmailVerified)
	assert.Nil(t, claims.PhoneVerified)
	assert.Nil(t, claims.LastLoginAt)
}

func TestDefaultJwt_GenerateRefreshToken(t *testing.T) {
	jwtService := createTestJwt()

	userID := uuid.New()
	username := "adminuser"
	email := "admin@example.com"
	phoneNumber := "+1234567890"
	roleStr := "ADMIN"
	emailVerified := true
	phoneVerified := true
	lastLoginAt := time.Now()

	refreshToken, err := jwtService.GenerateRefreshToken(&userID, &username, &email, &phoneNumber, &roleStr, &emailVerified, &phoneVerified, &lastLoginAt)

	assert.NoError(t, err)
	assert.NotNil(t, refreshToken)
	assert.NotEmpty(t, refreshToken.Token)
	assert.NotEmpty(t, refreshToken.TokenBase64)
	assert.True(t, refreshToken.ExpiredAt.After(time.Now()))
	// Ensure refresh token uses RefreshExpiration (2h in createTestJwt)
	expectedMin := time.Now().Add(2 * time.Hour).Add(-5 * time.Second)
	expectedMax := time.Now().Add(2 * time.Hour).Add(5 * time.Second)
	assert.True(t, refreshToken.ExpiredAt.After(expectedMin))
	assert.True(t, refreshToken.ExpiredAt.Before(expectedMax))

	// Validate that the generated token can be parsed
	claims, err := jwtService.ValidateToken(refreshToken.Token)
	assert.NoError(t, err)
	assert.Equal(t, userID, *claims.UserID)
	assert.Equal(t, username, *claims.Username)
	assert.Equal(t, email, *claims.Email)
	assert.Equal(t, phoneNumber, *claims.PhoneNumber)
	assert.Equal(t, roleStr, *claims.Role)
	assert.Equal(t, emailVerified, *claims.EmailVerified)
	assert.Equal(t, phoneVerified, *claims.PhoneVerified)
	assert.Equal(t, refreshToken.TokenBase64, *claims.RefreshTokenBase64)

	// Validate that TokenBase64 is valid base64
	_, err = base64.StdEncoding.DecodeString(refreshToken.TokenBase64)
	assert.NoError(t, err)
}

func TestDefaultJwt_GenerateAccessTokenWithExpiration(t *testing.T) {
	jwtService := createTestJwt()

	now := time.Now()
	userID := uuid.New()
	username := "testuser"
	email := "test@example.com"
	phoneNumber := "+1234567890"
	roleStr := "USER"
	emailVerified := true
	phoneVerified := false
	lastLoginAt := now

	claims := &jwtpkg.Claims{
		UserID:        &userID,
		Username:      &username,
		Email:         &email,
		PhoneNumber:   &phoneNumber,
		Role:          &roleStr,
		EmailVerified: &emailVerified,
		PhoneVerified: &phoneVerified,
		LastLoginAt:   &lastLoginAt,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "test-issuer",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	tokenString, err := jwtService.GenerateAccessTokenWithExpiration(claims)

	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Validate that the token can be parsed and contains correct claims
	validatedClaims, err := jwtService.ValidateToken(tokenString)
	assert.NoError(t, err)
	assert.Equal(t, claims.UserID, validatedClaims.UserID)
	assert.Equal(t, claims.Username, validatedClaims.Username)
	assert.Equal(t, claims.Email, validatedClaims.Email)
	assert.Equal(t, claims.PhoneNumber, validatedClaims.PhoneNumber)
	assert.Equal(t, claims.Role, validatedClaims.Role)
	assert.Equal(t, claims.EmailVerified, validatedClaims.EmailVerified)
	assert.Equal(t, claims.PhoneVerified, validatedClaims.PhoneVerified)
	// Compare times with a small tolerance for precision differences
	if claims.LastLoginAt != nil && validatedClaims.LastLoginAt != nil {
		assert.WithinDuration(t, *claims.LastLoginAt, *validatedClaims.LastLoginAt, time.Second)
	} else {
		assert.Equal(t, claims.LastLoginAt, validatedClaims.LastLoginAt)
	}
	assert.Equal(t, claims.Issuer, validatedClaims.Issuer)
}

func TestGenerateRandomBase64(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"Length 16", 16},
		{"Length 32", 32},
		{"Length 64", 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base64String, err := jwtpkg.GenerateRandomBase64(tt.length)

			assert.NoError(t, err)
			assert.NotEmpty(t, base64String)

			// Decode to verify it's valid base64
			decoded, err := base64.StdEncoding.DecodeString(base64String)
			assert.NoError(t, err)
			assert.Equal(t, tt.length, len(decoded))

			// Generate another one and ensure they're different
			base64String2, err := jwtpkg.GenerateRandomBase64(tt.length)
			assert.NoError(t, err)
			assert.NotEqual(t, base64String, base64String2)
		})
	}
}

func TestGenerateRandomBase64_ZeroLength(t *testing.T) {
	base64String, err := jwtpkg.GenerateRandomBase64(0)

	assert.NoError(t, err)
	assert.Empty(t, base64String)
}

func TestDefaultJwt_ParseToken_WrongSigningMethod(t *testing.T) {
	jwtService := createTestJwt()

	// Create a token with a different signing method (RS256) and valid signature
	// that will pass parsing but fail during key function validation
	header := `{"alg":"RS256","typ":"JWT"}`
	payload := `{"iss":"test-issuer"}`

	headerB64 := base64.RawURLEncoding.EncodeToString([]byte(header))
	payloadB64 := base64.RawURLEncoding.EncodeToString([]byte(payload))
	signature := base64.RawURLEncoding.EncodeToString([]byte("valid_signature_bytes_here"))

	malformedToken := headerB64 + "." + payloadB64 + "." + signature

	_, err := jwtService.ParseToken(malformedToken)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected signing method")
}

func generateExpiredToken(_ *jwtpkg.DefaultJwt) string {
	pastTime := time.Now().Add(-2 * time.Hour)
	claims := &jwtpkg.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "test-issuer",
			IssuedAt:  jwt.NewNumericDate(pastTime),
			ExpiresAt: jwt.NewNumericDate(pastTime.Add(-1 * time.Hour)), // Already expired
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret-key-12345"))
	return tokenString
}

func TestNewJwt(t *testing.T) {
	testConfig := config.JwtConfig{
		Issuer:           "test-issuer",
		SecretKey:        "test-secret",
		AccessExpiration: 30 * time.Minute,
	}

	jwtService := jwtpkg.NewJwt(testConfig)

	assert.NotNil(t, jwtService)
	assert.IsType(t, &jwtpkg.DefaultJwt{}, jwtService)

	// Just verify that the service is created correctly and works
	expectedSeconds := int64(1800) // 30 minutes in seconds
	actualSeconds := jwtService.GetExpirationTime()
	assert.Equal(t, expectedSeconds, actualSeconds)
}

func TestDefaultJwt_ValidateToken_InvalidToken_Format(t *testing.T) {
	jwtService := createTestJwt()

	// Test with a token that has invalid signature
	invalidToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.invalid_signature"

	claims, err := jwtService.ValidateToken(invalidToken)

	assert.Error(t, err)
	assert.Nil(t, claims)
}
