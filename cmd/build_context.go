package cmd

import (
	"context"
	"devopsmaestro/builders"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/operators"
	"devopsmaestro/pkg/buildargs/resolver"
	"devopsmaestro/pkg/registry"
	"fmt"
	"io"
	"os"

	"github.com/rmkohlman/MaestroNvim/nvimops/plugin"
	"github.com/rmkohlman/MaestroSDK/render"
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
	cacheReadiness   *registry.CacheReadiness

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

	// output is the writer for all build output. In single-workspace mode
	// this is os.Stdout. In parallel mode, each workspace gets a buffer
	// that is flushed atomically after the build completes.
	output io.Writer
}

// out returns the output writer, defaulting to os.Stdout.
func (bc *buildContext) out() io.Writer {
	if bc.output != nil {
		return bc.output
	}
	return os.Stdout
}

// renderInfo writes an info-level message to the build output.
func (bc *buildContext) renderInfo(msg string) {
	render.MsgTo(bc.out(), "", render.Message{Level: render.LevelInfo, Content: msg})
}

// renderInfof writes a formatted info-level message to the build output.
func (bc *buildContext) renderInfof(format string, args ...any) {
	render.MsgTo(bc.out(), "", render.Message{Level: render.LevelInfo, Content: fmt.Sprintf(format, args...)})
}

// renderProgress writes a progress-level message to the build output.
func (bc *buildContext) renderProgress(msg string) {
	render.MsgTo(bc.out(), "", render.Message{Level: render.LevelProgress, Content: msg})
}

// renderProgressf writes a formatted progress-level message to the build output.
func (bc *buildContext) renderProgressf(format string, args ...any) {
	render.MsgTo(bc.out(), "", render.Message{Level: render.LevelProgress, Content: fmt.Sprintf(format, args...)})
}

// renderWarning writes a warning-level message to the build output.
func (bc *buildContext) renderWarning(msg string) {
	render.MsgTo(bc.out(), "", render.Message{Level: render.LevelWarning, Content: msg})
}

// renderWarningf writes a formatted warning-level message to the build output.
func (bc *buildContext) renderWarningf(format string, args ...any) {
	render.MsgTo(bc.out(), "", render.Message{Level: render.LevelWarning, Content: fmt.Sprintf(format, args...)})
}

// renderSuccess writes a success-level message to the build output.
func (bc *buildContext) renderSuccess(msg string) {
	render.MsgTo(bc.out(), "", render.Message{Level: render.LevelSuccess, Content: msg})
}

// renderSuccessf writes a formatted success-level message to the build output.
func (bc *buildContext) renderSuccessf(format string, args ...any) {
	render.MsgTo(bc.out(), "", render.Message{Level: render.LevelSuccess, Content: fmt.Sprintf(format, args...)})
}

// renderPlain writes a plain text message to the build output.
func (bc *buildContext) renderPlain(msg string) {
	fmt.Fprintln(bc.out(), msg)
}

// renderBlank writes a blank line to the build output.
func (bc *buildContext) renderBlank() {
	fmt.Fprintln(bc.out())
}
