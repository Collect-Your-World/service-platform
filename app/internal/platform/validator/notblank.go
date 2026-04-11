package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type NotBlankValidator struct{}

func NewNotBlankValidator() IValidator {
	return &NotBlankValidator{}
}

func (v *NotBlankValidator) Register() (validator.Func, string) {
	return func(fl validator.FieldLevel) bool {
		field := fl.Field()
		switch field.Kind() {
		case reflect.String:
			return strings.TrimSpace(field.String()) != ""
		default:
			return true
		}
	}, "notblank"
}
