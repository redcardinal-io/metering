package errors

import (
	"errors"
	"net/http"
)

// MeterError represents domain-specific errors with HTTP status codes
type MeterError struct {
	Code    string
	Message string
	Status  int
}

func (e MeterError) Error() string {
	return e.Message
}

// Common meter errors with appropriate HTTP status codes
var (
	ErrMeterNotFound = MeterError{
		Code:    "METER_NOT_FOUND",
		Message: "meter not found",
		Status:  http.StatusNotFound,
	}

	ErrMeterAlreadyExists = MeterError{
		Code:    "METER_ALREADY_EXISTS",
		Message: "meter with this slug already exists",
		Status:  http.StatusConflict,
	}

	ErrInvalidMeterInput = MeterError{
		Code:    "INVALID_METER_INPUT",
		Message: "invalid meter input",
		Status:  http.StatusBadRequest,
	}

	ErrDatabaseOperation = MeterError{
		Code:    "DATABASE_ERROR",
		Message: "database operation failed",
		Status:  http.StatusInternalServerError,
	}
)

// IsMeterError checks if an error is of type MeterError
func IsMeterError(err error) (MeterError, bool) {
	var meterErr MeterError
	if errors.As(err, &meterErr) {
		return meterErr, true
	}
	return MeterError{}, false
}

// WithMessage returns a new error with an updated message
func (e MeterError) WithMessage(message string) MeterError {
	e.Message = message
	return e
}
