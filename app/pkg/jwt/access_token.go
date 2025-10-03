package jwt

import "time"

type AccessToken struct {
	Token     string
	ExpiredAt time.Time
}
