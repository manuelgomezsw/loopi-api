package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError represents an application-level error with HTTP status code.
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// Common application errors.
var (
	ErrNotFound          = &AppError{Code: http.StatusNotFound, Message: "resource not found"}
	ErrUnauthorized      = &AppError{Code: http.StatusUnauthorized, Message: "unauthorized"}
	ErrForbidden         = &AppError{Code: http.StatusForbidden, Message: "forbidden"}
	ErrBadRequest        = &AppError{Code: http.StatusBadRequest, Message: "bad request"}
	ErrInternalServer    = &AppError{Code: http.StatusInternalServerError, Message: "internal server error"}
	ErrConflict          = &AppError{Code: http.StatusConflict, Message: "resource already exists"}
	ErrInvalidCredentials = &AppError{Code: http.StatusUnauthorized, Message: "invalid credentials"}
)

// New creates a new AppError with a custom message.
func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Wrap wraps an error with an AppError.
func Wrap(err error, appErr *AppError) *AppError {
	return &AppError{
		Code:    appErr.Code,
		Message: appErr.Message,
		Err:     err,
	}
}

// Is checks if the target error is an AppError with the same code.
func Is(err, target error) bool {
	return errors.Is(err, target)
}
