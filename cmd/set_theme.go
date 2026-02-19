// Package cmd provides CLI commands for theme management.
// This file implements the 'dvm set theme' command for hierarchical theme setting.
package cmd

import (
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/colors/resolver"
	"devopsmaestro/pkg/nvimops/theme/library"
	"devopsmaestro/pkg/resource"
	"devopsmaestro/pkg/resource/handlers"
	"devopsmaestro/render"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Flags for set theme command
var (
	setThemeEcosystem   string
	setThemeDomain      string
	setThemeApp         string
	setThemeWorkspace   string
	setThemeOutput      string
	setThemeDryRun      bool
	setThemeShowCascade bool
)

// ThemeSetResult represents the result of setting a theme
type ThemeSetResult struct {
	Level          string            `yaml:"level" json:"level"`
	ObjectName     string            `yaml:"objectName" json:"objectName"`
	Theme          string            `yaml:"theme" json:"theme"`
	PreviousTheme  string            `yaml:"previousTheme,omitempty" json:"previousTheme,omitempty"`
	EffectiveTheme string            `yaml:"effectiveTheme" json:"effectiveTheme"`
	CascadeInfo    *ThemeCascadeInfo `yaml:"cascadeInfo,omitempty" json:"cascadeInfo,omitempty"`
}

// ThemeCascadeInfo contains information about theme cascade effects
type ThemeCascadeInfo struct {
	AffectedLevels []string      `yaml:"affectedLevels" json:"affectedLevels"`
	ResolutionPath []CascadeStep `yaml:"resolutionPath" json:"resolutionPath"`
}

// CascadeStep represents one step in the theme resolution path
type CascadeStep struct {
	Level    string `yaml:"level" json:"level"`
	Name     string `yaml:"name" json:"name"`
	Theme    string `yaml:"theme,omitempty" json:"theme,omitempty"`
	HasTheme bool   `yaml:"hasTheme" json:"hasTheme"`
	Error    string `yaml:"error,omitempty" json:"error,omitempty"`
}

// setThemeCmd sets theme at hierarchy level
var setThemeCmd = &cobra.Command{
	Use:   "theme <theme-name>",
	Short: "Set theme at hierarchy level",
	Long: `Set Neovim theme at ecosystem, domain, app, or workspace level.

Themes cascade down the hierarchy unless overridden:
  Ecosystem → Domain → App → Workspace

Use empty string "" to clear override and inherit from parent level.

Examples:
  dvm set theme coolnight-synthwave --workspace dev
  dvm set theme tokyonight-night --app my-api  
  dvm set theme "" --workspace dev  # clear, inherit from app
  dvm set theme gruvbox-dark --domain auth --ecosystem platform

Available themes:
  Library themes (34+ available instantly): coolnight-ocean, tokyonight-night, catppuccin-mocha, etc.
  Use 'dvm get nvim themes' to see all available themes (user + library).`,
	Args: cobra.ExactArgs(1),
	RunE: runSetTheme,
}

func init() {
	setCmd.AddCommand(setThemeCmd)

	// Add hierarchy level flags (mutually exclusive)
	setThemeCmd.Flags().StringVar(&setThemeEcosystem, "ecosystem", "", "Set theme at ecosystem level")
	setThemeCmd.Flags().StringVar(&setThemeDomain, "domain", "", "Set theme at domain level")
	setThemeCmd.Flags().StringVar(&setThemeApp, "app", "", "Set theme at app level")
	setThemeCmd.Flags().StringVar(&setThemeWorkspace, "workspace", "", "Set theme at workspace level")

	// Add kubectl-style flags
	setThemeCmd.Flags().StringVarP(&setThemeOutput, "output", "o", "", "Output format (json, yaml, plain, table, colored)")
	setThemeCmd.Flags().BoolVar(&setThemeDryRun, "dry-run", false, "Preview changes without applying")
	setThemeCmd.Flags().BoolVar(&setThemeShowCascade, "show-cascade", false, "Show theme cascade effect")

	// Ensure exactly one level flag is specified
	setThemeCmd.MarkFlagsOneRequired("ecosystem", "domain", "app", "workspace")
	setThemeCmd.MarkFlagsMutuallyExclusive("ecosystem", "domain", "app", "workspace")
}

func runSetTheme(cmd *cobra.Command, args []string) error {
	themeName := args[0]

	// Validate theme exists (unless clearing with empty string)
	if themeName != "" {
		if err := validateThemeExists(themeName); err != nil {
			return err
		}
	}

	// Build resource context
	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	// Determine which hierarchy level to set and execute
	var result *ThemeSetResult
	if setThemeEcosystem != "" {
		result, err = setEcosystemTheme(cmd, ctx, setThemeEcosystem, themeName)
	} else if setThemeDomain != "" {
		result, err = setDomainTheme(cmd, ctx, setThemeDomain, themeName)
	} else if setThemeApp != "" {
		result, err = setAppTheme(cmd, ctx, setThemeApp, themeName)
	} else if setThemeWorkspace != "" {
		result, err = setWorkspaceTheme(cmd, ctx, setThemeWorkspace, themeName)
	} else {
		return fmt.Errorf("no hierarchy level specified")
	}

	if err != nil {
		return err
	}

	// Handle dry run
	if setThemeDryRun {
		result.ObjectName = result.ObjectName + " (dry-run)"
	}

	// Add cascade information if requested
	if setThemeShowCascade {
		cascadeInfo, err := buildCascadeInfo(cmd, ctx, result)
		if err == nil {
			result.CascadeInfo = cascadeInfo
		}
	}

	// Output result using structured rendering
	opts := render.Options{
		Type:  render.TypeKeyValue,
		Title: fmt.Sprintf("Theme Set: %s", result.Level),
	}

	return render.OutputWith(setThemeOutput, result, opts)
}

// getEffectiveTheme determines what theme will be active after the change
func getEffectiveTheme(newTheme, previousTheme string) string {
	if newTheme == "" {
		if previousTheme == "" {
			return resolver.DefaultTheme
		}
		return previousTheme // Will inherit from parent
	}
	return newTheme
}

// buildCascadeInfo builds cascade information for the result
func buildCascadeInfo(cmd *cobra.Command, ctx resource.Context, result *ThemeSetResult) (*ThemeCascadeInfo, error) {
	// For now, return basic cascade info
	// This could be enhanced to show actual resolution path
	return &ThemeCascadeInfo{
		AffectedLevels: []string{result.Level},
		ResolutionPath: []CascadeStep{
			{
				Level:    result.Level,
				Name:     result.ObjectName,
				Theme:    result.Theme,
				HasTheme: result.Theme != "",
			},
		},
	}, nil
}

// validateThemeExists checks if theme exists in library or store
func validateThemeExists(themeName string) error {
	// Check if theme exists in library
	if library.Has(themeName) {
		return nil
	}

	// TODO: Check custom theme store when available
	// For now, only validate against library
	return fmt.Errorf("theme %q not found. Library themes (34+ available): coolnight-ocean, tokyonight-night, etc. Use 'dvm get nvim themes' to see all available themes", themeName)
}

// setEcosystemTheme sets theme at ecosystem level using resource handlers
func setEcosystemTheme(cmd *cobra.Command, ctx resource.Context, ecosystemName, themeName string) (*ThemeSetResult, error) {
	// Get ecosystem resource using handlers
	res, err := resource.Get(ctx, handlers.KindEcosystem, ecosystemName)
	if err != nil {
		return nil, fmt.Errorf("ecosystem %q not found: %w", ecosystemName, err)
	}

	ecosystemRes := res.(*handlers.EcosystemResource)
	ecosystem := ecosystemRes.Ecosystem()

	// Capture previous theme
	previousTheme := ""
	if ecosystem.Theme.Valid {
		previousTheme = ecosystem.Theme.String
	}

	// Handle dry run
	if setThemeDryRun {
		return &ThemeSetResult{
			Level:          "ecosystem",
			ObjectName:     ecosystemName,
			Theme:          themeName,
			PreviousTheme:  previousTheme,
			EffectiveTheme: getEffectiveTheme(themeName, previousTheme),
		}, nil
	}

	// Set or clear theme
	if themeName == "" {
		ecosystem.Theme.Valid = false
		ecosystem.Theme.String = ""
	} else {
		ecosystem.Theme.Valid = true
		ecosystem.Theme.String = themeName
	}

	// Update using resource handler pattern (need to create YAML and apply)
	ecosystemYAML := ecosystem.ToYAML()
	yamlData, err := yaml.Marshal(ecosystemYAML)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ecosystem YAML: %w", err)
	}

	_, err = resource.Apply(ctx, yamlData, "set-theme")
	if err != nil {
		return nil, fmt.Errorf("failed to update ecosystem: %w", err)
	}

	return &ThemeSetResult{
		Level:          "ecosystem",
		ObjectName:     ecosystemName,
		Theme:          themeName,
		PreviousTheme:  previousTheme,
		EffectiveTheme: getEffectiveTheme(themeName, previousTheme),
	}, nil
}

