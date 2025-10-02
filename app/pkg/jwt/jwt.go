package jwt

import (
	"backend/service-platform/app/internal/config"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	TokenTypeBearer = "Bearer"
)

type DefaultJwt struct {
	config config.JwtConfig
}

func NewJwt(jwtConfig config.JwtConfig) Jwt {
	return &DefaultJwt{
		config: jwtConfig,
	}
}

func (m *DefaultJwt) GetExpirationTime() int64 {
	return int64(m.config.AccessExpiration.Seconds())
}

func (m *DefaultJwt) ParseToken(token string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.config.SecretKey), nil
	})
}

func (m *DefaultJwt) ValidateToken(token string) (*Claims, error) {
	t, err := m.ParseToken(token)
	if err != nil {
		return nil, err
	}

	if !t.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := t.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (m *DefaultJwt) GenerateAccessToken(
	userID *uuid.UUID,
	username *string,
	email *string,
	phoneNumber *string,
	role *string,
	emailVerified *bool,
	phoneVerified *bool,
	lastLoginAt *time.Time,
) (*AccessToken, error) {
	now := time.Now()
	claims := &Claims{
		UserID:             userID,
		Username:           username,
		Email:              email,
		PhoneNumber:        phoneNumber,
		Role:               role,
		EmailVerified:      emailVerified,
		PhoneVerified:      phoneVerified,
		LastLoginAt:        lastLoginAt,
		RefreshTokenBase64: nil,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.config.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.AccessExpiration)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	token, err := m.GenerateAccessTokenWithExpiration(claims)
	if err != nil {
		return nil, err
	}
	return &AccessToken{
		Token:     token,
		ExpiredAt: claims.RegisteredClaims.ExpiresAt.Time,
	}, nil
}

func (m *DefaultJwt) GenerateRefreshToken(
	userID *uuid.UUID,
	username *string,
	email *string,
	phoneNumber *string,
	role *string,
	emailVerified *bool,
	phoneVerified *bool,
	lastLoginAt *time.Time,
) (*RefreshToken, error) {
	tokenBase64, err := GenerateRandomBase64(32)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	claims := &Claims{
		UserID:             userID,
		Username:           username,
		Email:              email,
		PhoneNumber:        phoneNumber,
		Role:               role,
		EmailVerified:      emailVerified,
		PhoneVerified:      phoneVerified,
		LastLoginAt:        lastLoginAt,
		RefreshTokenBase64: &tokenBase64,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.config.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.RefreshExpiration)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	token, err := m.GenerateAccessTokenWithExpiration(claims)
	if err != nil {
		return nil, err
	}
	return &RefreshToken{
		Token:       token,
		TokenBase64: tokenBase64,
		ExpiredAt:   claims.RegisteredClaims.ExpiresAt.Time,
	}, nil
}

func (m *DefaultJwt) GenerateAccessTokenWithExpiration(claims *Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.config.SecretKey))
}

func GenerateRandomBase64(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

func (m *DefaultJwt) GetClaims(c echo.Context) (*Claims, error) {
	authorizationHeader := c.Request().Header.Get("Authorization")
	if strings.TrimSpace(authorizationHeader) == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	// Remove "Bearer " prefix
	token := strings.Replace(authorizationHeader, "Bearer ", "", 1)

	claims, err := m.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	return claims, nil
}
