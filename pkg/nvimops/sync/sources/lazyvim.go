// Package sources provides concrete implementations of SourceHandler for different
// Neovim configuration frameworks like LazyVim, AstroNvim, etc.
package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"devopsmaestro/pkg/nvimops/plugin"
	"devopsmaestro/pkg/nvimops/sync"
	"gopkg.in/yaml.v3"
)

// LazyVimHandler implements SourceHandler for the LazyVim configuration framework.
// It fetches plugin definitions from the LazyVim GitHub repository and converts
// them to our standard NvimPlugin YAML format.
type LazyVimHandler struct {
	// client is the HTTP client used for GitHub API requests
	client *http.Client

	// baseURL is the base GitHub API URL for LazyVim
	baseURL string

	// lastSyncSHA tracks the last commit SHA that was synced for caching
	lastSyncSHA string

	// version tracks the LazyVim version for labeling
	version string
}

// GitHubContent represents a file or directory from GitHub API
type GitHubContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	Size        int    `json:"size"`
	SHA         string `json:"sha"`
	URL         string `json:"url"`
	GitURL      string `json:"git_url"`
	HTMLURL     string `json:"html_url"`
	DownloadURL string `json:"download_url"`
}

// GitHubRelease represents a release from GitHub API
type GitHubRelease struct {
	TagName string    `json:"tag_name"`
	Name    string    `json:"name"`
	Draft   bool      `json:"draft"`
	Created time.Time `json:"created_at"`
}

// NewLazyVimHandler creates a new LazyVim source handler.
func NewLazyVimHandler() sync.SourceHandler {
	return &LazyVimHandler{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.github.com/repos/LazyVim/LazyVim",
	}
}

// Name returns the unique identifier for this source.
func (h *LazyVimHandler) Name() string {
	return "lazyvim"
}

// Description returns a human-readable description of the source.
func (h *LazyVimHandler) Description() string {
	return "LazyVim - A Neovim config for lazy people"
}

// Validate checks if the LazyVim source is accessible.
func (h *LazyVimHandler) Validate(ctx context.Context) error {
	// Try to fetch the repository info to validate access
	url := h.baseURL
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to access LazyVim repository: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("LazyVim repository returned status %d", resp.StatusCode)
	}

	return nil
}

// ListAvailable returns all plugins available from LazyVim.
func (h *LazyVimHandler) ListAvailable(ctx context.Context) ([]sync.AvailablePlugin, error) {
	// First, get the latest version for labeling
	if err := h.fetchLatestVersion(ctx); err != nil {
		// Don't fail if we can't get version, just continue without it
		h.version = "unknown"
	}

	// Get the list of plugin files from LazyVim
	pluginFiles, err := h.fetchPluginFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch plugin files: %w", err)
	}

	var availablePlugins []sync.AvailablePlugin

	// Process each plugin file
	for _, file := range pluginFiles {
		if !strings.HasSuffix(file.Name, ".lua") {
			continue
		}

		// Fetch and parse the file content
		plugins, err := h.parsePluginFile(ctx, file)
		if err != nil {
			// Log the error but continue with other files
			continue
		}

		availablePlugins = append(availablePlugins, plugins...)
	}

	return availablePlugins, nil
}

// Sync imports plugins from LazyVim based on the provided options.
func (h *LazyVimHandler) Sync(ctx context.Context, options sync.SyncOptions) (*sync.SyncResult, error) {
	result := &sync.SyncResult{
		SourceName: h.Name(),
	}

	// Get all available plugins
	availablePlugins, err := h.ListAvailable(ctx)
	if err != nil {
		result.AddError(fmt.Errorf("failed to list available plugins: %w", err))
		return result, nil
	}

	result.TotalAvailable = len(availablePlugins)

	// Filter plugins based on sync options
	var filteredPlugins []sync.AvailablePlugin
	for _, plugin := range availablePlugins {
		if options.MatchesAvailablePlugin(plugin) {
			filteredPlugins = append(filteredPlugins, plugin)
		}
	}

	// Convert each plugin to YAML and write to files (if not dry run)
	var syncedPluginNames []string
	for _, availablePlugin := range filteredPlugins {
		pluginYAML := h.convertToPluginYAML(availablePlugin)

		if !options.DryRun {
			// Write the YAML file to options.TargetDir
			if options.TargetDir != "" {
				filename := filepath.Join(options.TargetDir, availablePlugin.Name+".yaml")

				// Create directory if it doesn't exist
				if err := os.MkdirAll(options.TargetDir, 0755); err != nil {
					result.AddError(fmt.Errorf("failed to create target directory: %w", err))
					continue
				}

				// Convert to YAML bytes
				yamlData, err := yaml.Marshal(pluginYAML)
				if err != nil {
					result.AddError(fmt.Errorf("failed to serialize plugin %s: %w", availablePlugin.Name, err))
					continue
				}

				// Write file
				if err := os.WriteFile(filename, yamlData, 0644); err != nil {
					result.AddError(fmt.Errorf("failed to write plugin %s: %w", availablePlugin.Name, err))
					continue
				}
			}

			result.AddPluginCreated(availablePlugin.Name)
			syncedPluginNames = append(syncedPluginNames, availablePlugin.Name)
		} else {
			result.AddPluginCreated(availablePlugin.Name)
			syncedPluginNames = append(syncedPluginNames, availablePlugin.Name)
		}
	}

	// Create package if PackageCreator is available and we have synced plugins
	if options.PackageCreator != nil && len(syncedPluginNames) > 0 {
		if !options.DryRun {
			if err := options.PackageCreator.CreatePackage(h.Name(), syncedPluginNames); err != nil {
				result.AddError(fmt.Errorf("failed to create package: %w", err))
			} else {
				result.AddPackageCreated(h.Name())
			}
		} else {
			result.AddPackageCreated(h.Name())
		}
	}

	return result, nil
}

