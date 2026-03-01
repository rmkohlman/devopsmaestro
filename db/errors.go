package db

import "fmt"

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
	_, ok := err.(*ErrNotFound)
	return ok
}
