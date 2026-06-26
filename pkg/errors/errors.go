package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents a machine-readable error code.
type ErrorCode string

const (
	// General errors
	ErrInternal       ErrorCode = "INTERNAL_ERROR"
	ErrNotFound       ErrorCode = "NOT_FOUND"
	ErrInvalidInput   ErrorCode = "INVALID_INPUT"
	ErrUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrForbidden      ErrorCode = "FORBIDDEN"
	ErrConflict       ErrorCode = "CONFLICT"
	ErrRateLimited    ErrorCode = "RATE_LIMITED"
	ErrServiceUnavail ErrorCode = "SERVICE_UNAVAILABLE"

	// Domain errors
	ErrUserNotFound      ErrorCode = "USER_NOT_FOUND"
	ErrProblemNotFound   ErrorCode = "PROBLEM_NOT_FOUND"
	ErrSubmissionNotFound ErrorCode = "SUBMISSION_NOT_FOUND"
	ErrInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	ErrTokenExpired      ErrorCode = "TOKEN_EXPIRED"
	ErrTokenInvalid      ErrorCode = "TOKEN_INVALID"
	ErrExecutionFailed   ErrorCode = "EXECUTION_FAILED"
	ErrLanguageUnsupported ErrorCode = "LANGUAGE_UNSUPPORTED"
)

// Error is the custom application error type.
type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
	Err     error     `json:"-"`
	Status  int       `json:"-"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error {
	return e.Err
}

// WithDetails adds details to the error.
func (e *Error) WithDetails(details string) *Error {
	return &Error{
		Code:    e.Code,
		Message: e.Message,
		Details: details,
		Err:     e.Err,
		Status:  e.Status,
	}
}

// WithError wraps an existing error.
func (e *Error) WithError(err error) *Error {
	return &Error{
		Code:    e.Code,
		Message: e.Message,
		Err:     err,
		Status:  e.Status,
	}
}

// HTTPStatus returns the HTTP status code for the error.
func (e *Error) HTTPStatus() int {
	if e.Status > 0 {
		return e.Status
	}
	return HTTPStatusCode(e.Code)
}

// HTTPStatusCode maps error codes to HTTP status codes.
func HTTPStatusCode(code ErrorCode) int {
	switch code {
	case ErrInternal:
		return http.StatusInternalServerError
	case ErrNotFound, ErrUserNotFound, ErrProblemNotFound, ErrSubmissionNotFound:
		return http.StatusNotFound
	case ErrInvalidInput, ErrInvalidCredentials, ErrLanguageUnsupported:
		return http.StatusBadRequest
	case ErrUnauthorized, ErrTokenExpired, ErrTokenInvalid:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrConflict:
		return http.StatusConflict
	case ErrRateLimited:
		return http.StatusTooManyRequests
	case ErrServiceUnavail:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// New creates a new Error with the given code and message.
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Status:  HTTPStatusCode(code),
	}
}

// Wrap wraps an error with additional context.
func Wrap(err error, code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
		Status:  HTTPStatusCode(code),
	}
}

// Is checks if the target error has the given code.
func Is(err error, code ErrorCode) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*Error); ok {
		return e.Code == code
	}
	return false
}

// Predefined error constructors

func Internal(message string) *Error {
	return New(ErrInternal, message)
}

func NotFound(message string) *Error {
	return New(ErrNotFound, message)
}

func InvalidInput(message string) *Error {
	return New(ErrInvalidInput, message)
}

func Unauthorized(message string) *Error {
	return New(ErrUnauthorized, message)
}

func Forbidden(message string) *Error {
	return New(ErrForbidden, message)
}

func Conflict(message string) *Error {
	return New(ErrConflict, message)
}

func RateLimited(message string) *Error {
	return New(ErrRateLimited, message)
}

func ServiceUnavailable(message string) *Error {
	return New(ErrServiceUnavail, message)
}
