package httpClientUtil

import (
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// NewRestyClient returns a new resty client with the given request timeout.
// If requestTimeout is 0, no timeout will be set. The returned client will
// also log the raw response body at debug level using the logger from the
// request context.
func NewRestyClient(requestTimeout time.Duration, log *zap.Logger) *resty.Client {
	client := resty.New().
		SetRetryCount(3)

	if requestTimeout > 0 {
		client.SetTimeout(requestTimeout)
	}

	return client
}
