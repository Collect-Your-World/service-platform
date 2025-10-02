package jwt

import "time"

type RefreshToken struct {
	Token       string
	TokenBase64 string
	ExpiredAt   time.Time
}
