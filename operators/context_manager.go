package operators

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ContextManager defines the interface for managing the active app/workspace context (kubectl-style).
// This enables DI and testing by allowing mock implementations.
type ContextManager interface {
	// GetActiveApp returns the active app name.
	// Precedence: DVM_APP env var → context.yaml → error.
	GetActiveApp() (string, error)

	// GetActiveWorkspace returns the active workspace name.
	// Precedence: DVM_WORKSPACE env var → context.yaml → error.
	GetActiveWorkspace() (string, error)

	// SetApp sets the active app and clears the workspace.
	SetApp(appName string) error

	// SetWorkspace sets the active workspace (requires active app).
	SetWorkspace(workspaceName string) error

	// ClearApp clears the active app and workspace.
	ClearApp() error

	// ClearWorkspace clears the active workspace (keeps app).
	ClearWorkspace() error

	// LoadContext loads the context from the YAML file.
	LoadContext() (*ContextConfig, error)

	// SaveContext saves the context to the YAML file.
	SaveContext(ctx *ContextConfig) error

	// GetContextSummary returns a human-readable summary of the current context.
	GetContextSummary() (string, error)
}

// Compile-time interface compliance check
var _ ContextManager = (*DefaultContextManager)(nil)

// DefaultContextManager is the standard implementation of ContextManager
// that persists context to a YAML file on disk.
type DefaultContextManager struct {
	contextFilePath string
}

// ContextConfig represents the context configuration
type ContextConfig struct {
	CurrentApp       string `yaml:"current_app"`
	CurrentWorkspace string `yaml:"current_workspace"`
}

// NewContextManager creates a new context manager.
// Returns the ContextManager interface for loose coupling.
func NewContextManager() (ContextManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	dvmDir := filepath.Join(homeDir, ".devopsmaestro")
	contextPath := filepath.Join(dvmDir, "context.yaml")

	return &DefaultContextManager{
		contextFilePath: contextPath,
	}, nil
}

// GetActiveApp returns the active app name
// Precedence: DVM_APP env var → context.yaml → error
func (cm *DefaultContextManager) GetActiveApp() (string, error) {
	// Check environment variable first
	if app := os.Getenv("DVM_APP"); app != "" {
		return app, nil
	}

	// Check context file
	ctx, err := cm.LoadContext()
	if err != nil {
		return "", fmt.Errorf("no active app context (use 'dvm use app <name>' or set DVM_APP)")
	}

	if ctx.CurrentApp == "" {
		return "", fmt.Errorf("no active app context (use 'dvm use app <name>')")
	}

	return ctx.CurrentApp, nil
}

// GetActiveWorkspace returns the active workspace name
// Precedence: DVM_WORKSPACE env var → context.yaml → error
func (cm *DefaultContextManager) GetActiveWorkspace() (string, error) {
	// Check environment variable first
	if workspace := os.Getenv("DVM_WORKSPACE"); workspace != "" {
		return workspace, nil
	}

	// Check context file
	ctx, err := cm.LoadContext()
	if err != nil {
		return "", fmt.Errorf("no active workspace context (use 'dvm use workspace <name>' or set DVM_WORKSPACE)")
	}

	if ctx.CurrentWorkspace == "" {
		return "", fmt.Errorf("no active workspace context (use 'dvm use workspace <name>')")
	}

	return ctx.CurrentWorkspace, nil
}

// SetApp sets the active app
func (cm *DefaultContextManager) SetApp(appName string) error {
	ctx, _ := cm.LoadContext() // Ignore error if file doesn't exist
	if ctx == nil {
		ctx = &ContextConfig{}
	}

	ctx.CurrentApp = appName
	// Clear workspace when switching apps
	ctx.CurrentWorkspace = ""

	return cm.SaveContext(ctx)
}

// SetWorkspace sets the active workspace (requires active app)
func (cm *DefaultContextManager) SetWorkspace(workspaceName string) error {
	ctx, err := cm.LoadContext()
	if err != nil || ctx.CurrentApp == "" {
		return fmt.Errorf("no active app (use 'dvm use app <name>' first)")
	}

	ctx.CurrentWorkspace = workspaceName
	return cm.SaveContext(ctx)
}

// ClearApp clears the active app and workspace
func (cm *DefaultContextManager) ClearApp() error {
	ctx := &ContextConfig{
		CurrentApp:       "",
		CurrentWorkspace: "",
	}
	return cm.SaveContext(ctx)
}

// ClearWorkspace clears the active workspace (keeps app)
func (cm *DefaultContextManager) ClearWorkspace() error {
	ctx, err := cm.LoadContext()
	if err != nil {
		return err
	}

	ctx.CurrentWorkspace = ""
	return cm.SaveContext(ctx)
}

// LoadContext loads the context from the YAML file
func (cm *DefaultContextManager) LoadContext() (*ContextConfig, error) {
	data, err := os.ReadFile(cm.contextFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty context if file doesn't exist
			return &ContextConfig{}, nil
		}
		return nil, fmt.Errorf("failed to read context file: %w", err)
	}

	var ctx ContextConfig
	if err := yaml.Unmarshal(data, &ctx); err != nil {
		return nil, fmt.Errorf("failed to parse context file: %w", err)
	}

	return &ctx, nil
}

// SaveContext saves the context to the YAML file
func (cm *DefaultContextManager) SaveContext(ctx *ContextConfig) error {
	// Ensure directory exists
	dir := filepath.Dir(cm.contextFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create context directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(ctx)
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	// Write to file
	if err := os.WriteFile(cm.contextFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write context file: %w", err)
	}

	return nil
}

// GetContextSummary returns a human-readable summary of the current context
func (cm *DefaultContextManager) GetContextSummary() (string, error) {
	ctx, err := cm.LoadContext()
	if err != nil {
		return "", err
	}

	if ctx.CurrentApp == "" {
		return "No active context", nil
	}

	if ctx.CurrentWorkspace == "" {
		return fmt.Sprintf("App: %s (no workspace selected)", ctx.CurrentApp), nil
	}

	return fmt.Sprintf("App: %s | Workspace: %s", ctx.CurrentApp, ctx.CurrentWorkspace), nil
}
