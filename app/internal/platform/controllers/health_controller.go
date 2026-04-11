package controllers

import (
	"net/http"

	"backend/service-platform/app/internal/common/models"
	"backend/service-platform/app/internal/managers"
	"backend/service-platform/app/internal/platform/runtime"

	"github.com/labstack/echo/v4"
)

type HealthController struct {
	res      runtime.Resource
	managers *managers.Managers
}

func NewHealthController(m *managers.Managers, res runtime.Resource) *HealthController {
	return &HealthController{
		res:      res,
		managers: m,
	}
}

func (c *HealthController) HealthCheck(ec echo.Context) error {
	return ec.JSON(http.StatusOK, models.ToSuccessResponse(models.HealthResponse{
		Status: "up",
	}))
}
