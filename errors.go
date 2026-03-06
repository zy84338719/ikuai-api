package ikuaisdk

import "fmt"

type ErrorCode int

const (
	ErrCodeUnknown ErrorCode = iota
	ErrCodeLoginFailed
	ErrCodeNotLoggedIn
	ErrCodeRequestFailed
	ErrCodeInvalidResponse
	ErrCodeVersionNotSupported
	ErrCodeValidationFailed
	ErrCodeRateLimited
	ErrCodeTimeout
	ErrCodeConnectionLost
	ErrCodeUnauthorized
	ErrCodeForbidden
)

type SDKError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *SDKError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[SDK Error %d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[SDK Error %d] %s", e.Code, e.Message)
}

func (e *SDKError) Unwrap() error {
	return e.Cause
}

func NewSDKError(code ErrorCode, message string, cause error) *SDKError {
	return &SDKError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

func IsSDKError(err error) bool {
	_, ok := err.(*SDKError)
	return ok
}

func GetErrorCode(err error) ErrorCode {
	if sdkErr, ok := err.(*SDKError); ok {
		return sdkErr.Code
	}
	return ErrCodeUnknown
}
