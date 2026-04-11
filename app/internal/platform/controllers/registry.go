package controllers

import (
	authctrl "backend/service-platform/app/internal/auth/controllers"
	"backend/service-platform/app/internal/managers"
	"backend/service-platform/app/internal/platform/runtime"
	userctrl "backend/service-platform/app/internal/user/controllers"
)

type Controllers struct {
	AuthController   *authctrl.AuthController
	HealthController *HealthController
	UserController   *userctrl.UserController
}

func NewControllers(m *managers.Managers, res runtime.Resource) *Controllers {
	return &Controllers{
		AuthController:   authctrl.NewAuthController(m, res),
		HealthController: NewHealthController(m, res),
		UserController:   userctrl.NewUserController(m, res),
	}
}
