package controller

import (
	"backend/service-platform/app/internal/runtime"
	"backend/service-platform/app/manager"
)

type Controllers struct {
	AuthController   *AuthController
	HealthController *HealthController
}

func NewControllers(managers *manager.Managers, res runtime.Resource) *Controllers {
	return &Controllers{
		AuthController:   NewAuthController(managers, res),
		HealthController: NewHealthController(managers, res),
	}
}
