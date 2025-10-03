package controller

import (
	"net/http"

	"backend/service-platform/app/api/client/response"
	"backend/service-platform/app/internal/runtime"
	"backend/service-platform/app/manager"

	"github.com/labstack/echo/v4"
)

type HealthController struct {
	res      runtime.Resource
	managers *manager.Managers
}

func NewHealthController(managers *manager.Managers, res runtime.Resource) *HealthController {
	return &HealthController{
		res:      res,
		managers: managers,
	}
}

// HealthCheck godoc
//
//	@Summary		Verify health
//	@Description	Verify health for BE service
//	@Tags			system
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.HealthResponse
//	@Failure		500
//	@Router			/health [get]
func (c *HealthController) HealthCheck(ec echo.Context) error {
	return ec.JSON(http.StatusOK, response.ToSuccessResponse(response.HealthResponse{
		Status: "up",
	}))
}
