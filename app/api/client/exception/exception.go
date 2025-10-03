package exception

import (
	"errors"
	"net/http"

	"backend/service-platform/app/api/client/response"

	"github.com/labstack/echo/v4"
)

type ErrorModel response.GeneralResponse[any]

// WrapError use to wrap any error into echo.HTTPError
// This method can be used widely in both repositories and services
// Parameters:
// - err: The original error that occurred. If you want to add stacktrace, please use fmt.Errorf("Your error message", err)
// - message[0]: Optional additional message returned to the client
func WrapError(err error, message string, data ...any) *echo.HTTPError {
	var (
		httpError *echo.HTTPError
		ok        bool
	)
	if err == nil {
		httpError = echo.NewHTTPError(http.StatusInternalServerError)
	} else {
		if !errors.As(err, &httpError) {
			httpError = echo.NewHTTPError(http.StatusInternalServerError).WithInternal(err)
		}
	}
	msg := httpError.Message
	if msg == nil {
		msg = &ErrorModel{}
	}
	errMsg, ok := msg.(*ErrorModel)
	if !ok {
		if orgMsg, ok := msg.(string); ok {
			errMsg = &ErrorModel{
				Message: orgMsg,
			}
		} else {
			errMsg = &ErrorModel{}
		}
	}

	if len(data) > 0 {
		errMsg.Data = data[0]
	}

	if message != "" {
		errMsg.Message = message
	} else {
		errMsg.Message = http.StatusText(httpError.Code)
	}
	httpError.Message = errMsg

	return httpError
}

// NewError use to create an echo.HTTPError for client response from an error
// This method is only used in services/managers layer to transfer error information to clients
// Althought its name is NewError, it reserves the original echo.HTTPError if the error is already an echo.HTTPError
//
// Parameters:
// - err: The original error that occurred. If you want to add stacktrace, please use errors.Wrap(err, "Your error message")
// - httpCode: The HTTP status code to be associated with the error
// - appCode: The application-specific error code in exception.go
// - messages[0]: Optional additional message returned to the client
func NewError(err error, httpCode int, appCode int, message string, data ...any) *echo.HTTPError {
	var httpError = WrapError(err, message, data...)
	httpError.Code = appCode
	msgErr := httpError.Message.(*ErrorModel)
	msgErr.Code = appCode
	if message != "" {
		msgErr.Message = message
	}
	httpError.Code = httpCode
	return httpError
}

func NewInternalServerError(err error, appCode int, message string, data ...any) error {
	return NewError(err, http.StatusInternalServerError, appCode, message, data...)
}

func NewBadRequestError(err error, appCode int, message string, data ...any) error {
	return NewError(err, http.StatusBadRequest, appCode, message, data...)
}

func NewUnauthorizedError(err error, appCode int, message string, data ...any) error {
	return NewError(err, http.StatusUnauthorized, appCode, message, data...)
}

func NewForbiddenError(err error, appCode int, message string, data ...any) error {
	return NewError(err, http.StatusForbidden, appCode, message, data...)
}

func NewNotFoundError(err error, appCode int, message string, data ...any) error {
	return NewError(err, http.StatusNotFound, appCode, message, data...)
}

func NewConflictError(err error, appCode int, message string, data ...any) error {
	return NewError(err, http.StatusConflict, appCode, message, data...)
}

func NewTooManyRequestsError(err error, appCode int, message string, data ...any) error {
	return NewError(err, http.StatusTooManyRequests, appCode, message, data...)
}

func NewUnavailableForLegalReasonsError(err error, appCode int, message string, data ...any) error {
	return NewError(err, http.StatusUnavailableForLegalReasons, appCode, message, data...)
}

func NewAcceptedButNotProcessedError(err error, appCode int, message string, data ...any) error {
	return NewError(err, http.StatusAccepted, appCode, message, data...)
}

func NewFeatureIsUnavailableError() error {
	return NewError(nil, http.StatusNotFound, int(ErrorCodeFeatureIsUnavailable), ErrFeatureIsUnavailable.Error())
}
