package cmd

import (
	"context"
	"devopsmaestro/builders"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/operators"
	"devopsmaestro/pkg/buildargs/resolver"
	"github.com/rmkohlman/MaestroNvim/nvimops/plugin"
)

// buildContext carries shared state across the build pipeline phases.
// Each phase method reads/writes fields as needed, keeping the orchestrator
// function clean and each phase independently testable.
type buildContext struct {
	// Injected dependencies
	ds  db.DataStore
	ctx context.Context

	// Resolved workspace target
	app           *models.App
	workspace     *models.Workspace
	appName       string
	workspaceName string

	// Platform & registry
	platform         *operators.Platform
	registryEndpoint string
	registryEnvVars  map[string]string

	// Dockerfile detection
	hasDockerfile  bool
	dockerfilePath string

	// Workspace spec
	workspaceYAML models.WorkspaceYAML

	// Paths
	homeDir    string
	sourcePath string
	stagingDir string

	// Language detection
	languageName string
	version      string

	// Nvim
	pluginManifest *plugin.PluginManifest

	// Build args cascade (resolved once, used twice: Dockerfile gen + build args)
	cascadeResolution *resolver.BuildArgsResolution

	// Build artifacts
	imageName     string
	dvmDockerfile string

	// Image builder (set during buildImage, closed by caller)
	builder builders.ImageBuilder
}
