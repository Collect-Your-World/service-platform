package cookie

import (
	"net/http"
	"os"
	"strings"
	"time"
)

func isHTTPS(req *http.Request) bool {
	if req == nil {
		return false
	}
	if req.TLS != nil {
		return true
	}
	xfProto := req.Header.Get("X-Forwarded-Proto")
	xfProtocol := req.Header.Get("X-Forwarded-Protocol")
	return strings.HasPrefix(strings.ToLower(xfProto), "https") || strings.HasPrefix(strings.ToLower(xfProtocol), "https")
}

func NewCookie(name string, value string, expiry time.Duration, req *http.Request) *http.Cookie {
	expiresAt := time.Now().Add(expiry)
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   isHTTPS(req) || os.Getenv("APP_ENV") == "production",
		SameSite: http.SameSiteStrictMode,
		Expires:  expiresAt,
		MaxAge:   int(expiry.Seconds()),
	}
	return c
}

func ExpireCookie(name string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	}
}

func NewRefreshTokenCookie(req *http.Request, token string, expiry time.Duration) *http.Cookie {
	return NewCookie("refresh_token", token, expiry, req)
}
