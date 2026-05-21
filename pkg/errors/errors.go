package errors

import (
	"errors"
	"net/http"
)

type AppError struct {
	Message string
	Code    int
}

func (e *AppError) Error() string {
	return e.Message
}

var (
	ErrNotFound     = NewError("not found", http.StatusNotFound)
	ErrUnauthorized = NewError("unauthorized", http.StatusUnauthorized)
	ErrBadRequest   = NewError("bad request", http.StatusBadRequest)
	ErrInternal     = NewError("internal server error", http.StatusInternalServerError)
)

func NewError(message string, code int) *AppError {
	return &AppError{Message: message, Code: code}
}

func HTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return http.StatusInternalServerError
}
