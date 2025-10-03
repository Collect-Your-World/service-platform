package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Claims struct {
	UserID             *uuid.UUID `json:"user_id,omitempty"`
	Username           *string    `json:"username,omitempty"`
	Email              *string    `json:"email,omitempty"`
	PhoneNumber        *string    `json:"phone_number,omitempty"`
	Role               *string    `json:"role,omitempty"`
	EmailVerified      *bool      `json:"email_verified,omitempty"`
	PhoneVerified      *bool      `json:"phone_verified,omitempty"`
	LastLoginAt        *time.Time `json:"last_login_at,omitempty"`
	RefreshTokenBase64 *string    `json:"refresh_token"`
	jwt.RegisteredClaims
}

type Jwt interface {
	GetExpirationTime() int64
	ParseToken(token string) (*jwt.Token, error)
	ValidateToken(token string) (*Claims, error)
	GenerateAccessToken(
		userID *uuid.UUID,
		username *string,
		email *string,
		phoneNumber *string,
		role *string,
		emailVerified *bool,
		phoneVerified *bool,
		lastLoginAt *time.Time,
	) (*AccessToken, error)
	GenerateRefreshToken(
		userID *uuid.UUID,
		username *string,
		email *string,
		phoneNumber *string,
		role *string,
		emailVerified *bool,
		phoneVerified *bool,
		lastLoginAt *time.Time,
	) (*RefreshToken, error)
	GenerateAccessTokenWithExpiration(claims *Claims) (string, error)
	GetClaims(c echo.Context) (*Claims, error)
}
