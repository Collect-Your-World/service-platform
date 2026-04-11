package controllers

import (
	"errors"
	"net/http"

	"backend/service-platform/app/internal/auth/constants/role"
	authmanagers "backend/service-platform/app/internal/auth/managers"
	authmodels "backend/service-platform/app/internal/auth/models"
	"backend/service-platform/app/internal/common/models"
	"backend/service-platform/app/internal/managers"
	"backend/service-platform/app/internal/platform/runtime"
	"backend/service-platform/app/pkg/jwt"
	utilcookie "backend/service-platform/app/pkg/util/cookie"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type AuthController struct {
	res      runtime.Resource
	managers *managers.Managers
	jwt      jwt.Jwt
}

func NewAuthController(m *managers.Managers, res runtime.Resource) *AuthController {
	jwtService := jwt.NewJwt(res.Config.JwtConfig)
	return &AuthController{
		res:      res,
		managers: m,
		jwt:      jwtService,
	}
}

func (c *AuthController) Register(ec echo.Context) error {
	ctx := ec.Request().Context()
	var req authmodels.RegisterRequest
	if err := ec.Bind(&req); err != nil {
		return ec.JSON(http.StatusBadRequest, models.ToErrorResponse(http.StatusBadRequest, "Invalid request"))
	}
	if err := ec.Validate(&req); err != nil {
		return ec.JSON(http.StatusBadRequest, models.ToErrorResponse(http.StatusBadRequest, "Invalid data"))
	}

	if err := c.managers.AuthManager.Register(ctx, req); err != nil {
		if errors.Is(err, authmanagers.ErrEmailAlreadyExists) || errors.Is(err, authmanagers.ErrUsernameAlreadyExisted) {
			return ec.JSON(http.StatusConflict, models.ToErrorResponse(http.StatusConflict, err.Error()))
		}
		return ec.JSON(http.StatusInternalServerError, models.ToErrorResponse(http.StatusInternalServerError, "Internal server error"))
	}
	return ec.JSON(http.StatusOK, models.ToSuccessResponse("registered"))
}

func (c *AuthController) Login(ec echo.Context) error {
	ctx := ec.Request().Context()
	var req authmodels.AuthUserRequest
	if err := ec.Bind(&req); err != nil {
		c.res.Logger.Error("Failed to bind request", zap.Error(err))
		return ec.JSON(http.StatusBadRequest, models.ToErrorResponse(http.StatusBadRequest, "Invalid request format"))
	}

	if err := ec.Validate(&req); err != nil {
		c.res.Logger.Error("Request validation failed", zap.Error(err))
		return ec.JSON(http.StatusBadRequest, models.ToErrorResponse(http.StatusBadRequest, "Invalid request data"))
	}

	res, err := c.managers.AuthManager.Login(ctx, req)
	if err != nil {
		c.res.Logger.Error("Login failed", zap.Error(err))
		if errors.Is(err, authmanagers.ErrInvalidCredentials) {
			return ec.JSON(http.StatusUnauthorized, models.ToErrorResponse(http.StatusUnauthorized, "Invalid credentials"))
		}
		return ec.JSON(http.StatusInternalServerError, models.ToErrorResponse(http.StatusInternalServerError, "Internal server error"))
	}

	ec.SetCookie(utilcookie.NewRefreshTokenCookie(ec.Request(), res.RefreshToken, c.res.Config.JwtConfig.RefreshExpiration))
	res.RefreshToken = ""
	return ec.JSON(http.StatusOK, models.ToSuccessResponse(res))
}

func (c *AuthController) RefreshToken(ec echo.Context) error {
	rtCookie, errCookie := ec.Cookie("refresh_token")
	if errCookie != nil || rtCookie == nil || rtCookie.Value == "" {
		return ec.JSON(http.StatusUnauthorized, models.ToErrorResponse(http.StatusUnauthorized, "Missing refresh token"))
	}
	authResp, err := c.managers.AuthManager.RefreshToken(ec.Request().Context(), authmodels.RefreshTokenRequest{RefreshToken: rtCookie.Value})
	if err != nil {
		c.res.Logger.Error("Token refresh failed", zap.Error(err))
		if errors.Is(err, authmanagers.ErrInvalidRefreshToken) || errors.Is(err, authmanagers.ErrRefreshTokenRevoked) || errors.Is(err, authmanagers.ErrRefreshTokenExpired) {
			return ec.JSON(http.StatusUnauthorized, models.ToErrorResponse(http.StatusUnauthorized, err.Error()))
		}
		return ec.JSON(http.StatusInternalServerError, models.ToErrorResponse(http.StatusInternalServerError, "Internal server error"))
	}
	ec.SetCookie(utilcookie.NewRefreshTokenCookie(ec.Request(), authResp.RefreshToken, c.res.Config.JwtConfig.RefreshExpiration))
	authResp.RefreshToken = ""
	return ec.JSON(http.StatusOK, models.ToSuccessResponse(authResp))
}

func (c *AuthController) Logout(ec echo.Context) error {
	rtCookie, errCookie := ec.Cookie("refresh_token")
	if errCookie != nil || rtCookie == nil || rtCookie.Value == "" {
		return ec.JSON(http.StatusUnauthorized, models.ToErrorResponse(http.StatusUnauthorized, "Missing refresh token"))
	}
	err := c.managers.AuthManager.Logout(ec.Request().Context(), authmodels.LogoutRequest{RefreshToken: rtCookie.Value})
	if err != nil {
		c.res.Logger.Error("Logout failed", zap.Error(err))
		return ec.JSON(http.StatusInternalServerError, models.ToErrorResponse(http.StatusInternalServerError, "Internal server error"))
	}

	ec.SetCookie(utilcookie.ExpireCookie("refresh_token"))
	return ec.JSON(http.StatusOK, models.ToSuccessResponse("Logged out successfully"))
}

func (c *AuthController) Me(ec echo.Context) error {

	claims, err := c.jwt.GetClaims(ec)
	if err != nil {
		c.res.Logger.Error("Failed to get claims", zap.Error(err))
		return ec.JSON(http.StatusUnauthorized, models.ToErrorResponse(http.StatusUnauthorized, "Authentication required"))
	}
	if claims.UserID == nil || claims.Username == nil {
		return ec.JSON(http.StatusUnauthorized, models.ToErrorResponse(http.StatusUnauthorized, "Authentication required"))
	}

	var userRole role.Role
	if claims.Role != nil {
		userRole = role.Role(*claims.Role)
	} else {
		userRole = role.User
	}

	emailVerified := false
	if claims.EmailVerified != nil {
		emailVerified = *claims.EmailVerified
	}

	phoneVerified := false
	if claims.PhoneVerified != nil {
		phoneVerified = *claims.PhoneVerified
	}

	meResponse := authmodels.MeResponse{
		ID:            *claims.UserID,
		Username:      *claims.Username,
		Email:         claims.Email,
		PhoneNumber:   claims.PhoneNumber,
		Role:          userRole,
		EmailVerified: emailVerified,
		PhoneVerified: phoneVerified,
		LastLoginAt:   claims.LastLoginAt,
	}

	return ec.JSON(http.StatusOK, models.ToSuccessResponse(meResponse))
}
