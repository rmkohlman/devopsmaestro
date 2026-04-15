package registry

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// ErrBinaryNotInstalled — sentinel error tests
// =============================================================================

func TestErrBinaryNotInstalled_Sentinel(t *testing.T) {
	assert.NotNil(t, ErrBinaryNotInstalled, "ErrBinaryNotInstalled must not be nil")
	assert.Equal(t, "binary not installed", ErrBinaryNotInstalled.Error())
}

func TestErrBinaryNotInstalled_IsMatch_WhenWrapped(t *testing.T) {
	wrapped := fmt.Errorf("%w: squid (install with: brew install squid)", ErrBinaryNotInstalled)

	assert.True(t, errors.Is(wrapped, ErrBinaryNotInstalled),
		"errors.Is must match ErrBinaryNotInstalled when wrapped with %%w")
}

func TestErrBinaryNotInstalled_IsMatch_DirectError(t *testing.T) {
	assert.True(t, errors.Is(ErrBinaryNotInstalled, ErrBinaryNotInstalled),
		"errors.Is must match the sentinel against itself")
}

func TestErrBinaryNotInstalled_IsNoMatch_OtherErrors(t *testing.T) {
	otherErr := fmt.Errorf("binary installed but failed to start")

	assert.False(t, errors.Is(otherErr, ErrBinaryNotInstalled),
		"errors.Is must NOT match an unrelated error against ErrBinaryNotInstalled")
}

func TestErrBinaryNotInstalled_IsNoMatch_ErrBinaryNotFound(t *testing.T) {
	// ErrBinaryNotFound and ErrBinaryNotInstalled are distinct sentinels
	assert.False(t, errors.Is(ErrBinaryNotFound, ErrBinaryNotInstalled),
		"ErrBinaryNotFound must not match ErrBinaryNotInstalled")
	assert.False(t, errors.Is(ErrBinaryNotInstalled, ErrBinaryNotFound),
		"ErrBinaryNotInstalled must not match ErrBinaryNotFound")
}

func TestErrBinaryNotInstalled_IsMatch_MultiLevelWrap(t *testing.T) {
	inner := fmt.Errorf("%w: squid", ErrBinaryNotInstalled)
	outer := fmt.Errorf("registry start failed: %w", inner)

	assert.True(t, errors.Is(outer, ErrBinaryNotInstalled),
		"errors.Is must unwrap through multiple levels to find ErrBinaryNotInstalled")
}