// fetchLatestVersion gets the latest release version from GitHub API.
func (h *LazyVimHandler) fetchLatestVersion(ctx context.Context) error {
	url := h.baseURL + "/releases/latest"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// No releases, try to get default branch info
		return h.fetchDefaultBranchSHA(ctx)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return err
	}

	h.version = release.TagName
	return nil
}

// fetchDefaultBranchSHA gets the SHA of the default branch for versioning.
func (h *LazyVimHandler) fetchDefaultBranchSHA(ctx context.Context) error {
	url := h.baseURL + "/branches/main"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get default branch info: status %d", resp.StatusCode)
	}

	var branch struct {
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&branch); err != nil {
		return err
	}

	h.version = branch.Commit.SHA[:7] // Short SHA
	h.lastSyncSHA = branch.Commit.SHA
	return nil
}

// fetchPluginFiles gets the list of plugin files from the LazyVim repository.
func (h *LazyVimHandler) fetchPluginFiles(ctx context.Context) ([]GitHubContent, error) {
	url := h.baseURL + "/contents/lua/lazyvim/plugins"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch plugin files: status %d", resp.StatusCode)
	}

	var contents []GitHubContent
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return nil, err
	}

	return contents, nil
}

// parsePluginFile fetches and parses a single Lua plugin file.
func (h *LazyVimHandler) parsePluginFile(ctx context.Context, file GitHubContent) ([]sync.AvailablePlugin, error) {
	if file.DownloadURL == "" {
		return nil, fmt.Errorf("no download URL for file %s", file.Name)
	}

	// Fetch the raw file content
	req, err := http.NewRequestWithContext(ctx, "GET", file.DownloadURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch file content: status %d", resp.StatusCode)
	}

	// Read the content
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse the Lua content to extract plugins
	return h.parseLuaContent(string(content), file.Name)
}

// parseLuaContent extracts plugin specifications from Lua code.
// This is a pragmatic regex-based approach for common LazyVim patterns.
func (h *LazyVimHandler) parseLuaContent(content, filename string) ([]sync.AvailablePlugin, error) {
	var plugins []sync.AvailablePlugin

	// Extract the category from filename
	category := h.extractCategoryFromFilename(filename)

	// Pattern to match plugin specifications like: { "repo/name", ... }
	// This is a simplified regex - in production, you'd want more sophisticated parsing
	pluginRegex := regexp.MustCompile(`\{\s*["']([^/]+/[^"']+)["'][^}]*\}`)
	matches := pluginRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		repo := match[1]
		pluginName := h.extractPluginNameFromRepo(repo)

		// Create the available plugin
		availablePlugin := sync.AvailablePlugin{
			Name:        fmt.Sprintf("lazyvim-%s", pluginName),
			Description: fmt.Sprintf("LazyVim plugin: %s", repo),
			Category:    category,
			Repo:        repo,
			SourceName:  h.Name(),
			Labels: map[string]string{
				"source":       "lazyvim",
				"category":     category,
				"lazyvim-file": filename,
			},
		}

		if h.version != "" {
			availablePlugin.Labels["lazyvim-version"] = h.version
		}

		// Try to extract additional configuration from the full match
		fullMatch := match[0]
		availablePlugin.Config = h.extractConfigFromMatch(fullMatch)
		availablePlugin.Dependencies = h.extractDependenciesFromMatch(fullMatch)

		plugins = append(plugins, availablePlugin)
	}

	return plugins, nil
}

