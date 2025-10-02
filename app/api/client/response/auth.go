package response

import (
	"backend/service-platform/app/database/constant/role"
	"time"

	"github.com/google/uuid"
)

type AuthResponse struct {
	Username     *string      `json:"username,omitempty"`
	Roles        *[]role.Role `json:"roles,omitempty"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	TokenType    string       `json:"token_type,default='Bearer'"`
}

type MeResponse struct {
	ID            uuid.UUID  `json:"id"`
	Username      string     `json:"username"`
	Email         *string    `json:"email,omitempty"`
	PhoneNumber   *string    `json:"phone_number,omitempty"`
	Role          role.Role  `json:"role"`
	EmailVerified bool       `json:"email_verified"`
	PhoneVerified bool       `json:"phone_verified"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}
