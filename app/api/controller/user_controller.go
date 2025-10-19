package controller

import (
	"backend/service-platform/app/api/client/response"
	"backend/service-platform/app/database/constant/currency"
	txconst "backend/service-platform/app/database/constant/transaction"
	"backend/service-platform/app/internal/runtime"
	"backend/service-platform/app/manager"
	"backend/service-platform/app/pkg/jwt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type UserController struct {
	res      runtime.Resource
	managers *manager.Managers
	jwt      jwt.Jwt
}

func NewUserController(managers *manager.Managers, res runtime.Resource) *UserController {
	return &UserController{res: res, managers: managers, jwt: jwt.NewJwt(res.Config.JwtConfig)}
}

func (c *UserController) GetBalances(ec echo.Context) error {
	claims, err := c.jwt.GetClaims(ec)
	if err != nil || claims.UserID == nil {
		return ec.JSON(http.StatusUnauthorized, response.ToErrorResponse(http.StatusUnauthorized, "Authentication required"))
	}

	currencyParam := ec.QueryParam("currency")

	if currencyParam != "" {
		cur := currency.Currency(currencyParam)
		if cur != currency.COIN && cur != currency.SPIN {
			return ec.JSON(http.StatusBadRequest, response.ToErrorResponse(http.StatusBadRequest, "invalid currency"))
		}
		items, err := c.managers.UserBalanceManager.GetBalanceByCurrency(ec.Request().Context(), *claims.UserID, cur)
		if err != nil {
			return ec.JSON(http.StatusInternalServerError, response.ToErrorResponse(http.StatusInternalServerError, "Internal server error"))
		}
		return ec.JSON(http.StatusOK, response.ToSuccessResponse(items))
	}

	items, err := c.managers.UserBalanceManager.GetAllBalances(ec.Request().Context(), *claims.UserID)
	if err != nil {
		return ec.JSON(http.StatusInternalServerError, response.ToErrorResponse(http.StatusInternalServerError, "Internal server error"))
	}
	return ec.JSON(http.StatusOK, response.ToSuccessResponse(items))
}

func (c *UserController) GetBalanceHistory(ec echo.Context) error {
	claims, err := c.jwt.GetClaims(ec)
	if err != nil || claims.UserID == nil {
		return ec.JSON(http.StatusUnauthorized, response.ToErrorResponse(http.StatusUnauthorized, "Authentication required"))
	}

	currencyParam := ec.QueryParam("currency")
	typeParam := ec.QueryParam("type")

	var curPtr *currency.Currency
	if currencyParam == "" {
		return ec.JSON(http.StatusBadRequest, response.ToErrorResponse(http.StatusBadRequest, "currency is required"))
	}
	cur := currency.Currency(currencyParam)
	if cur != currency.COIN && cur != currency.SPIN {
		return ec.JSON(http.StatusBadRequest, response.ToErrorResponse(http.StatusBadRequest, "invalid currency"))
	}
	curPtr = &cur

	var typePtr *string
	if typeParam != "" {
		t := typeParam
		typePtr = &t
	}

	// Convert string type to typed enum if provided
	var typedTypePtr *txconst.Type
	if typePtr != nil {
		tt := txconst.Type(*typePtr)
		typedTypePtr = &tt
	}

	items, err := c.managers.UserBalanceManager.GetHistory(ec.Request().Context(), *claims.UserID, curPtr, typedTypePtr)
	if err != nil {
		return ec.JSON(http.StatusInternalServerError, response.ToErrorResponse(http.StatusInternalServerError, "Internal server error"))
	}
	dtos := make([]response.BalanceTransactionResponse, 0, len(items))
	for _, it := range items {
		dtos = append(dtos, response.BalanceTransactionResponse{
			ID:        it.ID,
			Amount:    it.Amount,
			Currency:  it.Currency,
			Type:      it.Type,
			Source:    it.Source,
			Status:    it.Status,
			CreatedAt: it.CreatedAt,
		})
	}
	return ec.JSON(http.StatusOK, response.ToSuccessResponse(dtos))
}
