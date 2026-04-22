package cmd

import (
	"fmt"
	"os"

	"devopsmaestro/builders"
	"devopsmaestro/operators"
	themeresolver "devopsmaestro/pkg/colors/resolver"
	"devopsmaestro/pkg/nvimbridge"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroTerminal/terminalops/shell"

	"github.com/spf13/cobra"
)

// ContextOutput represents context for output formatting
type ContextOutput struct {
	CurrentEcosystem string `yaml:"currentEcosystem" json:"currentEcosystem"`
	CurrentDomain    string `yaml:"currentDomain" json:"currentDomain"`
	CurrentSystem    string `yaml:"currentSystem" json:"currentSystem"`
	CurrentApp       string `yaml:"currentApp" json:"currentApp"`
	CurrentWorkspace string `yaml:"currentWorkspace" json:"currentWorkspace"`
}

func getContext(cmd *cobra.Command) error {
	// Read from database context (authoritative source for all 4 hierarchy levels)
	ds, err := getDataStore(cmd)
	if err != nil {
		return fmt.Errorf("failed to get data store: %w", err)
	}

	dbCtx, err := ds.GetContext()
	if err != nil {
		return fmt.Errorf("failed to load context: %w", err)
	}

	// Resolve IDs to names, tracking source
	var ecosystemName, domainName, systemName, appName, workspaceName string
	var ecosystemSource, domainSource, systemSource, appSource, workspaceSource string

	if dbCtx != nil {
		if dbCtx.ActiveEcosystemID != nil {
			if eco, err := ds.GetEcosystemByID(*dbCtx.ActiveEcosystemID); err == nil {
				ecosystemName = eco.Name
				ecosystemSource = "global"
			}
		}
		if dbCtx.ActiveDomainID != nil {
			if dom, err := ds.GetDomainByID(*dbCtx.ActiveDomainID); err == nil {
				domainName = dom.Name
				domainSource = "global"
			}
		}
		if dbCtx.ActiveSystemID != nil {
			if sys, err := ds.GetSystemByID(*dbCtx.ActiveSystemID); err == nil {
				systemName = sys.Name
				systemSource = "global"
			}
		}
		if dbCtx.ActiveAppID != nil {
			if app, err := ds.GetAppByID(*dbCtx.ActiveAppID); err == nil {
				appName = app.Name
				appSource = "global"
			}
		}
		if dbCtx.ActiveWorkspaceID != nil {
			if ws, err := ds.GetWorkspaceByID(*dbCtx.ActiveWorkspaceID); err == nil {
				workspaceName = ws.Name
				workspaceSource = "global"
			}
		}
	}

	// Check env var overrides (DVM_ECOSYSTEM, DVM_DOMAIN, DVM_SYSTEM, DVM_APP, DVM_WORKSPACE)
	if envEco := os.Getenv("DVM_ECOSYSTEM"); envEco != "" {
		ecosystemName = envEco
		ecosystemSource = "env: DVM_ECOSYSTEM"
	}
	if envDom := os.Getenv("DVM_DOMAIN"); envDom != "" {
		domainName = envDom
		domainSource = "env: DVM_DOMAIN"
	}
	if envSys := os.Getenv("DVM_SYSTEM"); envSys != "" {
		systemName = envSys
		systemSource = "env: DVM_SYSTEM"
	}
	if envApp := os.Getenv("DVM_APP"); envApp != "" {
		appName = envApp
		appSource = "env: DVM_APP"
	}
	if envWorkspace := os.Getenv("DVM_WORKSPACE"); envWorkspace != "" {
		workspaceName = envWorkspace
		workspaceSource = "env: DVM_WORKSPACE"
	}

	// Build structured data
	data := ContextOutput{
		CurrentEcosystem: ecosystemName,
		CurrentDomain:    domainName,
		CurrentSystem:    systemName,
		CurrentApp:       appName,
		CurrentWorkspace: workspaceName,
	}

	// Check if empty
	isEmpty := ecosystemName == "" && domainName == "" && systemName == "" && appName == "" && workspaceName == ""

	// For structured output (JSON/YAML), always output the data structure
	// For human output, show nice key-value display
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		return render.OutputWith(getOutputFormat, data, render.Options{})
	}

	// Human-readable output
	if isEmpty {
		return render.Output(nil, render.Options{
			Empty:        true,
			EmptyMessage: "No active context",
			EmptyHints: []string{
				"dvm use ecosystem <name>",
				"dvm use domain <name>",
				"dvm use system <name>",
				"dvm use app <name>",
				"dvm use workspace <name>",
			},
		})
	}

	// Build key-value pairs, showing "(none)" for unset levels
	displayWithSource := func(name, source string) string {
		if name == "" {
			return "(none)"
		}
		if source != "" {
			return fmt.Sprintf("%s (%s)", name, source)
		}
		return name
	}

	kvData := render.NewOrderedKeyValueData(
		render.KeyValue{Key: "Ecosystem", Value: displayWithSource(ecosystemName, ecosystemSource)},
		render.KeyValue{Key: "Domain", Value: displayWithSource(domainName, domainSource)},
		render.KeyValue{Key: "System", Value: displayWithSource(systemName, systemSource)},
		render.KeyValue{Key: "App", Value: displayWithSource(appName, appSource)},
		render.KeyValue{Key: "Workspace", Value: displayWithSource(workspaceName, workspaceSource)},
	)

	return render.OutputWith(getOutputFormat, kvData, render.Options{
		Type:  render.TypeKeyValue,
		Title: "Current Context",
	})
}

