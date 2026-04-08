package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rmkohlman/MaestroNvim/nvimops/plugin"
	"github.com/rmkohlman/MaestroSDK/render"

	"github.com/spf13/cobra"
)

// =============================================================================
// HEALTH COMMAND
// =============================================================================

var healthCmd = &cobra.Command{
	Use:   "health [plugin-name]",
	Short: "Check health of installed plugins",
	Long: `Run health checks for installed Neovim plugins.

Without arguments, checks all enabled plugins. With a name, checks only that plugin.

Health checks verify that plugins are installed and working correctly by checking:
  - Lua modules are loadable (require() succeeds)
  - Neovim commands exist
  - Treesitter parsers are installed
  - LSP servers are configured

By default, runs static checks only (no Neovim required).
Use --live to run checks inside Neovim (requires nvim on PATH).

Examples:
  nvp health                    # Static check all plugins
  nvp health telescope          # Check specific plugin
  nvp health --live             # Run checks inside Neovim
  nvp health -o json            # Output as JSON`,
	Args: cobra.MaximumNArgs(1),
	RunE: runHealth,
}

func init() {
	healthCmd.Flags().StringP("output", "o", "table", "Output format: table, json, yaml")
	healthCmd.Flags().Bool("live", false, "Run checks inside Neovim (requires nvim)")
	healthCmd.Flags().Bool("generate-script", false, "Output the Lua health check script (for debugging)")
}

func runHealth(cmd *cobra.Command, args []string) error {
	mgr, err := getManager()
	if err != nil {
		return err
	}
	defer mgr.Close()

	plugins, err := mgr.List()
	if err != nil {
		return fmt.Errorf("failed to list plugins: %w", err)
	}

	// Filter to specific plugin if name provided
	if len(args) > 0 {
		name := args[0]
		var filtered []*plugin.Plugin
		for _, p := range plugins {
			if p.Name == name {
				filtered = append(filtered, p)
			}
		}
		if len(filtered) == 0 {
			return fmt.Errorf("plugin not found: %s", name)
		}
		plugins = filtered
	}

	if len(plugins) == 0 {
		render.Info("No plugins installed")
		return nil
	}

	// Check for generate-script flag (debugging aid)
	genScript, _ := cmd.Flags().GetBool("generate-script")
	if genScript {
		luaScript := plugin.GenerateHealthCheckLua(plugins)
		fmt.Print(luaScript)
		return nil
	}

	live, _ := cmd.Flags().GetBool("live")
	format, _ := cmd.Flags().GetString("output")

	var reports []*plugin.PluginHealthReport

	if live {
		reports, err = runLiveHealthCheck(plugins)
		if err != nil {
			render.Errorf("Live health check failed: %v", err)
			render.Info("Falling back to static checks...")
			reports = runStaticHealthCheck(plugins)
		}
	} else {
		reports = runStaticHealthCheck(plugins)
	}

	return outputHealthReports(reports, format)
}

// runStaticHealthCheck performs health checks without Neovim.
func runStaticHealthCheck(plugins []*plugin.Plugin) []*plugin.PluginHealthReport {
	checker := plugin.NewHealthChecker()
	return checker.StaticCheck(plugins)
}

// runLiveHealthCheck generates a Lua script, runs it in Neovim, and parses results.
func runLiveHealthCheck(plugins []*plugin.Plugin) ([]*plugin.PluginHealthReport, error) {
	// Check if nvim is available
	nvimPath, err := exec.LookPath("nvim")
	if err != nil {
		return nil, fmt.Errorf("nvim not found on PATH: %w", err)
	}

	// Generate the health check Lua script
	luaScript := plugin.GenerateHealthCheckLua(plugins)

	// Write to temp file
	tmpDir, err := os.MkdirTemp("", "nvp-health-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	scriptPath := filepath.Join(tmpDir, "health_check.lua")
	if err := os.WriteFile(scriptPath, []byte(luaScript), 0644); err != nil {
		return nil, fmt.Errorf("failed to write script: %w", err)
	}

	// Run nvim headless with the script
	nvimCmd := exec.Command(nvimPath, "--headless", "-l", scriptPath)
	output, err := nvimCmd.Output()
	if err != nil {
		// Try to get stderr for debugging
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("nvim failed: %s\nstderr: %s",
				err, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("nvim failed: %w", err)
	}

	// Parse JSON output
	jsonOutput := strings.TrimSpace(string(output))
	if jsonOutput == "" {
		return nil, fmt.Errorf("nvim produced no output")
	}

	checker := plugin.NewHealthChecker()
	return checker.ParseNvimResults([]byte(jsonOutput))
}

// outputHealthReports formats and prints health reports.
func outputHealthReports(reports []*plugin.PluginHealthReport, format string) error {
	switch format {
	case "json":
		data, err := json.MarshalIndent(reports, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))

	case "table", "":
		return outputHealthTable(reports)

	default:
		return fmt.Errorf("unknown format: %s (supported: table, json)", format)
	}
	return nil
}

// outputHealthTable prints a table of health check results.
func outputHealthTable(reports []*plugin.PluginHealthReport) error {
	// Summary counters
	var healthy, unhealthy, unknown, skipped int

	fmt.Println()
	fmt.Printf("  %-25s %-12s %s\n", "PLUGIN", "STATUS", "DETAILS")
	fmt.Printf("  %-25s %-12s %s\n",
		strings.Repeat("─", 25),
		strings.Repeat("─", 12),
		strings.Repeat("─", 40))

	for _, r := range reports {
		statusIcon := statusToIcon(r.Status)
		details := summarizeResults(r)

		fmt.Printf("  %-25s %s %-10s %s\n",
			truncate(r.PluginName, 25),
			statusIcon,
			string(r.Status),
			details)

		switch r.Status {
		case plugin.HealthStatusHealthy:
			healthy++
		case plugin.HealthStatusUnhealthy:
			unhealthy++
		case plugin.HealthStatusSkipped:
			skipped++
		default:
			unknown++
		}
	}

	// Summary line
	fmt.Println()
	parts := []string{}
	if healthy > 0 {
		parts = append(parts, fmt.Sprintf("%d healthy", healthy))
	}
	if unhealthy > 0 {
		parts = append(parts, fmt.Sprintf("%d unhealthy", unhealthy))
	}
	if unknown > 0 {
		parts = append(parts, fmt.Sprintf("%d unknown", unknown))
	}
	if skipped > 0 {
		parts = append(parts, fmt.Sprintf("%d skipped", skipped))
	}
	fmt.Printf("  %d plugins checked: %s\n\n", len(reports), strings.Join(parts, ", "))

	return nil
}

// statusToIcon returns an icon for a health status.
func statusToIcon(status plugin.HealthStatus) string {
	switch status {
	case plugin.HealthStatusHealthy:
		return "✓"
	case plugin.HealthStatusUnhealthy:
		return "✗"
	case plugin.HealthStatusSkipped:
		return "⊘"
	default:
		return "?"
	}
}

// summarizeResults creates a brief summary of check results.
func summarizeResults(r *plugin.PluginHealthReport) string {
	if !r.Enabled {
		return "(disabled)"
	}
	if len(r.Results) == 0 {
		return "(no checks defined)"
	}

	var parts []string
	for _, result := range r.Results {
		parts = append(parts, fmt.Sprintf("%s:%s=%s",
			result.Check.Type, result.Check.Value, result.Status))
	}
	summary := strings.Join(parts, ", ")
	return truncate(summary, 50)
}

// truncate shortens a string to max length, adding "..." if truncated.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
