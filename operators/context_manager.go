package operators

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ContextManager manages the active project/workspace context (kubectl-style)
type ContextManager struct {
	contextFilePath string
}

// ContextConfig represents the context configuration
type ContextConfig struct {
	CurrentProject   string `yaml:"current_project"`
	CurrentWorkspace string `yaml:"current_workspace"`
}

// NewContextManager creates a new context manager
func NewContextManager() (*ContextManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	dvmDir := filepath.Join(homeDir, ".devopsmaestro")
	contextPath := filepath.Join(dvmDir, "context.yaml")

	return &ContextManager{
		contextFilePath: contextPath,
	}, nil
}

// GetActiveProject returns the active project name
// Precedence: DVM_PROJECT env var → context.yaml → error
func (cm *ContextManager) GetActiveProject() (string, error) {
	// Check environment variable first
	if project := os.Getenv("DVM_PROJECT"); project != "" {
		return project, nil
	}

	// Check context file
	ctx, err := cm.LoadContext()
	if err != nil {
		return "", fmt.Errorf("no active project context (use 'dvm use project <name>' or set DVM_PROJECT)")
	}

	if ctx.CurrentProject == "" {
		return "", fmt.Errorf("no active project context (use 'dvm use project <name>')")
	}

	return ctx.CurrentProject, nil
}

// GetActiveWorkspace returns the active workspace name
// Precedence: DVM_WORKSPACE env var → context.yaml → error
func (cm *ContextManager) GetActiveWorkspace() (string, error) {
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

// SetProject sets the active project
func (cm *ContextManager) SetProject(projectName string) error {
	ctx, _ := cm.LoadContext() // Ignore error if file doesn't exist
	if ctx == nil {
		ctx = &ContextConfig{}
	}

	ctx.CurrentProject = projectName
	// Clear workspace when switching projects
	ctx.CurrentWorkspace = ""

	return cm.SaveContext(ctx)
}

// SetWorkspace sets the active workspace (requires active project)
func (cm *ContextManager) SetWorkspace(workspaceName string) error {
	ctx, err := cm.LoadContext()
	if err != nil || ctx.CurrentProject == "" {
		return fmt.Errorf("no active project (use 'dvm use project <name>' first)")
	}

	ctx.CurrentWorkspace = workspaceName
	return cm.SaveContext(ctx)
}

// ClearProject clears the active project and workspace
func (cm *ContextManager) ClearProject() error {
	ctx := &ContextConfig{
		CurrentProject:   "",
		CurrentWorkspace: "",
	}
	return cm.SaveContext(ctx)
}

// ClearWorkspace clears the active workspace (keeps project)
func (cm *ContextManager) ClearWorkspace() error {
	ctx, err := cm.LoadContext()
	if err != nil {
		return err
	}

	ctx.CurrentWorkspace = ""
	return cm.SaveContext(ctx)
}

// LoadContext loads the context from the YAML file
func (cm *ContextManager) LoadContext() (*ContextConfig, error) {
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
func (cm *ContextManager) SaveContext(ctx *ContextConfig) error {
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
func (cm *ContextManager) GetContextSummary() (string, error) {
	ctx, err := cm.LoadContext()
	if err != nil {
		return "", err
	}

	if ctx.CurrentProject == "" {
		return "No active context", nil
	}

	if ctx.CurrentWorkspace == "" {
		return fmt.Sprintf("Project: %s (no workspace selected)", ctx.CurrentProject), nil
	}

	return fmt.Sprintf("Project: %s | Workspace: %s", ctx.CurrentProject, ctx.CurrentWorkspace), nil
}
