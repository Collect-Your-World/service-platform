package router

import (
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	platctrl "backend/service-platform/app/internal/platform/controllers"
	echomw "backend/service-platform/app/internal/platform/middleware"
	"backend/service-platform/app/internal/platform/runtime"
	"backend/service-platform/app/internal/platform/validator"
	"backend/service-platform/app/internal/repository"
	ctxutil "backend/service-platform/app/pkg/util/context"
	echoUtil "backend/service-platform/app/pkg/util/echo"
	_ "backend/service-platform/docs"
)

const (
	apiV1BasePath = "/api/v1"
	swaggerPath   = "/api/v1/swagger/*"
	healthPath    = "/health"

	authPrefix  = "/auth"
	usersPrefix = "/users"
)

type Router struct {
	*echo.Echo
	res          runtime.Resource
	vals         *validator.Validators
	middleware   *echomw.Middleware
	controllers  *platctrl.Controllers
	repositories *repository.Repositories
}

func NewRouter(
	res runtime.Resource,
	vals *validator.Validators,
	middleware *echomw.Middleware,
	controllers *platctrl.Controllers,
	repositories *repository.Repositories,
) *Router {
	if controllers == nil {
		panic("controllers cannot be nil")
	}
	if vals == nil {
		panic("validators cannot be nil")
	}

	r := &Router{
		Echo:         echo.New(),
		res:          res,
		vals:         vals,
		middleware:   middleware,
		controllers:  controllers,
		repositories: repositories,
	}

	r.setupEcho()
	r.setupMiddlewares()
	r.setupSwagger()
	r.setupHealthRoutes()
	r.setupRoutes()

	return r
}

func (r *Router) setupEcho() {
	r.Echo.HidePort = true
	r.Echo.HideBanner = true
	r.Echo.Validator = r.vals
}

func (r *Router) setupMiddlewares() {
	r.Echo.Use(echoMiddleware.RequestID())
	r.Echo.Use(echoUtil.SetupCORSMiddleware(r.res))
	r.Echo.Use(echoUtil.SetupLoggerMiddleware(r.res))
}

func (r *Router) setupSwagger() {
	env := ctxutil.GetAppModeFromEnv()
	if env == ctxutil.AppModeDev || env == ctxutil.AppModeLocal {
		r.Echo.Debug = true
		r.Echo.GET(swaggerPath, echoSwagger.WrapHandler)
	}
}

func (r *Router) setupHealthRoutes() {
	r.Echo.GET(healthPath, r.controllers.HealthController.HealthCheck)
}

func (r *Router) setupRoutes() {
	apiGroup := r.Echo.Group(apiV1BasePath)

	r.setupAuthRoutes(apiGroup)
	r.setupUserRoutes(apiGroup)
}

func (r *Router) setupAuthRoutes(apiGroup *echo.Group) {
	authGroup := apiGroup.Group(authPrefix)
	authGroup.POST("/register", r.controllers.AuthController.Register)
	authGroup.POST("/login", r.controllers.AuthController.Login)
	authGroup.POST("/logout", r.controllers.AuthController.Logout)
	authGroup.POST("/refresh-token", r.controllers.AuthController.RefreshToken)
	authGroup.GET("/me", r.controllers.AuthController.Me, r.middleware.RequireAuth())
}

func (r *Router) setupUserRoutes(apiGroup *echo.Group) {
	usersGroup := apiGroup.Group(usersPrefix)
	usersGroup.GET("/balances", r.controllers.UserController.GetBalances, r.middleware.RequireAuth())
	usersGroup.GET("/balances/history", r.controllers.UserController.GetBalanceHistory, r.middleware.RequireAuth())
}
