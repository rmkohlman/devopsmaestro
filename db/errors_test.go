package db

import (
	"errors"
	"fmt"
	"testing"
)

// =============================================================================
// ErrNotFound Tests
// =============================================================================

func TestErrNotFound_Error(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		key      interface{}
		want     string
	}{
		{
			name:     "string key",
			resource: "workspace",
			key:      "my-workspace",
			want:     "workspace 'my-workspace' not found",
		},
		{
			name:     "int key",
			resource: "ecosystem",
			key:      42,
			want:     "ecosystem '42' not found",
		},
		{
			name:     "int64 key",
			resource: "domain",
			key:      int64(123),
			want:     "domain '123' not found",
		},
		{
			name:     "empty key",
			resource: "app",
			key:      "",
			want:     "app '' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &ErrNotFound{
				Resource: tt.resource,
				Key:      tt.key,
			}
			if got := err.Error(); got != tt.want {
				t.Errorf("ErrNotFound.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewErrNotFound(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		key      interface{}
		want     string
	}{
		{
			name:     "creates correct error",
			resource: "plugin",
			key:      "nvim-tree",
			want:     "plugin 'nvim-tree' not found",
		},
		{
			name:     "with numeric key",
			resource: "theme",
			key:      99,
			want:     "theme '99' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewErrNotFound(tt.resource, tt.key)
			if err == nil {
				t.Fatal("NewErrNotFound() returned nil")
			}
			if got := err.Error(); got != tt.want {
				t.Errorf("NewErrNotFound().Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "ErrNotFound pointer",
			err:  &ErrNotFound{Resource: "workspace", Key: "test"},
			want: true,
		},
		{
			name: "NewErrNotFound result",
			err:  NewErrNotFound("ecosystem", 1),
			want: true,
		},
		{
			name: "other error",
			err:  errors.New("some other error"),
			want: false,
		},
		{
			name: "fmt.Errorf error",
			err:  fmt.Errorf("workspace not found"),
			want: false,
		},
		{
			name: "wrapped ErrNotFound with fmt.Errorf",
			err:  fmt.Errorf("failed to get: %w", NewErrNotFound("app", "myapp")),
			want: true, // IsNotFound uses errors.As, which supports wrapped errors
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// TestIsNotFound_ErrorsIs tests that IsNotFound supports wrapped errors via errors.As
func TestIsNotFound_ErrorsIs(t *testing.T) {
	baseErr := NewErrNotFound("workspace", "test")
	wrappedErr := fmt.Errorf("operation failed: %w", baseErr)

	// IsNotFound uses errors.As, so it supports wrapped errors
	if !IsNotFound(wrappedErr) {
		t.Error("IsNotFound() should return true for wrapped ErrNotFound (uses errors.As)")
	}

	// errors.Is also works with wrapped errors
	if !errors.Is(wrappedErr, baseErr) {
		t.Error("errors.Is() should work with wrapped ErrNotFound")
	}

	// errors.As extracts the typed error from the wrapping
	var notFoundErr *ErrNotFound
	if !errors.As(wrappedErr, &notFoundErr) {
		t.Error("errors.As() should extract ErrNotFound from wrapped error")
	}
	if notFoundErr.Resource != "workspace" {
		t.Errorf("extracted error has Resource = %q, want %q", notFoundErr.Resource, "workspace")
	}
}

func TestErrNotFound_TypeAssertion(t *testing.T) {
	// Test that type assertion works as expected
	err := NewErrNotFound("domain", 5)

	// Should be able to extract the typed error
	if notFoundErr, ok := err.(*ErrNotFound); !ok {
		t.Error("NewErrNotFound() should return *ErrNotFound")
	} else {
		if notFoundErr.Resource != "domain" {
			t.Errorf("Resource = %q, want %q", notFoundErr.Resource, "domain")
		}
		if notFoundErr.Key != 5 {
			t.Errorf("Key = %v, want %v", notFoundErr.Key, 5)
		}
	}
}
