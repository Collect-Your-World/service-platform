package validator

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"

	"backend/service-platform/app/api/client/response"
	"backend/service-platform/app/internal/runtime"
)

type IValidator interface {
	Register() (validator.Func, string)
}

type Validators struct {
	v          *validator.Validate
	validators []IValidator
}

func NewValidators(res runtime.Resource) *Validators {
	validators := []IValidator{
		NewNotBlankValidator(),
	}

	v := &Validators{
		v:          validator.New(),
		validators: validators,
	}

	// Setup all validators
	if err := v.Setup(); err != nil {
		panic(err)
	}

	return v
}

func (vl *Validators) Setup() error {
	for _, v := range vl.validators {
		fnc, tag := v.Register()
		if err := vl.v.RegisterValidation(tag, fnc); err != nil {
			return err
		}
	}
	return nil
}

func (vl *Validators) Validate(requestData any) error {
	if err := vl.v.Struct(requestData); err != nil {
		// Verify if the error is a validation error
		var validationErrs validator.ValidationErrors
		if ok := errors.As(err, &validationErrs); ok {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				response.GeneralResponse[any]{
					Code:         1,
					Message:      "Validation failed",
					ErrorDetails: getDetails(validationErrs),
				}).WithInternal(err)
		}
	}
	return nil
}

func getDetails(validationErrs validator.ValidationErrors) (out []response.ErrorDetail) {
	for _, vErr := range validationErrs {
		out = append(out, response.ErrorDetail{
			Key:     vErr.Namespace(),
			Field:   vErr.Field(),
			Message: fmt.Sprintf("Failed on the '%s' tag", vErr.Tag()),
		})
	}

	return out
}

var _ echo.Validator = &Validators{}
