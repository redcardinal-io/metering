package errors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

var (
	ErrInternalServer = errors.New("internal server error")
	ErrNotFound       = errors.New("resource not found")
	ErrBadRequest     = errors.New("bad request")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrConflict       = errors.New("conflict")
	ErrValidation     = errors.New("validation error")
	ErrTimeout        = errors.New("request timeout")
	ErrUnavailable    = errors.New("service unavailable")
)

// AppError represents an application error with context
type AppError struct {
	Err        error
	StatusCode int
	Code       string
	Message    string
	Internal   string
	Op         string
	Data       map[string]any
}

func (e *AppError) Error() string {
	return e.Message
}

// Unwrap provides compatibility with errors.Unwrap
func (e *AppError) Unwrap() error {
	return e.Err
}

// ErrorCode represents an error code
type ErrorCode string

// Application error codes
const (
	ECONFLICT      ErrorCode = "conflict"
	EINTERNAL      ErrorCode = "internal"
	EINVALID       ErrorCode = "invalid"
	ENOTFOUND      ErrorCode = "not_found"
	EUNAUTHORIZED  ErrorCode = "unauthorized"
	EFORBIDDEN     ErrorCode = "forbidden"
	ETIMEOUT       ErrorCode = "timeout"
	EUNAVAILABLE   ErrorCode = "unavailable"
	EUNPROCESSABLE ErrorCode = "unprocessable_entity"
	EDATABASE      ErrorCode = "database_error"
	EOLAP          ErrorCode = "olap_database_error"
	EMESSAGEBROKER ErrorCode = "message_broker_error"
	EVALIDATION    ErrorCode = "validation_error"
)

// Map standard errors to HTTP status codes
var errorStatusCodes = map[error]int{
	ErrInternalServer: http.StatusInternalServerError,
	ErrNotFound:       http.StatusNotFound,
	ErrBadRequest:     http.StatusBadRequest,
	ErrUnauthorized:   http.StatusUnauthorized,
	ErrForbidden:      http.StatusForbidden,
	ErrConflict:       http.StatusConflict,
	ErrValidation:     http.StatusUnprocessableEntity,
	ErrTimeout:        http.StatusRequestTimeout,
	ErrUnavailable:    http.StatusServiceUnavailable,
}

// Map error codes to HTTP status codes
var codeStatusCodes = map[ErrorCode]int{
	ECONFLICT:      http.StatusConflict,
	EINTERNAL:      http.StatusInternalServerError,
	EINVALID:       http.StatusBadRequest,
	ENOTFOUND:      http.StatusNotFound,
	EUNAUTHORIZED:  http.StatusUnauthorized,
	EFORBIDDEN:     http.StatusForbidden,
	ETIMEOUT:       http.StatusRequestTimeout,
	EUNAVAILABLE:   http.StatusServiceUnavailable,
	EUNPROCESSABLE: http.StatusUnprocessableEntity,
	EDATABASE:      http.StatusInternalServerError,
	EOLAP:          http.StatusInternalServerError,
	EMESSAGEBROKER: http.StatusInternalServerError,
	EVALIDATION:    http.StatusUnprocessableEntity,
}

// ErrorOption is a function that configures an AppError
type ErrorOption func(*AppError)

// New creates a new application error
func New(err error, code ErrorCode, message string, opts ...ErrorOption) *AppError {
	if err == nil {
		err = ErrInternalServer
	}

	// Default status code based on error code
	statusCode, ok := codeStatusCodes[code]
	if !ok {
		statusCode = http.StatusInternalServerError
	}

	// If err is already an AppError, preserve its fields unless overridden
	var appErr *AppError
	if errors.As(err, &appErr) {
		if message == "" {
			message = appErr.Message
		}
		if statusCode == http.StatusInternalServerError && appErr.StatusCode != 0 {
			statusCode = appErr.StatusCode
		}
	}

	// If message is empty, use error string
	if message == "" {
		message = err.Error()
	}

	data := make(map[string]any)
	if appErr.Data != nil {
		data = appErr.Data
	}

	ae := &AppError{
		Err:        err,
		StatusCode: statusCode,
		Code:       string(code),
		Message:    message,
		Internal:   err.Error(),
		Data:       data,
	}

	// Apply options
	for _, opt := range opts {
		opt(ae)
	}

	return ae
}

// WithStatusCode sets the status code for an error
func WithStatusCode(statusCode int) ErrorOption {
	return func(e *AppError) {
		e.StatusCode = statusCode
	}
}

// WithOperation sets the operation where the error occurred
func WithOperation(op string) ErrorOption {
	return func(e *AppError) {
		e.Op = op
	}
}

// WithData adds additional data to the error
func WithData(key string, value any) ErrorOption {
	return func(e *AppError) {
		e.Data[key] = value
	}
}

// WithInternal sets internal error details (not exposed to clients)
func WithInternal(internal string) ErrorOption {
	return func(e *AppError) {
		e.Internal = internal
	}
}

// GetErrorCode returns the error code from an AppError or an empty string
func GetErrorCode(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ""
}

// GetStatusCode returns the appropriate HTTP status code for an error
func GetStatusCode(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.StatusCode
	}

	// Check if it's a standard error
	if code, ok := errorStatusCodes[err]; ok {
		return code
	}

	if code, ok := codeStatusCodes[ErrorCode(err.Error())]; ok {
		return code
	}

	// Default to internal server error
	return http.StatusInternalServerError
}

// GetMessage returns a user-friendly error message
func GetMessage(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Message
	}
	return err.Error()
}

// Is provides compatibility with errors.Is
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As provides compatibility with errors.As
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Wrap wraps an error with a message
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// WrapWithCode wraps an error with a message and code
func WrapWithCode(err error, code ErrorCode, message string) error {
	if err == nil {
		return nil
	}
	return New(err, code, message)
}

type ErrorResponse struct {
	Status  int            `json:"-"`
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// NewErrorResponse creates a new error response from an error
func NewErrorResponse(err error) ErrorResponse {
	var appErr *AppError
	if errors.As(err, &appErr) {
		details := appErr.Data
		if details == nil {
			details = make(map[string]any)
		}

		statusCode := GetStatusCode(err)

		// Add operation for easier debugging if not empty
		if appErr.Op != "" {
			details["operation"] = appErr.Op
		}

		return ErrorResponse{
			Status:  statusCode,
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: details,
		}
	}

	// Handle non-AppError errors
	statusCode := GetStatusCode(err)
	code := "internal"

	if statusCode == http.StatusNotFound {
		code = "not_found"
	} else if statusCode == http.StatusBadRequest {
		code = "invalid"
	} else if statusCode == http.StatusUnauthorized {
		code = "unauthorized"
	} else if statusCode == http.StatusForbidden {
		code = "forbidden"
	}

	return ErrorResponse{
		Status:  statusCode,
		Code:    code,
		Message: err.Error(),
	}
}

func (e *ErrorResponse) ToJson() fiber.Map {
	return fiber.Map{
		"status":  e.Status,
		"code":    e.Code,
		"message": e.Message,
	}
}

func NewErrorResponseWithOpts(err error, code ErrorCode, message string, opts ...ErrorOption) ErrorResponse {
    appErr := New(err, code, message, opts...)
    return NewErrorResponse(appErr)
}
