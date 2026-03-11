package db

import (
	"errors"
	"fmt"
)

// ErrNotFound indicates the requested resource does not exist.
type ErrNotFound struct {
	Resource string
	Key      interface{}
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("%s '%v' not found", e.Resource, e.Key)
}

// NewErrNotFound creates a new ErrNotFound error.
func NewErrNotFound(resource string, key interface{}) error {
	return &ErrNotFound{Resource: resource, Key: key}
}

// IsNotFound checks if an error is an ErrNotFound.
func IsNotFound(err error) bool {
	var target *ErrNotFound
	return errors.As(err, &target)
}

// ErrUniqueViolation represents a database unique constraint violation.
type ErrUniqueViolation struct {
	Column string
	Value  string
}

func (e *ErrUniqueViolation) Error() string {
	if e.Column != "" {
		return fmt.Sprintf("unique constraint violation on %s: %s", e.Column, e.Value)
	}
	return "unique constraint violation"
}

// NewErrUniqueViolation creates a new unique constraint violation error.
func NewErrUniqueViolation(column, value string) *ErrUniqueViolation {
	return &ErrUniqueViolation{Column: column, Value: value}
}

// IsUniqueViolation checks if an error is a unique constraint violation.
func IsUniqueViolation(err error) bool {
	var target *ErrUniqueViolation
	return errors.As(err, &target)
}