// setDomainTheme sets theme at domain level using resource handlers
func setDomainTheme(cmd *cobra.Command, ctx resource.Context, domainName, themeName string) (*ThemeSetResult, error) {
	// Get domain resource using handlers
	res, err := resource.Get(ctx, handlers.KindDomain, domainName)
	if err != nil {
		return nil, fmt.Errorf("domain %q not found: %w", domainName, err)
	}

	domainRes := res.(*handlers.DomainResource)
	domain := domainRes.Domain()

	// Capture previous theme
	previousTheme := ""
	if domain.Theme.Valid {
		previousTheme = domain.Theme.String
	}

	// Handle dry run
	if setThemeDryRun {
		return &ThemeSetResult{
			Level:          "domain",
			ObjectName:     domainName,
			Theme:          themeName,
			PreviousTheme:  previousTheme,
			EffectiveTheme: getEffectiveTheme(themeName, previousTheme),
		}, nil
	}

	// Set or clear theme
	if themeName == "" {
		domain.Theme.Valid = false
		domain.Theme.String = ""
	} else {
		domain.Theme.Valid = true
		domain.Theme.String = themeName
	}

	// Update using resource handler pattern
	// Need to get ecosystem name for ToYAML
	ds := ctx.DataStore.(db.DataStore)
	ecosystem, err := ds.GetEcosystemByID(domain.EcosystemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ecosystem for domain: %w", err)
	}

	domainYAML := domain.ToYAML(ecosystem.Name)
	yamlData, err := yaml.Marshal(domainYAML)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal domain YAML: %w", err)
	}

	_, err = resource.Apply(ctx, yamlData, "set-theme")
	if err != nil {
		return nil, fmt.Errorf("failed to update domain: %w", err)
	}

	return &ThemeSetResult{
		Level:          "domain",
		ObjectName:     domainName,
		Theme:          themeName,
		PreviousTheme:  previousTheme,
		EffectiveTheme: getEffectiveTheme(themeName, previousTheme),
	}, nil
}

