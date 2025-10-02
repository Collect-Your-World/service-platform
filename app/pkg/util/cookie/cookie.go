package cookie

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// GenerateCookieValue creates a cookie value string for the session token
// with appropriate security settings based on the request and configuration.
func GenerateCookieValue(req *http.Request, sessionToken string, sessionTokenExpiry time.Duration) string {
	sessionExpiration := time.Now().Add(sessionTokenExpiry)
	isHTTPS := strings.HasPrefix(req.Header.Get("X-Forwarded-Proto"), "https") ||
		strings.HasPrefix(req.Header.Get("X-Forwarded-Protocol"), "https")

	cookieValue := fmt.Sprintf(
		"%s=%s; Path=/; HttpOnly; SameSite=Strict; Expires=%s",
		"sessionToken",
		sessionToken,
		sessionExpiration.Format(time.RFC1123),
	)

	if isHTTPS {
		cookieValue += "; Secure"
	}

	return cookieValue
}
