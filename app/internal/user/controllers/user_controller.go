package controllers

import (
	"net/http"

	"backend/service-platform/app/internal/common/models"
	"backend/service-platform/app/internal/managers"
	"backend/service-platform/app/internal/platform/runtime"
	"backend/service-platform/app/internal/user/constants/currency"
	txconst "backend/service-platform/app/internal/user/constants/transaction"
	usermodels "backend/service-platform/app/internal/user/models"
	"backend/service-platform/app/pkg/jwt"

	"github.com/labstack/echo/v4"
)

type UserController struct {
	res      runtime.Resource
	managers *managers.Managers
	jwt      jwt.Jwt
}

func NewUserController(m *managers.Managers, res runtime.Resource) *UserController {
	return &UserController{res: res, managers: m, jwt: jwt.NewJwt(res.Config.JwtConfig)}
}

func (c *UserController) GetBalances(ec echo.Context) error {
	claims, err := c.jwt.GetClaims(ec)
	if err != nil || claims.UserID == nil {
		return ec.JSON(http.StatusUnauthorized, models.ToErrorResponse(http.StatusUnauthorized, "Authentication required"))
	}

	currencyParam := ec.QueryParam("currency")

	if currencyParam != "" {
		cur := currency.Currency(currencyParam)
		if cur != currency.COIN && cur != currency.SPIN {
			return ec.JSON(http.StatusBadRequest, models.ToErrorResponse(http.StatusBadRequest, "invalid currency"))
		}
		items, err := c.managers.UserBalanceManager.GetBalanceByCurrency(ec.Request().Context(), *claims.UserID, cur)
		if err != nil {
			return ec.JSON(http.StatusInternalServerError, models.ToErrorResponse(http.StatusInternalServerError, "Internal server error"))
		}
		return ec.JSON(http.StatusOK, models.ToSuccessResponse(items))
	}

	items, err := c.managers.UserBalanceManager.GetAllBalances(ec.Request().Context(), *claims.UserID)
	if err != nil {
		return ec.JSON(http.StatusInternalServerError, models.ToErrorResponse(http.StatusInternalServerError, "Internal server error"))
	}
	return ec.JSON(http.StatusOK, models.ToSuccessResponse(items))
}

func (c *UserController) GetBalanceHistory(ec echo.Context) error {
	claims, err := c.jwt.GetClaims(ec)
	if err != nil || claims.UserID == nil {
		return ec.JSON(http.StatusUnauthorized, models.ToErrorResponse(http.StatusUnauthorized, "Authentication required"))
	}

	currencyParam := ec.QueryParam("currency")
	typeParam := ec.QueryParam("type")

	var curPtr *currency.Currency
	if currencyParam == "" {
		return ec.JSON(http.StatusBadRequest, models.ToErrorResponse(http.StatusBadRequest, "currency is required"))
	}
	cur := currency.Currency(currencyParam)
	if cur != currency.COIN && cur != currency.SPIN {
		return ec.JSON(http.StatusBadRequest, models.ToErrorResponse(http.StatusBadRequest, "invalid currency"))
	}
	curPtr = &cur

	var typePtr *string
	if typeParam != "" {
		t := typeParam
		typePtr = &t
	}

	var typedTypePtr *txconst.TransactionType
	if typePtr != nil {
		tt := txconst.TransactionType(*typePtr)
		typedTypePtr = &tt
	}

	items, err := c.managers.UserBalanceManager.GetHistory(ec.Request().Context(), *claims.UserID, curPtr, typedTypePtr)
	if err != nil {
		return ec.JSON(http.StatusInternalServerError, models.ToErrorResponse(http.StatusInternalServerError, "Internal server error"))
	}
	dtos := make([]usermodels.BalanceTransactionResponse, 0, len(items))
	for _, it := range items {
		var meta map[string]interface{}
		if it.Metadata != nil {
			meta = map[string]interface{}(it.Metadata)
		}
		dtos = append(dtos, usermodels.BalanceTransactionResponse{
			ID:            it.ID,
			Amount:        it.Amount,
			Currency:      string(it.Currency),
			Type:          string(it.Type),
			Source:        string(it.Source),
			Status:        string(it.Status),
			Metadata:      meta,
			BalanceBefore: it.BalanceBefore,
			BalanceAfter:  it.BalanceAfter,
			CreatedAt:     it.CreatedAt,
		})
	}
	return ec.JSON(http.StatusOK, models.ToSuccessResponse(dtos))
}