// setAppTheme sets theme at app level using resource handlers
func setAppTheme(cmd *cobra.Command, ctx resource.Context, appName, themeName string) (*ThemeSetResult, error) {
	// Get app resource using handlers
	res, err := resource.Get(ctx, handlers.KindApp, appName)
	if err != nil {
		return nil, fmt.Errorf("app %q not found: %w", appName, err)
	}

	appRes := res.(*handlers.AppResource)
	app := appRes.App()

	// Capture previous theme
	previousTheme := ""
	if app.Theme.Valid {
		previousTheme = app.Theme.String
	}

	// Handle dry run
	if setThemeDryRun {
		return &ThemeSetResult{
			Level:          "app",
			ObjectName:     appName,
			Theme:          themeName,
			PreviousTheme:  previousTheme,
			EffectiveTheme: getEffectiveTheme(themeName, previousTheme),
		}, nil
	}

	// Set or clear theme
	if themeName == "" {
		app.Theme.Valid = false
		app.Theme.String = ""
	} else {
		app.Theme.Valid = true
		app.Theme.String = themeName
	}

	// Update using resource handler pattern
	// Need to get domain name for ToYAML - extract from context or lookup
	ds := ctx.DataStore.(db.DataStore)
	domain, err := ds.GetDomainByID(app.DomainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain for app: %w", err)
	}

	appYAML := app.ToYAML(domain.Name)
	yamlData, err := yaml.Marshal(appYAML)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal app YAML: %w", err)
	}

	_, err = resource.Apply(ctx, yamlData, "set-theme")
	if err != nil {
		return nil, fmt.Errorf("failed to update app: %w", err)
	}

	return &ThemeSetResult{
		Level:          "app",
		ObjectName:     appName,
		Theme:          themeName,
		PreviousTheme:  previousTheme,
		EffectiveTheme: getEffectiveTheme(themeName, previousTheme),
	}, nil
}

// setWorkspaceTheme sets theme at workspace level using resource handlers
func setWorkspaceTheme(cmd *cobra.Command, ctx resource.Context, workspaceName, themeName string) (*ThemeSetResult, error) {
	// Get workspace resource using handlers
	res, err := resource.Get(ctx, handlers.KindWorkspace, workspaceName)
	if err != nil {
		return nil, fmt.Errorf("workspace %q not found: %w", workspaceName, err)
	}

	workspaceRes := res.(*handlers.WorkspaceResource)
	workspace := workspaceRes.Workspace()
	appName := workspaceRes.AppName()

	// Parse existing NvimStructure to get current theme
	var previousTheme string
	var nvimConfig models.NvimConfig

	if workspace.NvimStructure.Valid && workspace.NvimStructure.String != "" {
		// Try to parse existing structure - if it fails, we'll start fresh
		var wsYAML models.WorkspaceYAML
		if err := yaml.Unmarshal([]byte(workspace.NvimStructure.String), &wsYAML); err == nil {
			nvimConfig = wsYAML.Spec.Nvim
			previousTheme = nvimConfig.Theme
		}
	}

	// Handle dry run
	if setThemeDryRun {
		return &ThemeSetResult{
			Level:          "workspace",
			ObjectName:     workspaceName,
			Theme:          themeName,
			PreviousTheme:  previousTheme,
			EffectiveTheme: getEffectiveTheme(themeName, previousTheme),
		}, nil
	}

	// Update theme in NvimConfig
	nvimConfig.Theme = themeName

	// Create updated workspace YAML with theme
	workspaceYAML := workspace.ToYAML(appName)
	workspaceYAML.Spec.Nvim.Theme = themeName

	// Marshal and store back as NvimStructure
	yamlData, err := yaml.Marshal(workspaceYAML)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal workspace YAML: %w", err)
	}

	workspace.NvimStructure.String = string(yamlData)
	workspace.NvimStructure.Valid = true

	// Apply the updated workspace
	_, err = resource.Apply(ctx, yamlData, "set-theme")
	if err != nil {
		return nil, fmt.Errorf("failed to update workspace: %w", err)
	}

	return &ThemeSetResult{
		Level:          "workspace",
		ObjectName:     workspaceName,
		Theme:          themeName,
		PreviousTheme:  previousTheme,
		EffectiveTheme: getEffectiveTheme(themeName, previousTheme),
	}, nil
}
