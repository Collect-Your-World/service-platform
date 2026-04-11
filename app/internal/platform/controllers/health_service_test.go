package controllers_test

import (
	integrationtest "backend/service-platform/app/internal/integrationtest"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	commonmodels "backend/service-platform/app/internal/common/models"
	testutil "backend/service-platform/app/internal/testutil"
)

const (
	HealthEndpointK8S = "/health"
)

type HealthServiceSuite struct {
	integrationtest.RouterSuite
}

func TestHealthServiceSuite(t *testing.T) {
	suite.Run(t, new(HealthServiceSuite))
}

func (s *HealthServiceSuite) TestCheckHealth() {
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[commonmodels.HealthResponse]](s.Echo, http.MethodGet, HealthEndpointK8S, nil, nil)
	s.R.NoError(err)
	s.R.Equal(http.StatusOK, code)
	s.A.Equal("success", resp.Message)
	s.A.Equal("up", resp.Data.Status)
}