// PlatformOutput represents a platform for output
type PlatformOutput struct {
	Type         string `yaml:"type" json:"type"`
	Name         string `yaml:"name" json:"name"`
	SocketPath   string `yaml:"socketPath" json:"socketPath"`
	Profile      string `yaml:"profile,omitempty" json:"profile,omitempty"`
	IsContainerd bool   `yaml:"isContainerd" json:"isContainerd"`
	IsDocker     bool   `yaml:"isDockerCompatible" json:"isDockerCompatible"`
	Active       bool   `yaml:"active" json:"active"`
}

func getPlatforms(cmd *cobra.Command) error {
	detector, err := operators.NewPlatformDetector()
	if err != nil {
		return fmt.Errorf("failed to create platform detector: %w", err)
	}

	platforms := detector.DetectAll()

	if len(platforms) == 0 {
		return render.OutputWith(getOutputFormat, nil, render.Options{
			Empty:        true,
			EmptyMessage: "No container platforms detected",
			EmptyHints:   []string{"Install OrbStack, Colima, Docker Desktop, or Podman"},
		})
	}

	// Get active platform
	activePlatform, _ := detector.Detect()
	var activeName string
	if activePlatform != nil {
		activeName = string(activePlatform.Type)
	}

	// Build platform output data
	platformsOutput := make([]PlatformOutput, len(platforms))
	for i, p := range platforms {
		platformsOutput[i] = PlatformOutput{
			Type:         string(p.Type),
			Name:         p.Name,
			SocketPath:   p.SocketPath,
			Profile:      p.Profile,
			IsContainerd: p.IsContainerd(),
			IsDocker:     p.IsDockerCompatible(),
			Active:       string(p.Type) == activeName,
		}
	}

	// For JSON/YAML, output directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		return render.OutputWith(getOutputFormat, platformsOutput, render.Options{})
	}

	// For human output, build table data
	tableData := render.TableData{
		Headers: []string{"TYPE", "NAME", "SOCKET", "STATUS"},
		Rows:    make([][]string, len(platforms)),
	}

	for i, p := range platforms {
		status := ""
		if platformsOutput[i].Active {
			status = "● active"
		}

		socketDisplay := p.SocketPath
		if len(socketDisplay) > 45 {
			socketDisplay = "..." + socketDisplay[len(socketDisplay)-42:]
		}

		name := p.Name
		if p.IsContainerd() {
			name += " (containerd)"
		} else if p.IsDockerCompatible() {
			name += " (docker)"
		}

		tableData.Rows[i] = []string{
			string(p.Type),
			name,
			socketDisplay,
			status,
		}
	}

	return render.OutputWith(getOutputFormat, tableData, render.Options{
		Type: render.TypeTable,
	})
}

// DefaultsOutput represents default configuration values for output
type DefaultsOutput struct {
	Theme     map[string]interface{} `yaml:"theme" json:"theme"`
	Shell     map[string]interface{} `yaml:"shell" json:"shell"`
	Nvim      map[string]interface{} `yaml:"nvim" json:"nvim"`
	Container map[string]interface{} `yaml:"container" json:"container"`
}

func getDefaults(cmd *cobra.Command) error {
	// Get defaults from all packages (hardcoded defaults)
	themeDefaults := themeresolver.GetDefaults()
	shellDefaults := shell.GetDefaults()
	nvimDefaults := nvimbridge.GetDefaults()
	containerDefaults := builders.GetContainerDefaults()

	// Override with user-set defaults from database
	ds, err := getDataStore(cmd)
	if err == nil {
		// Check for user-set nvim package
		if userPkg, err := ds.GetDefault("nvim-package"); err == nil && userPkg != "" {
			nvimDefaults["pluginPackage"] = userPkg
		}
		// Check for user-set terminal package
		if userTermPkg, err := ds.GetDefault("terminal-package"); err == nil && userTermPkg != "" {
			shellDefaults["terminalPackage"] = userTermPkg
		}
		// Check for user-set global theme
		if userTheme, err := ds.GetDefault("theme"); err == nil && userTheme != "" {
			themeDefaults["global"] = userTheme
		}
	}

	// Build structured data
	data := DefaultsOutput{
		Theme:     themeDefaults,
		Shell:     shellDefaults,
		Nvim:      nvimDefaults,
		Container: containerDefaults,
	}

	// For JSON/YAML, output the data structure directly
	if getOutputFormat == "json" || getOutputFormat == "yaml" {
		return render.OutputWith(getOutputFormat, data, render.Options{})
	}

	// For human-readable output, show organized key-value display
	render.Blank()
	render.Info("Theme Defaults:")
	for key, value := range themeDefaults {
		render.Plainf("  %s: %v", key, value)
	}

	render.Blank()
	render.Info("Shell Defaults:")
	for key, value := range shellDefaults {
		render.Plainf("  %s: %v", key, value)
	}

	render.Blank()
	render.Info("Neovim Defaults:")
	for key, value := range nvimDefaults {
		render.Plainf("  %s: %v", key, value)
	}

	render.Blank()
	render.Info("Container Defaults:")
	for key, value := range containerDefaults {
		render.Plainf("  %s: %v", key, value)
	}

	return nil
}
