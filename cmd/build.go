package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

var (
	buildForce       bool
	buildNocache     bool
	buildTarget      string
	buildPush        bool
	buildRegistry    string
	buildTimeout     time.Duration
	buildFlags       HierarchyFlags
	buildDryRun      bool
	buildAll         bool
	buildDetach      bool
	buildConcurrency int
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build workspace container image",
	Long: `Build a development container image for the active workspace.

This command:
- Detects the app language
- Generates or extends Dockerfile with dev tools
- Emits ARG declarations for all spec.build.args keys (not ENV — credentials are not persisted in image layers)
- Injects CA certificates from MaestroVault when spec.build.caCerts is configured (fatal error if cert is missing or invalid)
- Sets the USER directive from container.user (defaults to "dev" if unset)
- Builds the image using the detected container platform
- Tags as dvm-<workspace>-<app>:latest
- Optionally pushes to local registry cache

Supports multiple platforms:
- OrbStack (uses Docker API)
- Docker Desktop (uses Docker API)
- Podman (uses Docker API)
- Colima with containerd (uses BuildKit API)

Use DVM_PLATFORM environment variable to select a specific platform.

Registry Integration:
  If registry.enabled is true in config and lifecycle is "on-demand" or "persistent",
  the build command will automatically start the registry before building. This provides
  image caching to speed up builds and reduce network usage.

Flags:
  -e, --ecosystem   Filter by ecosystem name
  -d, --domain      Filter by domain name  
  -a, --app         Filter by app name
  -w, --workspace   Filter by workspace name
  --no-cache        Build without using registry cache (pull fresh from upstream)
  --push            Push built image to local registry after build
  --registry        Override registry endpoint (default: from config)

Examples:
  dvm build
  dvm build --force
  dvm build --no-cache
  dvm build --push                        # Push to local registry
  dvm build -a portal                     # Build workspace in 'portal' app
  dvm build -e healthcare -a portal       # Specify ecosystem and app
  DVM_PLATFORM=colima dvm build
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate --all mutual exclusion with hierarchy flags
		allSet, _ := cmd.Flags().GetBool("all")
		if err := validateBuildAllMutualExclusion(allSet, buildFlags); err != nil {
			return err
		}

		// Route to parallel build path when --all or scope flags are set
		if shouldRouteToParallelBuild(buildFlags, allSet) {
			return runParallelBuild(cmd)
		}

		return buildWorkspace(cmd)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().BoolVar(&buildForce, "force", false, "Force rebuild even if image exists")
	buildCmd.Flags().BoolVar(&buildNocache, "no-cache", false, "Build without using cache (skip registry cache)")
	buildCmd.Flags().StringVar(&buildTarget, "target", "dev", "Build target stage (default: dev)")
	buildCmd.Flags().BoolVar(&buildPush, "push", false, "Push built image to local registry")
	buildCmd.Flags().StringVar(&buildRegistry, "registry", "", "Override registry endpoint (default: from config)")
	buildCmd.Flags().DurationVar(&buildTimeout, "timeout", 30*time.Minute, "Timeout for the build operation (e.g., 30m, 1h)")
	AddHierarchyFlags(buildCmd, &buildFlags)
	AddDryRunFlag(buildCmd, &buildDryRun)
	AddAllFlag(buildCmd, "Build all matching workspaces (use with -e/-d/-a to scope)")
	buildCmd.Flags().BoolVar(&buildDetach, "detach", false, "Run in background; monitor with 'dvm build status'")
	buildCmd.Flags().IntVar(&buildConcurrency, "concurrency", 8, "Max parallel builds (capped at 2x CPU cores)")
	buildCmd.AddCommand(buildStatusCmd)
}
