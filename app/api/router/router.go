package router

import (
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	"backend/service-platform/app/api/controller"
	"backend/service-platform/app/api/middleware"
	"backend/service-platform/app/database/repository"
	"backend/service-platform/app/internal/runtime"
	"backend/service-platform/app/internal/validator"
	ctxutil "backend/service-platform/app/pkg/util/context"
	echoUtil "backend/service-platform/app/pkg/util/echo"
	_ "backend/service-platform/docs"
)

const (
	// Base paths
	apiV1BasePath = "/api/v1"
	swaggerPath   = "/swagger/*"
	healthPath    = "/health"

	// Route prefixes
	authPrefix = "/auth"
)

type Router struct {
	*echo.Echo
	res          runtime.Resource
	vals         *validator.Validators
	middleware   *middleware.Middleware
	controllers  *controller.Controllers
	repositories *repository.Repositories
}

// NewRouter @title Stack
// @description This is API documentation of Stack
// @version 1.0
// @host localhost:8081
// @BasePath /api/v1
func NewRouter(
	res runtime.Resource,
	vals *validator.Validators,
	middleware *middleware.Middleware,
	controllers *controller.Controllers,
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
}

func (r *Router) setupAuthRoutes(apiGroup *echo.Group) {
	authGroup := apiGroup.Group(authPrefix)
	authGroup.POST("/register", r.controllers.AuthController.Register)
	authGroup.POST("/login", r.controllers.AuthController.Login)
	authGroup.POST("/logout", r.controllers.AuthController.Logout)
	authGroup.POST("/refresh", r.controllers.AuthController.RefreshToken)
	authGroup.GET("/me", r.controllers.AuthController.Me, r.middleware.RequireAuth())
}
