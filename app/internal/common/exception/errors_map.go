package exception

import (
	"errors"
	"fmt"
	"strings"
)

type ErrorCode int

const (
	ErrCodeNoError                   ErrorCode = iota // 0
	ErrorCodeEntityNotFound                           // 1
	ErrorCodeFailedBindingData                        // 2
	ErrorCodeValidationFailed                         // 3
	ErrorCodeInvalidOtpCode                           // 4
	ErrorCodeUnauthorized                             // 5
	ErrorCodeCodeRateLimitExceeded                    // 6
	ErrorCodeCodeClient                               // 7
	ErrorCodeInvalidParameter                         // 8
	ErrorCodeMissingUserContext                       // 9
	ErrorCodeInvalidOnboardingStatus                  // 10
	ErrorCodeTokenExpired                             // 11
	ErrorCodeInvalidToken                             // 12
	ErrorCodeInternalServer                           // 13
	ErrorCodeFeatureIsUnavailable                     // 14
)

var (
	ErrUnauthorized              = errors.New("request is unauthorized")
	ErrEntityNotFound            = errors.New("entity not found")
	ErrFailedBindingData         = errors.New("failed to bind data")
	ErrValidationFailed          = errors.New("validation failed")
	ErrInvalidOtpCode            = errors.New("invalid OTP code")
	ErrCodeRateLimitExceeded     = errors.New("rate limit exceeded")
	ErrCodeClient                = errors.New("client error")
	ErrInvalidParameter          = errors.New("invalid parameter")
	ErrMissingUserContext        = errors.New("missing user context")
	ErrInvalidOnboardingStatus   = errors.New("invalid onboarding status")
	ErrTokenExpired              = errors.New("token expired")
	ErrInvalidToken              = errors.New("invalid token")
	ErrInternalServer            = errors.New("internal server error")
	ErrUserNotFound              = errors.New("user not found")
	ErrInvalidResetPasswordToken = errors.New("invalid reset password token")
	ErrUsernameAlreadyExisted    = errors.New("username already exists")
	ErrEmailNotFound             = errors.New("email not found")
	ErrFeatureIsUnavailable      = errors.New("feature is unavailable")
)

var errorsMap = map[ErrorCode]error{
	ErrorCodeUnauthorized:            ErrUnauthorized,
	ErrorCodeEntityNotFound:          ErrEntityNotFound,
	ErrorCodeFailedBindingData:       ErrFailedBindingData,
	ErrorCodeValidationFailed:        ErrValidationFailed,
	ErrorCodeInvalidOtpCode:          ErrInvalidOtpCode,
	ErrorCodeCodeRateLimitExceeded:   ErrCodeRateLimitExceeded,
	ErrorCodeCodeClient:              ErrCodeClient,
	ErrorCodeInvalidParameter:        ErrInvalidParameter,
	ErrorCodeMissingUserContext:      ErrMissingUserContext,
	ErrorCodeInvalidOnboardingStatus: ErrInvalidOnboardingStatus,
	ErrorCodeTokenExpired:            ErrTokenExpired,
	ErrorCodeInvalidToken:            ErrInvalidToken,
	ErrorCodeInternalServer:          ErrInternalServer,
}

func GetErrorByCode(code ErrorCode) error {
	return errorsMap[code]
}

// ErrorWithContext attaches additional context to an error as key-value pairs.
//
// When to use this method:
//
//   - Use ErrorWithContext when you want to add contextual information (such as IDs, parameters, or state) to an error before returning or logging it.
//   - This is especially useful for debugging, tracing, or when errors are handled at higher levels and you want to know the circumstances under which they occurred.
//   - The context is provided as variadic arguments in key-value pairs (e.g., "userID", 123, "action", "update").
//
// Example:
//
//	err := GetErrorByCode(ErrorCodeUserNotFound)
//	errWithCtx := ErrorWithContext(err, "userID", 123, "operation", "fetchProfile")
//	// errWithCtx.Error() will be: "userID = 123 , operation = fetchProfile: User not found"
//
// Wrong usage example warning:
//   - Do NOT pass an odd number of context arguments; always provide key-value pairs.
//     For example, ErrorWithContext(err, "userID", 123, "operation") will append "missing ctx" as the value for "operation".
//   - Do NOT use this for sensitive data unless you are sure it is safe to log or expose.
func ErrorWithContext(err error, errorContext ...any) error {
	if ctx := formatKeyValuePairs(errorContext); ctx != "" {
		err = fmt.Errorf("%s: %w", ctx, err)
	} else {
		err = fmt.Errorf("%w", err)
	}
	return err
}

// JoinErrors combines two errors into one, optionally attaching context to the new error.
//
// When to use this method:
//   - Use JoinErrors when you want to aggregate multiple errors together, such as when collecting errors from multiple operations.
//   - This is useful for batch operations, multi-step processes, or when you want to return a single error that represents several failures.
//   - The new error can have context attached using key-value pairs, just like ErrorWithContext.
//
// Example:
//
//	err1 := GetErrorByCode(ErrorCodeUserNotFound)
//	err2 := GetErrorByCode(ErrorCodeInvalidParameter)
//	combined := JoinErrors(err1, err2, "param", "userID")
//	// combined.Error() will include both errors and the context for err2.
//
// Wrong usage example warning:
//   - Do NOT pass an odd number of context arguments; always provide key-value pairs.
func JoinErrors(errs error, newErr error, errorContext ...any) error {
	newErr = ErrorWithContext(newErr, errorContext...)
	return errors.Join(errs, newErr)
}

func formatKeyValuePairs(errorContext []any) string {
	if len(errorContext)%2 != 0 {
		errorContext = append(errorContext, "missing ctx")
	}
	pairs := make([]string, 0, len(errorContext)/2)
	for i := 0; i < len(errorContext); i += 2 {
		key := errorContext[i]
		value := errorContext[i+1]
		pairs = append(pairs, fmt.Sprintf("%v = %v", key, value))
	}
	return strings.Join(pairs, " , ")
}
