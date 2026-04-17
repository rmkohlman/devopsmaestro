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
	buildCleanCache  bool
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

Batch Builds:
  Scope filters (-e, -d, -a, -w) automatically build ALL matching workspaces
  in parallel. For example, if --ecosystem matches 5 workspaces, all 5 are built.
  Use --all / -A to build every workspace, optionally narrowed by scope filters.
  Use --concurrency to control parallelism (default: 8).

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
  -e, --ecosystem   Filter by ecosystem — builds ALL workspaces in ecosystem
  -d, --domain      Filter by domain — builds ALL workspaces in domain
  -a, --app         Filter by app — builds ALL workspaces for app
  -w, --workspace   Filter by workspace name
  -A, --all         Build all workspaces (combine with scope flags to narrow)
  --no-cache        Build without using registry cache (pull fresh from upstream)
  --push            Push built image to local registry after build
  --registry        Override registry endpoint (default: from config)

Examples:
  dvm build                               # Build active workspace
  dvm build --force
  dvm build --no-cache
  dvm build --push                        # Push to local registry
  dvm build -a portal                     # Build all workspaces in 'portal' app
  dvm build -e healthcare                 # Build all workspaces in ecosystem
  dvm build -e healthcare -d payments     # Build all workspaces in domain
  dvm build -A                            # Build every workspace
  dvm build -A -e healthcare              # Build all in ecosystem (same as -e alone)
  DVM_PLATFORM=colima dvm build
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate --all mutual exclusion with hierarchy flags
		allSet, _ := cmd.Flags().GetBool("all")
		if err := validateBuildAllMutualExclusion(allSet, buildFlags); err != nil {
			return err
		}

		// Route to parallel build path when --all or scope flags are set.
		// Issue #215: scope filters (--ecosystem, --domain, --app) auto-build
		// ALL matching workspaces instead of erroring with "ambiguous workspace
		// selection". This routes through resolveWorkspacesForParallelBuild
		// (which returns all matches) rather than resolveFromHierarchyFlags
		// (which throws AmbiguousError on multi-match).
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
	buildCmd.Flags().DurationVar(&buildTimeout, "timeout", 10*time.Minute, "Timeout for the build operation (e.g., 30m, 1h)")
	AddHierarchyFlags(buildCmd, &buildFlags)
	AddDryRunFlag(buildCmd, &buildDryRun)
	AddAllFlag(buildCmd, "Build all matching workspaces (use with -e/-d/-a to scope)")
	buildCmd.Flags().BoolVar(&buildDetach, "detach", false, "Run in background; monitor with 'dvm build status'")
	buildCmd.Flags().IntVar(&buildConcurrency, "concurrency", 8, "Max parallel builds (capped at 2x CPU cores)")
	buildCmd.Flags().BoolVar(&buildCleanCache, "clean-cache", false, "Aggressively clean before/after build: prune BuildKit cache, remove old workspace images, use registry cache, minimize disk footprint")
	buildCmd.AddCommand(buildStatusCmd)
}
