package resource

import (
	"testing"
)

// mockResource implements Resource for testing
type mockResource struct {
	kind string
	name string
}

func (m *mockResource) GetKind() string { return m.kind }
func (m *mockResource) GetName() string { return m.name }
func (m *mockResource) Validate() error { return nil }

// mockHandler implements Handler for testing
type mockHandler struct {
	kind      string
	resources map[string]*mockResource
}

func newMockHandler(kind string) *mockHandler {
	return &mockHandler{
		kind:      kind,
		resources: make(map[string]*mockResource),
	}
}

func (h *mockHandler) Kind() string { return h.kind }

func (h *mockHandler) Apply(ctx Context, data []byte) (Resource, error) {
	// Simple mock: extract name from YAML-like format
	res := &mockResource{kind: h.kind, name: "test-resource"}
	h.resources[res.name] = res
	return res, nil
}

func (h *mockHandler) Get(ctx Context, name string) (Resource, error) {
	if res, ok := h.resources[name]; ok {
		return res, nil
	}
	return nil, nil
}

func (h *mockHandler) List(ctx Context) ([]Resource, error) {
	result := make([]Resource, 0, len(h.resources))
	for _, res := range h.resources {
		result = append(result, res)
	}
	return result, nil
}

func (h *mockHandler) Delete(ctx Context, name string) error {
	delete(h.resources, name)
	return nil
}

func (h *mockHandler) ToYAML(res Resource) ([]byte, error) {
	return []byte("kind: " + res.GetKind() + "\nname: " + res.GetName()), nil
}

func TestDetectKind(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    string
		wantErr bool
	}{
		{
			name: "NvimPlugin kind",
			yaml: `apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: telescope`,
			want:    "NvimPlugin",
			wantErr: false,
		},
		{
			name: "NvimTheme kind",
			yaml: `apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: catppuccin`,
			want:    "NvimTheme",
			wantErr: false,
		},
		{
			name:    "missing kind",
			yaml:    `apiVersion: devopsmaestro.io/v1`,
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid YAML",
			yaml:    `{{{invalid`,
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty kind",
			yaml:    `kind: ""`,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectKind([]byte(tt.yaml))
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectKind() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DetectKind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry(t *testing.T) {
	// Clear registry before test
	ClearRegistry()
	defer ClearRegistry()

	t.Run("Register and GetHandler", func(t *testing.T) {
		handler := newMockHandler("TestKind")
		Register(handler)

		got := GetHandler("TestKind")
		if got == nil {
			t.Error("GetHandler() returned nil for registered handler")
		}
		if got.Kind() != "TestKind" {
			t.Errorf("GetHandler().Kind() = %v, want TestKind", got.Kind())
		}
	})

	t.Run("GetHandler unknown kind", func(t *testing.T) {
		got := GetHandler("UnknownKind")
		if got != nil {
			t.Error("GetHandler() should return nil for unknown kind")
		}
	})

	t.Run("MustGetHandler unknown kind", func(t *testing.T) {
		_, err := MustGetHandler("UnknownKind")
		if err == nil {
			t.Error("MustGetHandler() should return error for unknown kind")
		}
	})

	t.Run("RegisteredKinds", func(t *testing.T) {
		ClearRegistry()
		Register(newMockHandler("Kind1"))
		Register(newMockHandler("Kind2"))

		kinds := RegisteredKinds()
		if len(kinds) != 2 {
			t.Errorf("RegisteredKinds() returned %d kinds, want 2", len(kinds))
		}
	})

	t.Run("Register duplicate panics", func(t *testing.T) {
		ClearRegistry()
		Register(newMockHandler("DuplicateKind"))

		defer func() {
			if r := recover(); r == nil {
				t.Error("Register() should panic on duplicate kind")
			}
		}()
		Register(newMockHandler("DuplicateKind"))
	})
}

func TestApply(t *testing.T) {
	ClearRegistry()
	defer ClearRegistry()

	handler := newMockHandler("NvimPlugin")
	Register(handler)

	yaml := `kind: NvimPlugin
metadata:
  name: test`

	ctx := Context{}
	res, err := Apply(ctx, []byte(yaml), "test.yaml")
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if res.GetKind() != "NvimPlugin" {
		t.Errorf("Apply() resource kind = %v, want NvimPlugin", res.GetKind())
	}
}

func TestApplyUnknownKind(t *testing.T) {
	ClearRegistry()
	defer ClearRegistry()

	yaml := `kind: UnknownKind
metadata:
  name: test`

	ctx := Context{}
	_, err := Apply(ctx, []byte(yaml), "test.yaml")
	if err == nil {
		t.Error("Apply() should return error for unknown kind")
	}
}