// extractCategoryFromFilename determines plugin category from the LazyVim filename.
func (h *LazyVimHandler) extractCategoryFromFilename(filename string) string {
	name := strings.TrimSuffix(filename, ".lua")
	switch name {
	case "coding":
		return "coding"
	case "colorscheme":
		return "theme"
	case "editor":
		return "editor"
	case "formatting":
		return "formatting"
	case "linting":
		return "linting"
	case "treesitter":
		return "syntax"
	case "ui":
		return "ui"
	case "util":
		return "utility"
	default:
		// For files in subdirectories like lsp/init.lua
		if strings.Contains(name, "lsp") {
			return "lsp"
		}
		return "misc"
	}
}

// extractPluginNameFromRepo extracts a plugin name from a GitHub repository.
func (h *LazyVimHandler) extractPluginNameFromRepo(repo string) string {
	parts := strings.Split(repo, "/")
	if len(parts) >= 2 {
		name := parts[1]
		// Remove common suffixes
		name = strings.TrimSuffix(name, ".nvim")
		name = strings.TrimSuffix(name, "-nvim")
		name = strings.TrimSuffix(name, ".vim")
		// Remove nvim- prefix for cleaner names
		name = strings.TrimPrefix(name, "nvim-")
		return name
	}
	return repo
}

// extractConfigFromMatch tries to extract configuration from a plugin match.
func (h *LazyVimHandler) extractConfigFromMatch(match string) string {
	// This is a simplified extraction - in practice, you'd want more sophisticated parsing
	configRegex := regexp.MustCompile(`config\s*=\s*function\(\)[^}]*end`)
	if configMatch := configRegex.FindString(match); configMatch != "" {
		return configMatch
	}

	// Look for opts table
	optsRegex := regexp.MustCompile(`opts\s*=\s*\{[^}]*\}`)
	if optsMatch := optsRegex.FindString(match); optsMatch != "" {
		return optsMatch
	}

	return ""
}

// extractDependenciesFromMatch tries to extract dependencies from a plugin match.
func (h *LazyVimHandler) extractDependenciesFromMatch(match string) []string {
	var dependencies []string

	// Look for dependencies array
	depRegex := regexp.MustCompile(`dependencies\s*=\s*\{([^}]*)\}`)
	depMatch := depRegex.FindStringSubmatch(match)
	if len(depMatch) >= 2 {
		// Extract individual dependencies
		depContent := depMatch[1]
		depItemRegex := regexp.MustCompile(`["']([^/]+/[^"']+)["']`)
		depMatches := depItemRegex.FindAllStringSubmatch(depContent, -1)

		for _, depItem := range depMatches {
			if len(depItem) >= 2 {
				dependencies = append(dependencies, depItem[1])
			}
		}
	}

	return dependencies
}

// convertToPluginYAML converts an AvailablePlugin to our standard Plugin YAML format.
func (h *LazyVimHandler) convertToPluginYAML(available sync.AvailablePlugin) *plugin.PluginYAML {
	pluginYAML := plugin.NewPluginYAML(available.Name, available.Repo)

	// Set metadata
	pluginYAML.Metadata.Description = available.Description
	pluginYAML.Metadata.Category = available.Category
	pluginYAML.Metadata.Labels = make(map[string]string)

	// Copy labels
	for k, v := range available.Labels {
		pluginYAML.Metadata.Labels[k] = v
	}

	// Set configuration if available
	if available.Config != "" {
		pluginYAML.Spec.Config = available.Config
	}

	// Convert dependencies
	for _, dep := range available.Dependencies {
		pluginYAML.Spec.Dependencies = append(pluginYAML.Spec.Dependencies, plugin.DependencyYAML{
			Repo: dep,
		})
	}

	// LazyVim plugins are typically lazy-loaded
	pluginYAML.Spec.Lazy = true

	return pluginYAML
}

// RegisterLazyVimHandler registers the LazyVim handler in the provided registry,
// replacing any existing placeholder registration.
func RegisterLazyVimHandler(registry *sync.SourceRegistry) error {
	// Get the LazyVim source info from the registry (it should exist from builtin sources)
	info, err := registry.GetSourceInfo("lazyvim")
	if err != nil {
		// If not found, create default info
		info = &sync.SourceInfo{
			Name:        "lazyvim",
			Description: "LazyVim - A Neovim config for lazy people",
			URL:         "https://github.com/LazyVim/LazyVim",
			Type:        string(sync.SourceTypeGitHub),
			ConfigKeys:  []string{"repo_url", "branch", "plugins_dir"},
		}
	}

	// Create the registration with the actual handler
	registration := sync.HandlerRegistration{
		Name: "lazyvim",
		Info: *info,
		CreateFunc: func() sync.SourceHandler {
			return NewLazyVimHandler()
		},
	}

	// Unregister the placeholder if it exists
	if registry.IsRegistered("lazyvim") {
		_ = registry.Unregister("lazyvim")
	}

	// Register the actual handler
	return registry.Register(registration)
}
