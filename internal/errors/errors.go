package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Standard application error codes as required by API specification.
const (
	CodeValidationError       = "VALIDATION_ERROR"
	CodeEmployeeNotFound      = "EMPLOYEE_NOT_FOUND"
	CodeEmployeeEmailConflict = "EMPLOYEE_EMAIL_CONFLICT"
	CodeDatabaseError         = "DATABASE_ERROR"
	CodeInternalServerError   = "INTERNAL_SERVER_ERROR"
)

// Common repository sentinel errors.
var (
	ErrNotFound       = errors.New("record not found")
	ErrDuplicateEmail = errors.New("employee email already exists")
)

// ErrorDetail represents the structured error response payload.
type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// ErrorResponse represents the outer JSON envelope for all API error responses.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// AppError represents a typed application error that carries HTTP status,
// client-safe error code, user message, optional field details, and the internal underlying cause.
type AppError struct {
	StatusCode int
	Code       string
	Message    string
	Details    map[string]string
	Err        error // Internal underlying error, never exposed directly to clients
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// ResponsePayload returns the sanitized ErrorResponse safe for client consumption.
func (e *AppError) ResponsePayload() ErrorResponse {
	return ErrorResponse{
		Error: ErrorDetail{
			Code:    e.Code,
			Message: e.Message,
			Details: e.Details,
		},
	}
}

// NewValidationError creates a 400 Bad Request error with field validation details.
func NewValidationError(message string, details map[string]string) *AppError {
	return &AppError{
		StatusCode: http.StatusBadRequest,
		Code:       CodeValidationError,
		Message:    message,
		Details:    details,
	}
}

// NewNotFoundError creates a 404 Not Found error.
func NewNotFoundError(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusNotFound,
		Code:       CodeEmployeeNotFound,
		Message:    message,
	}
}

// NewConflictError creates a 409 Conflict error for duplicate email or resource conflicts.
func NewConflictError(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusConflict,
		Code:       CodeEmployeeEmailConflict,
		Message:    message,
	}
}

// NewDatabaseError creates a 500 Internal Server Error masking internal DB details from clients.
func NewDatabaseError(err error) *AppError {
	return &AppError{
		StatusCode: http.StatusInternalServerError,
		Code:       CodeDatabaseError,
		Message:    "A database error occurred",
		Err:        err,
	}
}

// NewInternalError creates a 500 Internal Server Error masking internal details from clients.
func NewInternalError(err error) *AppError {
	return &AppError{
		StatusCode: http.StatusInternalServerError,
		Code:       CodeInternalServerError,
		Message:    "An unexpected error occurred",
		Err:        err,
	}
}
