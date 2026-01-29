package builders

import (
	"context"
	"testing"
)

// MockImageBuilder is a mock implementation of ImageBuilder for testing
type MockImageBuilder struct {
	BuildFunc       func(ctx context.Context, opts BuildOptions) error
	ImageExistsFunc func(ctx context.Context) (bool, error)
	CloseFunc       func() error

	// Track calls for verification
	BuildCalled       bool
	ImageExistsCalled bool
	CloseCalled       bool
	LastBuildOpts     BuildOptions
}

func (m *MockImageBuilder) Build(ctx context.Context, opts BuildOptions) error {
	m.BuildCalled = true
	m.LastBuildOpts = opts
	if m.BuildFunc != nil {
		return m.BuildFunc(ctx, opts)
	}
	return nil
}

func (m *MockImageBuilder) ImageExists(ctx context.Context) (bool, error) {
	m.ImageExistsCalled = true
	if m.ImageExistsFunc != nil {
		return m.ImageExistsFunc(ctx)
	}
	return false, nil
}

func (m *MockImageBuilder) Close() error {
	m.CloseCalled = true
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// Verify MockImageBuilder implements ImageBuilder interface
var _ ImageBuilder = (*MockImageBuilder)(nil)

func TestImageBuilderInterface(t *testing.T) {
	t.Run("interface has required methods", func(t *testing.T) {
		// This test verifies at compile time that ImageBuilder has the right methods
		var builder ImageBuilder = &MockImageBuilder{}

		ctx := context.Background()

		// Test Build method
		err := builder.Build(ctx, BuildOptions{})
		if err != nil {
			t.Errorf("Build() unexpected error: %v", err)
		}

		// Test ImageExists method
		exists, err := builder.ImageExists(ctx)
		if err != nil {
			t.Errorf("ImageExists() unexpected error: %v", err)
		}
		if exists {
			t.Error("ImageExists() should return false by default")
		}

		// Test Close method
		err = builder.Close()
		if err != nil {
			t.Errorf("Close() unexpected error: %v", err)
		}
	})
}

func TestBuildOptions(t *testing.T) {
	tests := []struct {
		name string
		opts BuildOptions
	}{
		{
			name: "empty options",
			opts: BuildOptions{},
		},
		{
			name: "with build args",
			opts: BuildOptions{
				BuildArgs: map[string]string{
					"GO_VERSION":   "1.21",
					"NODE_VERSION": "20",
				},
			},
		},
		{
			name: "with target",
			opts: BuildOptions{
				Target: "production",
			},
		},
		{
			name: "with no cache",
			opts: BuildOptions{
				NoCache: true,
			},
		},
		{
			name: "with pull",
			opts: BuildOptions{
				Pull: true,
			},
		},
		{
			name: "all options",
			opts: BuildOptions{
				BuildArgs: map[string]string{"VERSION": "1.0"},
				Target:    "builder",
				NoCache:   true,
				Pull:      true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockImageBuilder{}
			ctx := context.Background()

			err := mock.Build(ctx, tt.opts)
			if err != nil {
				t.Errorf("Build() error = %v", err)
			}

			if !mock.BuildCalled {
				t.Error("Build() was not called")
			}

			// Verify options were passed correctly
			if len(tt.opts.BuildArgs) != len(mock.LastBuildOpts.BuildArgs) {
				t.Errorf("BuildArgs count mismatch: got %d, want %d",
					len(mock.LastBuildOpts.BuildArgs), len(tt.opts.BuildArgs))
			}

			if mock.LastBuildOpts.Target != tt.opts.Target {
				t.Errorf("Target = %q, want %q", mock.LastBuildOpts.Target, tt.opts.Target)
			}

			if mock.LastBuildOpts.NoCache != tt.opts.NoCache {
				t.Errorf("NoCache = %v, want %v", mock.LastBuildOpts.NoCache, tt.opts.NoCache)
			}

			if mock.LastBuildOpts.Pull != tt.opts.Pull {
				t.Errorf("Pull = %v, want %v", mock.LastBuildOpts.Pull, tt.opts.Pull)
			}
		})
	}
}

func TestMockImageBuilder_Tracking(t *testing.T) {
	mock := &MockImageBuilder{}
	ctx := context.Background()

	// Initially nothing should be called
	if mock.BuildCalled {
		t.Error("BuildCalled should be false initially")
	}
	if mock.ImageExistsCalled {
		t.Error("ImageExistsCalled should be false initially")
	}
	if mock.CloseCalled {
		t.Error("CloseCalled should be false initially")
	}

	// Call Build
	mock.Build(ctx, BuildOptions{})
	if !mock.BuildCalled {
		t.Error("BuildCalled should be true after Build()")
	}

	// Call ImageExists
	mock.ImageExists(ctx)
	if !mock.ImageExistsCalled {
		t.Error("ImageExistsCalled should be true after ImageExists()")
	}

	// Call Close
	mock.Close()
	if !mock.CloseCalled {
		t.Error("CloseCalled should be true after Close()")
	}
}

func TestMockImageBuilder_CustomBehavior(t *testing.T) {
	t.Run("custom build function", func(t *testing.T) {
		expectedErr := context.DeadlineExceeded
		mock := &MockImageBuilder{
			BuildFunc: func(ctx context.Context, opts BuildOptions) error {
				return expectedErr
			},
		}

		err := mock.Build(context.Background(), BuildOptions{})
		if err != expectedErr {
			t.Errorf("Build() error = %v, want %v", err, expectedErr)
		}
	})

	t.Run("custom image exists function", func(t *testing.T) {
		mock := &MockImageBuilder{
			ImageExistsFunc: func(ctx context.Context) (bool, error) {
				return true, nil
			},
		}

		exists, err := mock.ImageExists(context.Background())
		if err != nil {
			t.Errorf("ImageExists() error = %v", err)
		}
		if !exists {
			t.Error("ImageExists() should return true")
		}
	})
}
