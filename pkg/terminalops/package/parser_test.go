package pkg

import (
	"strings"
	"testing"
)

func TestParseYAML(t *testing.T) {
	yamlData := `
apiVersion: devopsmaestro.io/v1
kind: TerminalPackage
metadata:
  name: test-package
  description: Test terminal package
  category: test
  tags: [zsh, starship]
spec:
  extends: core
  plugins: [zsh-autosuggestions, zsh-syntax-highlighting]
  prompts: [starship-minimal]
  profiles: [developer]
  enabled: true
`

	pkg, err := ParseYAML([]byte(yamlData))
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	if pkg.Name != "test-package" {
		t.Errorf("Expected name 'test-package', got %s", pkg.Name)
	}

	if pkg.Description != "Test terminal package" {
		t.Errorf("Expected description 'Test terminal package', got %s", pkg.Description)
	}

	if pkg.Extends != "core" {
		t.Errorf("Expected extends 'core', got %s", pkg.Extends)
	}

	if len(pkg.Plugins) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(pkg.Plugins))
	}

	if len(pkg.Tags) != 2 || pkg.Tags[0] != "zsh" || pkg.Tags[1] != "starship" {
		t.Errorf("Expected tags [zsh, starship], got %v", pkg.Tags)
	}
}

func TestParseYAMLWithStringOrSlice(t *testing.T) {
	// Test single string value
	yamlData := `
apiVersion: devopsmaestro.io/v1
kind: TerminalPackage
metadata:
  name: test-package
spec:
  plugins: single-plugin
  prompts: [prompt1, prompt2]
`

	pkg, err := ParseYAML([]byte(yamlData))
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	if len(pkg.Plugins) != 1 || pkg.Plugins[0] != "single-plugin" {
		t.Errorf("Expected plugins [single-plugin], got %v", pkg.Plugins)
	}

	if len(pkg.Prompts) != 2 || pkg.Prompts[0] != "prompt1" || pkg.Prompts[1] != "prompt2" {
		t.Errorf("Expected prompts [prompt1, prompt2], got %v", pkg.Prompts)
	}
}

func TestValidatePackageYAML(t *testing.T) {
	tests := []struct {
		name      string
		yaml      string
		expectErr bool
		errorMsg  string
	}{
		{
			name: "valid package",
			yaml: `
apiVersion: devopsmaestro.io/v1
kind: TerminalPackage
metadata:
  name: test-package
spec:
  plugins: [plugin1]
`,
			expectErr: false,
		},
		{
			name: "missing name",
			yaml: `
apiVersion: devopsmaestro.io/v1
kind: TerminalPackage
metadata: {}
spec:
  plugins: [plugin1]
`,
			expectErr: true,
			errorMsg:  "metadata.name is required",
		},
		{
			name: "invalid kind",
			yaml: `
apiVersion: devopsmaestro.io/v1
kind: WrongKind
metadata:
  name: test-package
spec:
  plugins: [plugin1]
`,
			expectErr: true,
			errorMsg:  "invalid kind",
		},
		{
			name: "self-extends",
			yaml: `
apiVersion: devopsmaestro.io/v1
kind: TerminalPackage
metadata:
  name: test-package
spec:
  extends: test-package
  plugins: [plugin1]
`,
			expectErr: true,
			errorMsg:  "package cannot extend itself",
		},
		{
			name: "empty plugin name",
			yaml: `
apiVersion: devopsmaestro.io/v1
kind: TerminalPackage
metadata:
  name: test-package
spec:
  plugins: ["", "valid-plugin"]
`,
			expectErr: true,
			errorMsg:  "plugin name at index 0 cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseYAML([]byte(tt.yaml))
			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestPackageToYAMLBytes(t *testing.T) {
	pkg := &Package{
		Name:        "test-package",
		Description: "Test package",
		Plugins:     []string{"plugin1"},
		Enabled:     true,
	}

	yamlBytes, err := pkg.ToYAMLBytes()
	if err != nil {
		t.Fatalf("ToYAMLBytes failed: %v", err)
	}

	if len(yamlBytes) == 0 {
		t.Error("Expected non-empty YAML bytes")
	}

	// Verify we can parse it back
	parsedPkg, err := ParseYAML(yamlBytes)
	if err != nil {
		t.Fatalf("Failed to parse generated YAML: %v", err)
	}

	if parsedPkg.Name != pkg.Name {
		t.Errorf("Expected name %s, got %s", pkg.Name, parsedPkg.Name)
	}
}
