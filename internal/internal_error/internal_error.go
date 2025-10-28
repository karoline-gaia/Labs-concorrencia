package internal_error

import "net/http"

type InternalError struct {
	Message string
	Err     string
	Code    int
}

func (i *InternalError) Error() string {
	return i.Message
}

func NewInternalServerError(message string) *InternalError {
	return &InternalError{
		Message: message,
		Err:     "internal_server_error",
		Code:    http.StatusInternalServerError,
	}
}

func NewBadRequestError(message string) *InternalError {
	return &InternalError{
		Message: message,
		Err:     "bad_request",
		Code:    http.StatusBadRequest,
	}
}

func NewNotFoundError(message string) *InternalError {
	return &InternalError{
		Message: message,
		Err:     "not_found",
		Code:    http.StatusNotFound,
	}
}
