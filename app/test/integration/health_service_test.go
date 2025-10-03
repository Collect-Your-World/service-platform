package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"backend/service-platform/app/api/client/response"
	httputil "backend/service-platform/app/test/util"
)

const (
	HealthEndpointK8S = "/health"
)

type HealthServiceSuite struct {
	RouterSuite
}

func TestHealthServiceSuite(t *testing.T) {
	suite.Run(t, new(HealthServiceSuite))
}

func (s *HealthServiceSuite) TestCheckHealth() {
	resp, code, err := httputil.RequestHTTP[response.GeneralResponse[response.HealthResponse]](s.e, http.MethodGet, HealthEndpointK8S, nil, nil)
	s.r.NoError(err)
	s.r.Equal(http.StatusOK, code)
	s.a.Equal("success", resp.Message)
	s.a.Equal("up", resp.Data.Status)
}
