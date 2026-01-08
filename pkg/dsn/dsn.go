// Package dsn provides interfaces and utilities for building and validating
// database connection strings (Data Source Names) for various database drivers.
package dsn

import "fmt"

var (
	// ErrMissingHost is the error message when host is not provided.
	ErrMissingHost = "is required"
	// ErrMissingPort is the error message when port is not provided.
	ErrMissingPort = "is required"
	// ErrMissingUser is the error message when user is not provided.
	ErrMissingUser = "is required"
	// ErrMissingPassword is the error message when password is not provided.
	ErrMissingPassword = "is required"
	// ErrMissingDatabase is the error message when database is not provided.
	ErrMissingDatabase = "is required"
	// ErrInvalidPort is the error message when port is outside valid range.
	ErrInvalidPort = "must be between 1 and 65535"
)

// Builder is an interface for building and validating database connection strings.
type Builder interface {
	// ConnectionString returns the formatted connection string for the database driver.
	ConnectionString() (string, error)
	// Validate checks if all required fields are properly set.
	Validate() error
	// Driver returns the name of the database driver.
	Driver() string
}

// ValidationError represents a validation error for a DSN field.
type ValidationError struct {
	// Driver is the name of the database driver
	Driver string
	// Field is the name of the field that failed validation
	Field string
	// Message describes the validation error
	Message string
}

// Error implements the error interface and returns a formatted error message.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s %s", e.Driver, e.Field, e.Message)
}

// NewValidationError creates a new ValidationError with the specified driver, field, and message.
func NewValidationError(driver, field, message string) *ValidationError {
	return &ValidationError{
		Driver:  driver,
		Field:   field,
		Message: message,
	}
}
