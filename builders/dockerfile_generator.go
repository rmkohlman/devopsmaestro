package builders

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"devopsmaestro/models"
	"devopsmaestro/utils"
	"github.com/rmkohlman/MaestroNvim/nvimops/plugin"
	"github.com/rmkohlman/MaestroSDK/paths"
)

// curlFlags provides hardened curl flags for builder stage downloads.
// -f: Fail on HTTP errors (4xx/5xx) instead of silently returning error pages
// -s: Silent mode (no progress meter)
// -S: Show errors even in silent mode
// -L: Follow redirects
// --retry 3: Retry transient failures up to 3 times
// --connect-timeout 30: Timeout connection phase after 30 seconds
const curlFlags = "-fsSL --retry 3 --connect-timeout 30"

// imageDigests maps known base image references (image:tag) to their SHA256 manifest digests.
// Pinning to digests ensures reproducible, tamper-proof builds — a compromised registry
// cannot serve a different image for the same mutable tag.
//
// These digests should be updated periodically (e.g., via Dependabot or Renovate).
// To refresh a digest: docker pull <image:tag> && docker inspect <image:tag> --format='{{index .RepoDigests 0}}'
//
// Static builder images (version-locked):
var imageDigests = map[string]string{
	// Builder stage images
	"debian:bookworm-slim": "sha256:f06537653ac770703bc45b4b113475bd402f451e85223f0f2837acbf89ab020a",
	"alpine:3.20":          "sha256:a4f4213abb84c497377b8544c81b3564f313746700372ec4fe84653e4fb03805",
	"ubuntu:22.04":         "sha256:ce4a593b4e323dcc3dd728e397e0a866a1bf516a1b7c31d6aa06991baec4f2e0",

	// Versioned language base images (common defaults)
	"python:3.11-slim-bookworm": "sha256:420310dd2ff7895895f0f1f9d15cae5a95dabceb8f1d6b9a23ef33c2c1c542c3",
	"python:3.12-slim-bookworm": "sha256:31c0807da611e2e377a2e9b566ad4eb038ac5a5838cbbbe6f2262259b5dc77a0",
	"golang:1.22-alpine":        "sha256:1699c10032ca2582ec89a24a1312d986a3f094aed3d5c1147b19880afe40e052",
	"golang:1.23-alpine":        "sha256:383395b794dffa5b53012a212365d40c8e37109a626ca30d6151c8348d380b5f",
	"node:20-alpine":            "sha256:b88333c42c23fbd91596ebd7fd10de239cedab9617de04142dde7315e3bc0afa",
	"node:22-alpine":            "sha256:8094c002d08262dba12645a3b4a15cd6cd627d30bc782f53229a2ec13ee22a00",
}

// pinnedImage returns a digest-pinned image reference if a known digest exists for the
// given image:tag. If no digest is known (e.g., user-specified custom image or unlisted
// version), it returns the original tag-based reference unchanged.
//
// Examples:
//
//	pinnedImage("debian:bookworm-slim")       → "debian:bookworm-slim@sha256:f065..."
//	pinnedImage("python:3.11-slim-bookworm")  → "python:3.11-slim-bookworm@sha256:4203..."
//	pinnedImage("myregistry/custom:v1")       → "myregistry/custom:v1"  (no digest known)
func pinnedImage(image string) string {
	if digest, ok := imageDigests[image]; ok {
		return image + "@" + digest
	}
	return image
}

// pinnedImageComment returns a Dockerfile comment if the image is NOT pinned to a digest.
// Returns empty string if the image is pinned (no warning needed).
func pinnedImageComment(image string) string {
	if _, ok := imageDigests[image]; !ok {
		return fmt.Sprintf("# WARNING: %s not pinned to digest — consider adding to imageDigests\n", image)
	}
	return ""
}

// DockerfileGenerator defines the interface for generating Dockerfiles for dev containers.
// Implementations produce Dockerfile content optimized with BuildKit features
// (parallel multi-stage builds, cache mounts) for a specific language/workspace combination.
type DockerfileGenerator interface {
	// SetPluginManifest sets the plugin manifest for conditional feature detection.
	// This should be called before Generate() to enable plugin-aware Dockerfile generation.
	SetPluginManifest(manifest *plugin.PluginManifest)

	// Generate creates a Dockerfile.dvm with dev stage.
	// Uses BuildKit features: parallel multi-stage builds, cache mounts.
	Generate() (string, error)
}

// Compile-time interface compliance check
var _ DockerfileGenerator = (*DefaultDockerfileGenerator)(nil)

// DefaultDockerfileGenerator is the standard implementation of DockerfileGenerator
// that generates Dockerfiles for dev containers.
type DefaultDockerfileGenerator struct {
	workspace           *models.Workspace
	workspaceYAML       models.WorkspaceSpec
	language            string
	version             string
	appPath             string
	stagingDir          string // Explicit staging directory path (used for file existence checks)
	baseDockerfile      string
	pluginManifest      *plugin.PluginManifest
	pathConfig          *paths.PathConfig
	isAlpine            bool // computed in generateBaseStage() based on the actual image chosen
	privateRepoInfo     *utils.PrivateRepoInfo
	additionalBuildArgs []string // Extra ARG names to declare in the Dockerfile (names only, not values)
}

// DockerfileGeneratorOptions contains all configuration for creating a DockerfileGenerator.
type DockerfileGeneratorOptions struct {
	Workspace           *models.Workspace
	WorkspaceSpec       models.WorkspaceSpec
	Language            string
	Version             string
	AppPath             string
	StagingDir          string // Explicit staging directory path (build context); if empty, falls back to PathConfig-based lookup
	BaseDockerfile      string
	PathConfig          *paths.PathConfig      // Injected for filesystem operations (nil = fallback to os.UserHomeDir)
	PrivateRepoInfo     *utils.PrivateRepoInfo // Injected for system dep detection (nil = auto-detect at build time)
	AdditionalBuildArgs []string               // Extra ARG names to declare in the Dockerfile (names only, not values)
}

// NewDockerfileGenerator creates a new Dockerfile generator.
// Returns the DockerfileGenerator interface for loose coupling.
func NewDockerfileGenerator(opts DockerfileGeneratorOptions) DockerfileGenerator {
	return &DefaultDockerfileGenerator{
		workspace:           opts.Workspace,
		workspaceYAML:       opts.WorkspaceSpec,
		language:            opts.Language,
		version:             opts.Version,
		appPath:             opts.AppPath,
		stagingDir:          opts.StagingDir,
		baseDockerfile:      opts.BaseDockerfile,
		pathConfig:          opts.PathConfig,
		privateRepoInfo:     opts.PrivateRepoInfo,
		additionalBuildArgs: opts.AdditionalBuildArgs,
	}
}

// SetPluginManifest sets the plugin manifest for conditional feature detection.
// This should be called before Generate() to enable plugin-aware Dockerfile generation.
func (g *DefaultDockerfileGenerator) SetPluginManifest(manifest *plugin.PluginManifest) {
	g.pluginManifest = manifest
}

// Generate creates a Dockerfile.dvm with dev stage
// Uses BuildKit features: parallel multi-stage builds, cache mounts
func (g *DefaultDockerfileGenerator) Generate() (string, error) {
	if g.workspace == nil {
		return "", fmt.Errorf("workspace must not be nil")
	}

	var dockerfile strings.Builder

	// Header comment
	dockerfile.WriteString("# syntax=docker/dockerfile:1\n")
	dockerfile.WriteString("# Generated by DevOpsMaestro\n")
	dockerfile.WriteString("# Development container with tools for coding\n")
	dockerfile.WriteString("# Optimized with BuildKit: parallel stages + cache mounts\n")
	dockerfile.WriteString("# Base images pinned to SHA256 digests for reproducible builds\n\n")

	// Use injected private repo info if available, otherwise auto-detect
	privateRepoInfo := g.privateRepoInfo
	if privateRepoInfo == nil {
		privateRepoInfo = utils.DetectPrivateRepos(g.appPath, g.language)
	}

	// For MVP: Always generate from scratch
	// Future: Support extending existing production Dockerfiles
	g.generateBaseStage(&dockerfile, privateRepoInfo)

	// Compute builder stages once — single source of truth for both emission and COPY
	stages := g.activeBuilderStages()

	// Parallel builder stages - BuildKit runs these concurrently
	g.emitBuilderStages(&dockerfile, stages)

	// Dev stage
	dockerfile.WriteString("# Development stage with additional tools\n")
	dockerfile.WriteString("FROM base AS dev\n\n")

	// Switch to root for installing dev tools
	dockerfile.WriteString("USER root\n\n")

	// Re-declare build args in dev stage — Docker ARG values do not carry across FROM boundaries.
	// Without this, proxy vars (http_proxy, https_proxy), PIP_INDEX_URL, and other user build args
	// would not be available to RUN commands in the dev stage (e.g., npm install -g neovim).
	g.emitAdditionalBuildArgs(&dockerfile, privateRepoInfo.RequiredBuildArgs)

	// Copy binaries from parallel builder stages
	g.emitCopyFromBuilders(&dockerfile, stages)

	// Generate dev stage content based on language and config
	g.generateDevStage(&dockerfile)

	// Create dev user if not exists
	g.generateDevUser(&dockerfile)

	// Create workspace directory with correct ownership (before switching to non-root user)
	// For bind mounts, host permissions take precedence, but this ensures the directory
	// exists with correct ownership when no volume is mounted
	workdir := g.workspaceYAML.Container.WorkingDir
	if workdir == "" {
		workdir = "/workspace"
	}
	user := g.effectiveUser()
	dockerfile.WriteString("# Ensure workspace directory exists with correct ownership\n")
	dockerfile.WriteString(fmt.Sprintf("RUN mkdir -p %s && chown %s:%s %s\n\n", workdir, user, user, workdir))

	// Add Neovim configuration after user is created
	if err := g.generateNvimSection(&dockerfile); err != nil {
		return "", fmt.Errorf("nvim section: %w", err)
	}

	// Switch to dev user
	dockerfile.WriteString(fmt.Sprintf("USER %s\n\n", user))

	// Set working directory
	dockerfile.WriteString(fmt.Sprintf("WORKDIR %s\n\n", workdir))

	// Set command
	if len(g.workspaceYAML.Container.Command) > 0 {
		cmd := strings.Join(g.workspaceYAML.Container.Command, "\", \"")
		dockerfile.WriteString(fmt.Sprintf("CMD [\"%s\"]\n", cmd))
	} else {
		// Default to a long-running process
		dockerfile.WriteString("CMD [\"/bin/zsh\", \"-c\", \"tail -f /dev/null\"]\n")
	}

	return dockerfile.String(), nil
}

func (g *DefaultDockerfileGenerator) generateBaseStage(dockerfile *strings.Builder, privateRepoInfo *utils.PrivateRepoInfo) {
	dockerfile.WriteString("# Base stage (auto-generated)\n")

	switch g.language {
	case "python":
		version := g.effectiveVersion()
		g.isAlpine = false
		baseImage := fmt.Sprintf("python:%s-slim-bookworm", version)
		dockerfile.WriteString(pinnedImageComment(baseImage))
		dockerfile.WriteString(fmt.Sprintf("FROM %s AS base\n\n", pinnedImage(baseImage)))

		// Declare build args after FROM so they're available in RUN commands
		if len(privateRepoInfo.RequiredBuildArgs) > 0 {
			dockerfile.WriteString("# Build arguments for private repositories\n")
			for _, arg := range privateRepoInfo.RequiredBuildArgs {
				dockerfile.WriteString(fmt.Sprintf("ARG %s\n", arg))
			}
			dockerfile.WriteString("\n")
		}

		// Emit additional build args (de-duplicated with RequiredBuildArgs)
		g.emitAdditionalBuildArgs(dockerfile, privateRepoInfo.RequiredBuildArgs)

		// Install git if needed for private repos
		if privateRepoInfo.NeedsGit {
			packages := []string{"git", "build-essential"}
			if privateRepoInfo.NeedsSSH {
				packages = append(packages, "openssh-client")
			}
			// Merge auto-detected system deps (e.g., libpq-dev for psycopg2)
			packages = appendUnique(packages, privateRepoInfo.SystemDeps...)
			// Merge user-specified base stage packages
			packages = appendUnique(packages, g.workspaceYAML.Build.BaseStage.Packages...)

			if len(privateRepoInfo.SystemDeps) > 0 {
				dockerfile.WriteString("# Install git for private repositories + auto-detected system dependencies\n")
				dockerfile.WriteString(fmt.Sprintf("# Auto-detected: %s\n", formatSystemDepSources(privateRepoInfo)))
			} else {
				dockerfile.WriteString("# Install git for private repositories\n")
			}
			dockerfile.WriteString(g.aptCacheMounts())
			dockerfile.WriteString("    apt-get update && apt-get install -y --no-install-recommends --fix-broken \\\n")
			for _, pkg := range packages {
				dockerfile.WriteString(fmt.Sprintf("    %s \\\n", pkg))
			}
			// Remove trailing backslash - end the command
			// Actually, we need to just end cleanly. With cache mounts, no cleanup needed.
			dockerfile.WriteString("    && true\n\n")

			// Setup SSH for git if needed
			if privateRepoInfo.NeedsSSH {
				dockerfile.WriteString("# Setup SSH for git operations\n")
				dockerfile.WriteString("RUN mkdir -p /root/.ssh && chmod 700 /root/.ssh\n")
				dockerfile.WriteString("# Mount SSH keys at build time using BuildKit secrets:\n")
				dockerfile.WriteString("# --mount=type=ssh\n")
				dockerfile.WriteString("RUN --mount=type=ssh \\\n")
				dockerfile.WriteString("    ssh-keyscan github.com >> /root/.ssh/known_hosts && \\\n")
				dockerfile.WriteString("    ssh-keyscan gitlab.com >> /root/.ssh/known_hosts\n\n")
			}
		} else {
			packages := []string{"build-essential"}
			// Merge auto-detected system deps (e.g., libpq-dev for psycopg2)
			packages = appendUnique(packages, privateRepoInfo.SystemDeps...)
			// Merge user-specified base stage packages
			packages = appendUnique(packages, g.workspaceYAML.Build.BaseStage.Packages...)

			// Add comment about auto-detected system deps
			if len(privateRepoInfo.SystemDeps) > 0 {
				dockerfile.WriteString("# Install build dependencies\n")
				dockerfile.WriteString(fmt.Sprintf("# Auto-detected system dependencies: %s\n", formatSystemDepSources(privateRepoInfo)))
			} else {
				dockerfile.WriteString("# Install build dependencies\n")
			}
			dockerfile.WriteString(g.aptCacheMounts())
			dockerfile.WriteString("    apt-get update && apt-get install -y --no-install-recommends --fix-broken \\\n")
			for _, pkg := range packages {
				dockerfile.WriteString(fmt.Sprintf("    %s \\\n", pkg))
			}
			dockerfile.WriteString("    && true\n\n")
		}

		// CA certificate injection (before pip install so pip trusts corporate CAs)
		g.emitCACertSection(dockerfile)

		// Dispatch on GitURLType for correct pip install strategy:
		//   "https" → pip install with ARG-based env vars (pip expands ${VAR} natively)
		//   "ssh"   → pip install with --mount=type=ssh
		//   "mixed" → pip install with SSH mount + ARG-based env vars
		//   default → plain pip install (no private repos)
		//
		// Check the staging directory (build context) for requirements.txt, not the
		// source path, to ensure the COPY command only appears when the file is
		// actually present in the Docker build context (fixes #228).
		reqCheckDir := g.appPath
		if sd := g.effectiveStagingDir(); sd != "" {
			reqCheckDir = sd
		}
		requirementsPath := filepath.Join(reqCheckDir, "requirements.txt")
		if _, err := os.Stat(requirementsPath); err == nil {
			switch privateRepoInfo.GitURLType {
			case "https":
				dockerfile.WriteString("# Install dependencies (pip expands ${VAR} from build args)\n")
				dockerfile.WriteString("# Falls back to direct PyPI access if HTTP proxy is unreachable\n")
				dockerfile.WriteString("COPY requirements.txt /tmp/\n")
				dockerfile.WriteString("RUN --mount=type=cache,target=/root/.cache/pip \\\n")
				dockerfile.WriteString("    pip install -r /tmp/requirements.txt \\\n")
				dockerfile.WriteString("    || (unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy \\\n")
				dockerfile.WriteString("    && pip install -r /tmp/requirements.txt)\n\n")

			case "ssh":
				dockerfile.WriteString("# Install dependencies with SSH key mount\n")
				dockerfile.WriteString("# Falls back to direct PyPI access if HTTP proxy is unreachable\n")
				dockerfile.WriteString("COPY requirements.txt /tmp/\n")
				dockerfile.WriteString("RUN --mount=type=ssh \\\n")
				dockerfile.WriteString("    --mount=type=cache,target=/root/.cache/pip \\\n")
				dockerfile.WriteString("    pip install -r /tmp/requirements.txt \\\n")
				dockerfile.WriteString("    || (unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy \\\n")
				dockerfile.WriteString("    && pip install -r /tmp/requirements.txt)\n\n")

			case "mixed":
				dockerfile.WriteString("# Install dependencies with SSH mount (pip expands ${VAR} from build args)\n")
				dockerfile.WriteString("# Falls back to direct PyPI access if HTTP proxy is unreachable\n")
				dockerfile.WriteString("COPY requirements.txt /tmp/\n")
				dockerfile.WriteString("RUN --mount=type=ssh \\\n")
				dockerfile.WriteString("    --mount=type=cache,target=/root/.cache/pip \\\n")
				dockerfile.WriteString("    pip install -r /tmp/requirements.txt \\\n")
				dockerfile.WriteString("    || (unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy \\\n")
				dockerfile.WriteString("    && pip install -r /tmp/requirements.txt)\n\n")

			default:
				dockerfile.WriteString("# Copy requirements and install\n")
				dockerfile.WriteString("# Falls back to direct PyPI access if HTTP proxy is unreachable\n")
				dockerfile.WriteString("COPY requirements.txt /tmp/\n")
				dockerfile.WriteString("RUN --mount=type=cache,target=/root/.cache/pip \\\n")
				dockerfile.WriteString("    pip install -r /tmp/requirements.txt \\\n")
				dockerfile.WriteString("    || (unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy \\\n")
				dockerfile.WriteString("    && pip install -r /tmp/requirements.txt)\n\n")
			}
		} else {
			dockerfile.WriteString("# No requirements.txt found — skipping pip install\n\n")
		}

	case "golang":
		version := g.effectiveVersion()
		g.isAlpine = true
		baseImage := fmt.Sprintf("golang:%s-alpine", version)
		dockerfile.WriteString(pinnedImageComment(baseImage))
		dockerfile.WriteString(fmt.Sprintf("FROM %s AS base\n\n", pinnedImage(baseImage)))

		// Emit additional build args (de-duplicated with RequiredBuildArgs)
		g.emitAdditionalBuildArgs(dockerfile, privateRepoInfo.RequiredBuildArgs)

		dockerfile.WriteString("# Install basic dependencies\n")
		dockerfile.WriteString(g.apkCacheMounts())
		if len(g.workspaceYAML.Build.CACerts) > 0 {
			dockerfile.WriteString("    apk add git ca-certificates\n\n")
		} else {
			dockerfile.WriteString("    apk add git\n\n")
		}

		// CA certificate injection
		g.emitCACertSection(dockerfile)

		// Configure git for private repos with token
		if privateRepoInfo.NeedsGitConfig {
			dockerfile.WriteString("# Configure git for private repositories\n")
			for _, arg := range privateRepoInfo.RequiredBuildArgs {
				if arg == "GITHUB_TOKEN" {
					dockerfile.WriteString(fmt.Sprintf("ARG %s\n", arg))
					dockerfile.WriteString("RUN git config --global url.\"https://${GITHUB_TOKEN}@github.com/\".insteadOf \"https://github.com/\"\n\n")
				}
			}
		}

	case "nodejs":
		version := g.effectiveVersion()
		g.isAlpine = true
		baseImage := fmt.Sprintf("node:%s-alpine", version)
		dockerfile.WriteString(pinnedImageComment(baseImage))
		dockerfile.WriteString(fmt.Sprintf("FROM %s AS base\n\n", pinnedImage(baseImage)))

		// Emit additional build args (de-duplicated with RequiredBuildArgs)
		g.emitAdditionalBuildArgs(dockerfile, privateRepoInfo.RequiredBuildArgs)

		if privateRepoInfo.NeedsGit {
			dockerfile.WriteString("# Install git for private repositories\n")
			dockerfile.WriteString(g.apkCacheMounts())
			if len(g.workspaceYAML.Build.CACerts) > 0 {
				dockerfile.WriteString("    apk add git ca-certificates\n\n")
			} else {
				dockerfile.WriteString("    apk add git\n\n")
			}
		} else if len(g.workspaceYAML.Build.CACerts) > 0 {
			dockerfile.WriteString("# Install ca-certificates for custom CA support\n")
			dockerfile.WriteString(g.apkCacheMounts())
			dockerfile.WriteString("    apk add ca-certificates\n\n")
		}

		// CA certificate injection
		g.emitCACertSection(dockerfile)

		// Setup .npmrc for private packages
		if len(privateRepoInfo.RequiredBuildArgs) > 0 {
			dockerfile.WriteString("# Configure npm for private packages\n")
			for _, arg := range privateRepoInfo.RequiredBuildArgs {
				if arg == "NPM_TOKEN" {
					dockerfile.WriteString(fmt.Sprintf("ARG %s\n", arg))
					dockerfile.WriteString("RUN echo \"//registry.npmjs.org/:_authToken=${NPM_TOKEN}\" > ~/.npmrc\n\n")
				}
			}
		}

	default:
		// Generic Ubuntu base
		g.isAlpine = false
		baseImage := "ubuntu:22.04"
		dockerfile.WriteString(fmt.Sprintf("FROM %s AS base\n\n", pinnedImage(baseImage)))

		// Emit additional build args (de-duplicated with RequiredBuildArgs)
		g.emitAdditionalBuildArgs(dockerfile, privateRepoInfo.RequiredBuildArgs)

		packages := []string{"ca-certificates"}
		if privateRepoInfo.NeedsGit {
			packages = append(packages, "git")
		}

		dockerfile.WriteString(g.aptCacheMounts())
		dockerfile.WriteString("    apt-get update && apt-get install -y --no-install-recommends --fix-broken \\\n")
		for i, pkg := range packages {
			if i < len(packages)-1 {
				dockerfile.WriteString(fmt.Sprintf("    %s \\\n", pkg))
			} else {
				dockerfile.WriteString(fmt.Sprintf("    %s\n", pkg))
			}
		}
		dockerfile.WriteString("\n")
	}
}

// builderStage represents a parallel builder stage and its corresponding COPY directives.
// This struct is the single source of truth for which builders are emitted and what
// COPY --from= directives are generated, eliminating the risk of conditional duplication.
type builderStage struct {
	name      string                            // stage name (e.g., "neovim-builder")
	emitFunc  func(dockerfile *strings.Builder) // generates the FROM ... RUN block
	copyLines []string                          // COPY --from=... lines for the dev stage
}

// activeBuilderStages returns the ordered list of builder stages that should be emitted
// for this workspace configuration. This is the single source of truth: both
// generateBuilderStages() and copyFromBuilders() iterate this list.
func (g *DefaultDockerfileGenerator) activeBuilderStages() []builderStage {
	isAlpine := g.isAlpineImage()
	var stages []builderStage

	// Neovim builder (only for Debian — Alpine uses apk)
	if !isAlpine {
		stages = append(stages, builderStage{
			name:     "neovim-builder",
			emitFunc: g.generateNeovimBuilder,
			copyLines: []string{
				"COPY --from=neovim-builder /opt/nvim/ /opt/nvim/",
				"RUN ln -sf /opt/nvim/bin/nvim /usr/local/bin/nvim",
			},
		})
	}

	// Lazygit builder
	stages = append(stages, builderStage{
		name: "lazygit-builder",
		emitFunc: func(df *strings.Builder) {
			g.generateLazygitBuilder(df, isAlpine)
		},
		copyLines: []string{
			"COPY --from=lazygit-builder /usr/local/bin/lazygit /usr/local/bin/lazygit",
		},
	})

	// Starship builder
	if g.workspaceYAML.Shell.Theme == "starship" || g.workspaceYAML.Shell.Theme == "" {
		stages = append(stages, builderStage{
			name:     "starship-builder",
			emitFunc: g.generateStarshipBuilder,
			copyLines: []string{
				"COPY --from=starship-builder /usr/local/bin/starship /usr/local/bin/starship",
			},
		})
	}

	// Tree-sitter CLI builder
	stages = append(stages, builderStage{
		name: "treesitter-builder",
		emitFunc: func(df *strings.Builder) {
			g.generateTreeSitterBuilder(df, isAlpine)
		},
		copyLines: []string{
			"COPY --from=treesitter-builder /usr/local/bin/tree-sitter /usr/local/bin/tree-sitter",
		},
	})

	// Go tools builder (only for golang workspaces with go-installable tools)
	if g.hasGoToolsBuilder() {
		stages = append(stages, builderStage{
			name:     "go-tools-builder",
			emitFunc: g.generateGoToolsBuilder,
			copyLines: []string{
				"COPY --from=go-tools-builder /go/bin/ /go/bin/",
			},
		})
	}

	// Opencode builder (opt-in via workspace tools config)
	if g.workspaceYAML.Tools.Opencode {
		stages = append(stages, builderStage{
			name: "opencode-builder",
			emitFunc: func(df *strings.Builder) {
				g.generateOpencodeBuilder(df, isAlpine)
			},
			copyLines: []string{
				"COPY --from=opencode-builder /usr/local/bin/opencode /usr/local/bin/opencode",
			},
		})
	}

	return stages
}

// emitBuilderStages emits parallel multi-stage builds for binary downloads.
// BuildKit runs these concurrently since they have no dependencies on each other.
//
// Builder image selection rationale:
//   - Neovim (Debian only): GitHub releases are glibc-linked, won't work on Alpine musl.
//     Alpine gets neovim via apk in the merged package install instead.
//   - Lazygit: Uses matching base (Alpine for Alpine targets, Debian for Debian) because
//     the binary is statically linked but download tools (curl vs wget, sed behavior) differ.
//   - Starship: Always Debian for consistency with Neovim builder (install script uses POSIX sh).
//   - Tree-sitter CLI: Alpine (small image, binary is statically linked).
//   - Go tools: Uses golang:X-alpine with a minimum version floor (goToolsMinGoVersion)
//     to satisfy gopls@latest requirements. May differ from the workspace base Go version.
//
// Cache mount strategy: Builder stages use cache mounts for package manager directories
// (apt, apk). BuildKit handles concurrent cache mounts with copy-on-write semantics —
// parallel stages sharing /var/cache/apk or /var/cache/apt will NOT deadlock or corrupt.
// go-tools-builder additionally uses Go module/build cache mounts.
// The dev stage (long-lived) uses cache mounts for all package managers.
func (g *DefaultDockerfileGenerator) emitBuilderStages(dockerfile *strings.Builder, stages []builderStage) {
	for _, stage := range stages {
		stage.emitFunc(dockerfile)
	}
}

// generateNeovimBuilder creates a parallel stage to download Neovim (Debian only).
// Pinned to a specific version with SHA256 checksum verification (see checksums.go).
func (g *DefaultDockerfileGenerator) generateNeovimBuilder(dockerfile *strings.Builder) {
	dockerfile.WriteString("# --- Parallel builder: Neovim ---\n")
	dockerfile.WriteString(fmt.Sprintf("FROM %s AS neovim-builder\n", pinnedImage("debian:bookworm-slim")))
	dockerfile.WriteString(g.aptCacheMountsLocked())
	dockerfile.WriteString("    set -e && \\\n")
	dockerfile.WriteString("    apt-get update && apt-get install -y --no-install-recommends curl ca-certificates && \\\n")
	dockerfile.WriteString("    ARCH=$(dpkg --print-architecture 2>/dev/null || uname -m) && \\\n")
	dockerfile.WriteString("    if [ \"$ARCH\" = \"arm64\" ] || [ \"$ARCH\" = \"aarch64\" ]; then \\\n")
	dockerfile.WriteString(fmt.Sprintf("        NVIM_ARCH=\"nvim-linux-arm64\"; NVIM_SHA256=\"%s\"; \\\n", neovimChecksumArm64))
	dockerfile.WriteString("    elif [ \"$ARCH\" = \"amd64\" ] || [ \"$ARCH\" = \"x86_64\" ]; then \\\n")
	dockerfile.WriteString(fmt.Sprintf("        NVIM_ARCH=\"nvim-linux-x86_64\"; NVIM_SHA256=\"%s\"; \\\n", neovimChecksumX86_64))
	dockerfile.WriteString("    else \\\n")
	dockerfile.WriteString("        echo \"ERROR: Unsupported architecture: $ARCH\"; exit 1; \\\n")
	dockerfile.WriteString("    fi && \\\n")
	dockerfile.WriteString(fmt.Sprintf("    curl %s -o \"${NVIM_ARCH}.tar.gz\" \"https://github.com/neovim/neovim/releases/download/v%s/${NVIM_ARCH}.tar.gz\" && \\\n", curlFlags, neovimVersion))
	dockerfile.WriteString("    echo \"${NVIM_SHA256}  ${NVIM_ARCH}.tar.gz\" | sha256sum -c - && \\\n")
	dockerfile.WriteString("    tar -C /opt -xzf \"${NVIM_ARCH}.tar.gz\" && \\\n")
	dockerfile.WriteString("    mv /opt/${NVIM_ARCH} /opt/nvim && \\\n")
	dockerfile.WriteString("    rm \"${NVIM_ARCH}.tar.gz\" && \\\n")
	dockerfile.WriteString("    test -x /opt/nvim/bin/nvim\n\n")
}

// generateLazygitBuilder creates a parallel stage to download lazygit.
// Pinned to a specific version with SHA256 checksum verification (see checksums.go).
func (g *DefaultDockerfileGenerator) generateLazygitBuilder(dockerfile *strings.Builder, isAlpine bool) {
	dockerfile.WriteString("# --- Parallel builder: lazygit ---\n")
	if isAlpine {
		dockerfile.WriteString(fmt.Sprintf("FROM %s AS lazygit-builder\n", pinnedImage("alpine:3.20")))
		dockerfile.WriteString(g.apkCacheMountsLocked())
		dockerfile.WriteString("    set -e && \\\n")
		dockerfile.WriteString("    apk add --no-cache curl && \\\n")
		dockerfile.WriteString("    ARCH=$(uname -m) && \\\n")
		dockerfile.WriteString("    if [ \"$ARCH\" = \"aarch64\" ]; then \\\n")
		dockerfile.WriteString(fmt.Sprintf("        LG_ARCH=\"arm64\"; LG_SHA256=\"%s\"; \\\n", lazygitChecksumArm64))
		dockerfile.WriteString("    elif [ \"$ARCH\" = \"x86_64\" ]; then \\\n")
		dockerfile.WriteString(fmt.Sprintf("        LG_ARCH=\"x86_64\"; LG_SHA256=\"%s\"; \\\n", lazygitChecksumX86_64))
		dockerfile.WriteString("    else \\\n")
		dockerfile.WriteString("        echo \"ERROR: Unsupported architecture: $ARCH\"; exit 1; \\\n")
		dockerfile.WriteString("    fi && \\\n")
	} else {
		dockerfile.WriteString(fmt.Sprintf("FROM %s AS lazygit-builder\n", pinnedImage("debian:bookworm-slim")))
		dockerfile.WriteString(g.aptCacheMountsLocked())
		dockerfile.WriteString("    set -e && \\\n")
		dockerfile.WriteString("    apt-get update && apt-get install -y --no-install-recommends curl ca-certificates && \\\n")
		dockerfile.WriteString("    ARCH=$(dpkg --print-architecture 2>/dev/null || uname -m) && \\\n")
		dockerfile.WriteString("    if [ \"$ARCH\" = \"arm64\" ] || [ \"$ARCH\" = \"aarch64\" ]; then \\\n")
		dockerfile.WriteString(fmt.Sprintf("        LG_ARCH=\"arm64\"; LG_SHA256=\"%s\"; \\\n", lazygitChecksumArm64))
		dockerfile.WriteString("    elif [ \"$ARCH\" = \"amd64\" ] || [ \"$ARCH\" = \"x86_64\" ]; then \\\n")
		dockerfile.WriteString(fmt.Sprintf("        LG_ARCH=\"x86_64\"; LG_SHA256=\"%s\"; \\\n", lazygitChecksumX86_64))
		dockerfile.WriteString("    else \\\n")
		dockerfile.WriteString("        echo \"ERROR: Unsupported architecture: $ARCH\"; exit 1; \\\n")
		dockerfile.WriteString("    fi && \\\n")
	}
	// Shared download + verify logic (identical for Alpine and Debian)
	dockerfile.WriteString(fmt.Sprintf("    curl %s -o lazygit.tar.gz \"https://github.com/jesseduffield/lazygit/releases/download/v%s/lazygit_%s_Linux_${LG_ARCH}.tar.gz\" && \\\n", curlFlags, lazygitVersion, lazygitVersion))
	dockerfile.WriteString("    echo \"${LG_SHA256}  lazygit.tar.gz\" | sha256sum -c - && \\\n")
	dockerfile.WriteString("    tar xf lazygit.tar.gz lazygit && \\\n")
	dockerfile.WriteString("    install lazygit /usr/local/bin && \\\n")
	dockerfile.WriteString("    rm lazygit lazygit.tar.gz && \\\n")
	dockerfile.WriteString("    test -x /usr/local/bin/lazygit\n\n")
}

// generateStarshipBuilder creates a parallel stage to download Starship prompt.
// Downloads the pre-built musl binary directly (works on both Alpine and Debian)
// with SHA256 checksum verification (see checksums.go).
func (g *DefaultDockerfileGenerator) generateStarshipBuilder(dockerfile *strings.Builder) {
	dockerfile.WriteString("# --- Parallel builder: Starship prompt ---\n")
	dockerfile.WriteString(fmt.Sprintf("FROM %s AS starship-builder\n", pinnedImage("debian:bookworm-slim")))
	dockerfile.WriteString(g.aptCacheMountsLocked())
	dockerfile.WriteString("    set -e && \\\n")
	dockerfile.WriteString("    apt-get update && apt-get install -y --no-install-recommends curl ca-certificates && \\\n")
	dockerfile.WriteString("    ARCH=$(dpkg --print-architecture 2>/dev/null || uname -m) && \\\n")
	dockerfile.WriteString("    if [ \"$ARCH\" = \"arm64\" ] || [ \"$ARCH\" = \"aarch64\" ]; then \\\n")
	dockerfile.WriteString(fmt.Sprintf("        STARSHIP_ARCH=\"aarch64-unknown-linux-musl\"; STARSHIP_SHA256=\"%s\"; \\\n", starshipChecksumArm64))
	dockerfile.WriteString("    elif [ \"$ARCH\" = \"amd64\" ] || [ \"$ARCH\" = \"x86_64\" ]; then \\\n")
	dockerfile.WriteString(fmt.Sprintf("        STARSHIP_ARCH=\"x86_64-unknown-linux-musl\"; STARSHIP_SHA256=\"%s\"; \\\n", starshipChecksumX86_64))
	dockerfile.WriteString("    else \\\n")
	dockerfile.WriteString("        echo \"ERROR: Unsupported architecture: $ARCH\"; exit 1; \\\n")
	dockerfile.WriteString("    fi && \\\n")
	dockerfile.WriteString(fmt.Sprintf("    curl %s -o /tmp/starship.tar.gz \"https://github.com/starship/starship/releases/download/v%s/starship-${STARSHIP_ARCH}.tar.gz\" && \\\n", curlFlags, starshipVersion))
	dockerfile.WriteString("    echo \"${STARSHIP_SHA256}  /tmp/starship.tar.gz\" | sha256sum -c - && \\\n")
	dockerfile.WriteString("    tar -C /usr/local/bin -xzf /tmp/starship.tar.gz starship && \\\n")
	dockerfile.WriteString("    rm /tmp/starship.tar.gz && \\\n")
	dockerfile.WriteString("    test -x /usr/local/bin/starship\n\n")
}

// generateTreeSitterBuilder creates a parallel stage to download tree-sitter CLI.
// Pinned to a specific version with SHA256 checksum verification (see checksums.go).
func (g *DefaultDockerfileGenerator) generateTreeSitterBuilder(dockerfile *strings.Builder, isAlpine bool) {
	dockerfile.WriteString("# --- Parallel builder: tree-sitter CLI ---\n")
	if isAlpine {
		dockerfile.WriteString(fmt.Sprintf("FROM %s AS treesitter-builder\n", pinnedImage("alpine:3.20")))
		dockerfile.WriteString(g.apkCacheMountsLocked())
		dockerfile.WriteString("    set -e && \\\n")
		dockerfile.WriteString("    apk add --no-cache curl && \\\n")
		dockerfile.WriteString("    ARCH=$(uname -m) && \\\n")
		dockerfile.WriteString("    if [ \"$ARCH\" = \"aarch64\" ]; then \\\n")
		dockerfile.WriteString(fmt.Sprintf("        TS_ARCH=\"arm64\"; TS_SHA256=\"%s\"; \\\n", treeSitterChecksumArm64))
		dockerfile.WriteString("    elif [ \"$ARCH\" = \"x86_64\" ]; then \\\n")
		dockerfile.WriteString(fmt.Sprintf("        TS_ARCH=\"x64\"; TS_SHA256=\"%s\"; \\\n", treeSitterChecksumX64))
		dockerfile.WriteString("    else echo \"ERROR: Unsupported architecture: $ARCH\"; exit 1; fi && \\\n")
	} else {
		dockerfile.WriteString(fmt.Sprintf("FROM %s AS treesitter-builder\n", pinnedImage("debian:bookworm-slim")))
		dockerfile.WriteString(g.aptCacheMountsLocked())
		dockerfile.WriteString("    set -e && \\\n")
		dockerfile.WriteString("    apt-get update && apt-get install -y --no-install-recommends curl ca-certificates && \\\n")
		dockerfile.WriteString("    ARCH=$(dpkg --print-architecture 2>/dev/null || uname -m) && \\\n")
		dockerfile.WriteString("    if [ \"$ARCH\" = \"arm64\" ] || [ \"$ARCH\" = \"aarch64\" ]; then \\\n")
		dockerfile.WriteString(fmt.Sprintf("        TS_ARCH=\"arm64\"; TS_SHA256=\"%s\"; \\\n", treeSitterChecksumArm64))
		dockerfile.WriteString("    elif [ \"$ARCH\" = \"amd64\" ] || [ \"$ARCH\" = \"x86_64\" ]; then \\\n")
		dockerfile.WriteString(fmt.Sprintf("        TS_ARCH=\"x64\"; TS_SHA256=\"%s\"; \\\n", treeSitterChecksumX64))
		dockerfile.WriteString("    else echo \"ERROR: Unsupported architecture: $ARCH\"; exit 1; fi && \\\n")
	}
	// Shared download + verify logic (identical for Alpine and Debian)
	dockerfile.WriteString(fmt.Sprintf("    curl %s -o /tmp/ts.gz \"https://github.com/tree-sitter/tree-sitter/releases/download/v%s/tree-sitter-linux-${TS_ARCH}.gz\" && \\\n", curlFlags, treeSitterVersion))
	dockerfile.WriteString("    echo \"${TS_SHA256}  /tmp/ts.gz\" | sha256sum -c - && \\\n")
	dockerfile.WriteString("    gunzip /tmp/ts.gz && chmod +x /tmp/ts && mv /tmp/ts /usr/local/bin/tree-sitter && \\\n")
	dockerfile.WriteString("    test -x /usr/local/bin/tree-sitter\n\n")
}

// generateOpencodeBuilder creates a parallel stage to download the opencode CLI.
// Uses musl-linked static binaries (works on both Alpine and Debian).
// Pinned to a specific version with SHA256 checksum verification (see checksums.go).
//
// Note: opencode releases use "x64" for amd64 and "arm64" for arm64 in their
// download URLs (e.g., opencode-linux-x64-musl.tar.gz).
func (g *DefaultDockerfileGenerator) generateOpencodeBuilder(dockerfile *strings.Builder, isAlpine bool) {
	dockerfile.WriteString("# --- Parallel builder: opencode ---\n")
	if isAlpine {
		dockerfile.WriteString(fmt.Sprintf("FROM %s AS opencode-builder\n", pinnedImage("alpine:3.20")))
		dockerfile.WriteString(g.apkCacheMountsLocked())
		dockerfile.WriteString("    set -e && \\\n")
		dockerfile.WriteString("    apk add --no-cache curl && \\\n")
		dockerfile.WriteString("    ARCH=$(uname -m) && \\\n")
		dockerfile.WriteString("    if [ \"$ARCH\" = \"aarch64\" ]; then \\\n")
		dockerfile.WriteString(fmt.Sprintf("        OC_ARCH=\"arm64\"; OC_SHA256=\"%s\"; \\\n", opencodeChecksumArm64))
		dockerfile.WriteString("    elif [ \"$ARCH\" = \"x86_64\" ]; then \\\n")
		dockerfile.WriteString(fmt.Sprintf("        OC_ARCH=\"x64\"; OC_SHA256=\"%s\"; \\\n", opencodeChecksumAmd64))
		dockerfile.WriteString("    else \\\n")
		dockerfile.WriteString("        echo \"ERROR: Unsupported architecture: $ARCH\"; exit 1; \\\n")
		dockerfile.WriteString("    fi && \\\n")
	} else {
		dockerfile.WriteString(fmt.Sprintf("FROM %s AS opencode-builder\n", pinnedImage("debian:bookworm-slim")))
		dockerfile.WriteString(g.aptCacheMountsLocked())
		dockerfile.WriteString("    set -e && \\\n")
		dockerfile.WriteString("    apt-get update && apt-get install -y --no-install-recommends curl ca-certificates && \\\n")
		dockerfile.WriteString("    ARCH=$(dpkg --print-architecture 2>/dev/null || uname -m) && \\\n")
		dockerfile.WriteString("    if [ \"$ARCH\" = \"arm64\" ] || [ \"$ARCH\" = \"aarch64\" ]; then \\\n")
		dockerfile.WriteString(fmt.Sprintf("        OC_ARCH=\"arm64\"; OC_SHA256=\"%s\"; \\\n", opencodeChecksumArm64))
		dockerfile.WriteString("    elif [ \"$ARCH\" = \"amd64\" ] || [ \"$ARCH\" = \"x86_64\" ]; then \\\n")
		dockerfile.WriteString(fmt.Sprintf("        OC_ARCH=\"x64\"; OC_SHA256=\"%s\"; \\\n", opencodeChecksumAmd64))
		dockerfile.WriteString("    else \\\n")
		dockerfile.WriteString("        echo \"ERROR: Unsupported architecture: $ARCH\"; exit 1; \\\n")
		dockerfile.WriteString("    fi && \\\n")
	}
	// Shared download + verify logic (identical for Alpine and Debian)
	dockerfile.WriteString(fmt.Sprintf("    curl %s -o /tmp/opencode.tar.gz \"https://github.com/anomalyco/opencode/releases/download/v%s/opencode-linux-${OC_ARCH}-musl.tar.gz\" && \\\n", curlFlags, opencodeVersion))
	dockerfile.WriteString("    printf '%s  /tmp/opencode.tar.gz\\n' \"${OC_SHA256}\" > /tmp/opencode.sha256 && \\\n")
	dockerfile.WriteString("    sha256sum -c - < /tmp/opencode.sha256 && \\\n")
	dockerfile.WriteString("    tar -C /tmp -xzf /tmp/opencode.tar.gz opencode && \\\n")
	dockerfile.WriteString("    install /tmp/opencode /usr/local/bin/opencode && \\\n")
	dockerfile.WriteString("    rm /tmp/opencode.tar.gz /tmp/opencode /tmp/opencode.sha256 && \\\n")
	dockerfile.WriteString("    test -x /usr/local/bin/opencode\n\n")
}

// getGoToolsList returns the resolved Go tools list (config or defaults)
func (g *DefaultDockerfileGenerator) getGoToolsList() []string {
	tools := g.workspaceYAML.Build.DevStage.DevTools
	if len(tools) == 0 {
		tools = g.getDefaultLanguageTools()
	}
	return tools
}

// hasGoToolsBuilder returns true if we need a parallel Go tools builder stage.
// This is true when the language is golang and there are go-installable tools
// (i.e., tools other than golangci-lint which uses a curl installer).
func (g *DefaultDockerfileGenerator) hasGoToolsBuilder() bool {
	if g.language != "golang" {
		return false
	}
	for _, tool := range g.getGoToolsList() {
		if tool != "golangci-lint" {
			return true
		}
	}
	return false
}

// generateGoToolsBuilder creates a parallel stage to install Go dev tools
func (g *DefaultDockerfileGenerator) generateGoToolsBuilder(dockerfile *strings.Builder) {
	tools := g.getGoToolsList()

	var installCmds []string
	for _, tool := range tools {
		switch tool {
		case "gopls":
			installCmds = append(installCmds, "go install golang.org/x/tools/gopls@latest")
		case "delve":
			installCmds = append(installCmds, "go install github.com/go-delve/delve/cmd/dlv@latest")
		case "golangci-lint":
			// golangci-lint uses curl installer, not go install; handled in dev stage
		default:
			installCmds = append(installCmds, fmt.Sprintf("go install %s@latest", tool))
		}
	}

	if len(installCmds) == 0 {
		return
	}

	dockerfile.WriteString("# --- Parallel builder: Go tools ---\n")
	goToolsImage := fmt.Sprintf("golang:%s-alpine", g.goToolsBuilderVersion())
	dockerfile.WriteString(pinnedImageComment(goToolsImage))
	dockerfile.WriteString(fmt.Sprintf("FROM %s AS go-tools-builder\n", pinnedImage(goToolsImage)))
	dockerfile.WriteString("RUN --mount=type=cache,target=/go/pkg/mod \\\n")
	dockerfile.WriteString("    --mount=type=cache,target=/root/.cache/go-build \\\n")
	dockerfile.WriteString("    " + strings.Join(installCmds, " && \\\n    ") + "\n\n")
}

// emitCopyFromBuilders emits COPY --from= directives to pull binaries from parallel stages.
// Iterates the pre-computed stages slice to stay in sync with emitBuilderStages().
func (g *DefaultDockerfileGenerator) emitCopyFromBuilders(dockerfile *strings.Builder, stages []builderStage) {
	dockerfile.WriteString("# Copy binaries from parallel builder stages\n")
	for _, stage := range stages {
		for _, line := range stage.copyLines {
			dockerfile.WriteString(line + "\n")
		}
	}
	dockerfile.WriteString("\n")
}

// effectiveVersion returns the language-appropriate version to use, falling back
// to a sensible default when the user hasn't specified one. This is the single
// source of truth for version defaulting, replacing per-language inline logic.
func (g *DefaultDockerfileGenerator) effectiveVersion() string {
	if g.version != "" {
		return g.version
	}
	switch g.language {
	case "python":
		return "3.11"
	case "golang":
		return "1.22"
	case "nodejs":
		return "20"
	default:
		return ""
	}
}

// effectiveGoVersion returns the version for Go-based stages.
// Delegates to effectiveVersion() for unified version defaulting.
func (g *DefaultDockerfileGenerator) effectiveGoVersion() string {
	return g.effectiveVersion()
}

// goToolsBuilderVersion returns the Go version for the go-tools-builder stage.
// gopls@latest requires Go >= 1.24, so we enforce a minimum floor. The workspace
// base image may use an older Go version (e.g., 1.21) for application compatibility,
// but the tools builder only compiles developer tools — the binaries are copied
// into the final image via COPY --from and don't require matching Go versions.
// See issue #247.
const goToolsMinGoVersion = "1.25"

func (g *DefaultDockerfileGenerator) goToolsBuilderVersion() string {
	v := g.effectiveGoVersion()
	if compareGoVersions(v, goToolsMinGoVersion) < 0 {
		return goToolsMinGoVersion
	}
	return v
}

// compareGoVersions compares two Go version strings like "1.21" and "1.25".
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
func compareGoVersions(a, b string) int {
	aParts := strings.SplitN(a, ".", 2)
	bParts := strings.SplitN(b, ".", 2)

	aMajor, aMinor := 0, 0
	bMajor, bMinor := 0, 0

	if len(aParts) >= 1 {
		fmt.Sscanf(aParts[0], "%d", &aMajor)
	}
	if len(aParts) >= 2 {
		fmt.Sscanf(aParts[1], "%d", &aMinor)
	}
	if len(bParts) >= 1 {
		fmt.Sscanf(bParts[0], "%d", &bMajor)
	}
	if len(bParts) >= 2 {
		fmt.Sscanf(bParts[1], "%d", &bMinor)
	}

	if aMajor != bMajor {
		if aMajor < bMajor {
			return -1
		}
		return 1
	}
	if aMinor != bMinor {
		if aMinor < bMinor {
			return -1
		}
		return 1
	}
	return 0
}

func (g *DefaultDockerfileGenerator) generateDevStage(dockerfile *strings.Builder) {
	// Get packages from config or use defaults
	packages := g.workspaceYAML.Build.DevStage.Packages
	if len(packages) == 0 {
		packages = g.getDefaultPackages()
	}

	isAlpine := g.isAlpineImage()

	// Merge all packages into a single install: dev packages + nvim deps + Mason toolchains
	// This eliminates redundant apt-get update/apk update calls
	allPackages := make([]string, 0, len(packages)+10)
	allPackages = append(allPackages, packages...)

	// Add nvim dependency packages (previously in installNvimDependencies)
	if g.workspaceYAML.Nvim.Structure != "none" {
		if isAlpine {
			allPackages = appendUnique(allPackages, "unzip", "build-base", "ripgrep", "fd")
			// Mason toolchains for Alpine
			allPackages = appendUnique(allPackages, "nodejs", "npm", "py3-pip", "cargo")
			// Neovim via apk (Alpine can't use GitHub releases - musl vs glibc)
			allPackages = appendUnique(allPackages, "neovim", "neovim-doc")
		} else {
			allPackages = appendUnique(allPackages, "unzip", "build-essential", "ripgrep", "fd-find")
			// Mason toolchains for Debian (Node.js installed separately via NodeSource)
			allPackages = appendUnique(allPackages, "python3-pip", "cargo")
		}
	}

	// Install all packages in one shot with cache mounts
	dockerfile.WriteString("# Install all dev tools, nvim dependencies, and Mason toolchains (merged)\n")
	if isAlpine {
		dockerfile.WriteString(g.apkCacheMounts())
		dockerfile.WriteString("    apk add \\\n")
	} else {
		dockerfile.WriteString(g.aptCacheMounts())
		dockerfile.WriteString("    apt-get update && apt-get install -y --no-install-recommends --fix-broken \\\n")
	}

	for i, pkg := range allPackages {
		if i < len(allPackages)-1 {
			dockerfile.WriteString(fmt.Sprintf("    %s \\\n", pkg))
		} else {
			dockerfile.WriteString(fmt.Sprintf("    %s\n", pkg))
		}
	}
	dockerfile.WriteString("\n")

	// Install Node.js 22 from NodeSource for Debian when nvim is enabled.
	// Runs AFTER the merged apt-get install so that curl is available.
	// Falls back to Debian's default nodejs+npm if NodeSource is unreachable.
	if g.workspaceYAML.Nvim.Structure != "none" && !g.isAlpineImage() {
		dockerfile.WriteString("# Install Node.js 22 from NodeSource (Mason toolchains require Node 22+)\n")
		dockerfile.WriteString("# Falls back to Debian default nodejs+npm if NodeSource is unreachable\n")
		dockerfile.WriteString(fmt.Sprintf("RUN (curl %s https://deb.nodesource.com/setup_22.x | bash - \\\n", curlFlags))
		dockerfile.WriteString("    && apt-get install -y --no-install-recommends nodejs) \\\n")
		dockerfile.WriteString("    || apt-get install -y --no-install-recommends nodejs npm\n\n")
	}

	// npm install neovim (needed by Mason) - with cache mount
	// Falls back to direct npmjs.org access if HTTP proxy/registry is unreachable
	if g.workspaceYAML.Nvim.Structure != "none" {
		dockerfile.WriteString("# Install neovim npm package for Mason\n")
		dockerfile.WriteString("# Falls back to direct npmjs.org access if HTTP proxy/registry is unreachable\n")
		dockerfile.WriteString("RUN --mount=type=cache,target=/root/.npm \\\n")
		dockerfile.WriteString("    npm install -g neovim || \\\n")
		dockerfile.WriteString("    (unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy NPM_CONFIG_REGISTRY npm_config_registry \\\n")
		dockerfile.WriteString("    && npm install -g neovim)\n\n")
	}

	// Install language-specific tools (non-Go, since Go tools come from parallel builder)
	languageTools := g.workspaceYAML.Build.DevStage.DevTools
	if len(languageTools) == 0 {
		languageTools = g.getDefaultLanguageTools()
	}
	if len(languageTools) > 0 && g.language != "golang" {
		g.installLanguageTools(dockerfile, languageTools)
	}

	// For golang, install golangci-lint separately via direct binary download + checksum.
	// Pinned to a specific version with SHA256 verification (see checksums.go).
	if g.language == "golang" {
		for _, tool := range languageTools {
			if tool == "golangci-lint" {
				dockerfile.WriteString("# Install golangci-lint via direct binary download with checksum verification\n")
				dockerfile.WriteString("RUN set -e && \\\n")
				dockerfile.WriteString("    ARCH=$(dpkg --print-architecture 2>/dev/null || (uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')) && \\\n")
				dockerfile.WriteString("    if [ \"$ARCH\" = \"arm64\" ] || [ \"$ARCH\" = \"aarch64\" ]; then \\\n")
				dockerfile.WriteString(fmt.Sprintf("        GCI_ARCH=\"arm64\"; GCI_SHA256=\"%s\"; \\\n", golangciLintChecksumArm64))
				dockerfile.WriteString("    elif [ \"$ARCH\" = \"amd64\" ] || [ \"$ARCH\" = \"x86_64\" ]; then \\\n")
				dockerfile.WriteString(fmt.Sprintf("        GCI_ARCH=\"amd64\"; GCI_SHA256=\"%s\"; \\\n", golangciLintChecksumAmd64))
				dockerfile.WriteString("    else \\\n")
				dockerfile.WriteString("        echo \"ERROR: Unsupported architecture: $ARCH\"; exit 1; \\\n")
				dockerfile.WriteString("    fi && \\\n")
				dockerfile.WriteString(fmt.Sprintf("    curl %s -o /tmp/golangci-lint.tar.gz \"https://github.com/golangci/golangci-lint/releases/download/v%s/golangci-lint-%s-linux-${GCI_ARCH}.tar.gz\" && \\\n", curlFlags, golangciLintVersion, golangciLintVersion))
				dockerfile.WriteString("    echo \"${GCI_SHA256}  /tmp/golangci-lint.tar.gz\" | sha256sum -c - && \\\n")
				dockerfile.WriteString(fmt.Sprintf("    tar -C /tmp -xzf /tmp/golangci-lint.tar.gz golangci-lint-%s-linux-${GCI_ARCH}/golangci-lint && \\\n", golangciLintVersion))
				dockerfile.WriteString(fmt.Sprintf("    install /tmp/golangci-lint-%s-linux-${GCI_ARCH}/golangci-lint $(go env GOPATH)/bin/ && \\\n", golangciLintVersion))
				dockerfile.WriteString("    rm -rf /tmp/golangci-lint.tar.gz /tmp/golangci-lint-*\n\n")
				break
			}
		}
	}

	// Custom commands
	for _, cmd := range g.workspaceYAML.Build.DevStage.CustomCommands {
		dockerfile.WriteString(fmt.Sprintf("RUN %s\n", cmd))
	}

	if len(g.workspaceYAML.Build.DevStage.CustomCommands) > 0 {
		dockerfile.WriteString("\n")
	}
}

// appendUnique appends items to a slice only if they're not already present
func appendUnique(slice []string, items ...string) []string {
	existing := make(map[string]bool, len(slice))
	for _, s := range slice {
		existing[s] = true
	}
	for _, item := range items {
		if !existing[item] {
			slice = append(slice, item)
			existing[item] = true
		}
	}
	return slice
}

// formatSystemDepSources creates a human-readable string of auto-detected system deps
// for Dockerfile comments. Example: "psycopg2 -> libpq-dev, lxml -> libxml2-dev libxslt1-dev"
func formatSystemDepSources(info *utils.PrivateRepoInfo) string {
	if info == nil || len(info.SystemDepSources) == 0 {
		return ""
	}

	// Group deps by their source Python package
	sourceToDepsList := make(map[string][]string)
	for dep, source := range info.SystemDepSources {
		sourceToDepsList[source] = append(sourceToDepsList[source], dep)
	}

	// Sort deps within each source for deterministic output
	for source := range sourceToDepsList {
		sort.Strings(sourceToDepsList[source])
	}

	// Build formatted string
	var parts []string
	for source, deps := range sourceToDepsList {
		parts = append(parts, fmt.Sprintf("%s -> %s", source, strings.Join(deps, " ")))
	}

	// Sort for deterministic output
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

func (g *DefaultDockerfileGenerator) generateDevUser(dockerfile *strings.Builder) {
	uid := g.workspaceYAML.Container.UID
	gid := g.workspaceYAML.Container.GID
	user := g.workspaceYAML.Container.User

	if uid == 0 {
		uid = 1000
	}
	if gid == 0 {
		gid = 1000
	}
	if user == "" {
		user = "dev"
	}

	// Detect if this is an Alpine-based image
	isAlpine := g.isAlpineImage()

	if isAlpine {
		// Alpine: use addgroup/adduser (BusyBox)
		dockerfile.WriteString("# Create dev user (Alpine/BusyBox)\n")
		dockerfile.WriteString(fmt.Sprintf("RUN addgroup -g %d %s 2>/dev/null || true\n", gid, user))
		dockerfile.WriteString(fmt.Sprintf("RUN adduser -D -u %d -G %s -s /bin/zsh %s 2>/dev/null || true\n\n", uid, user, user))
	} else {
		// Debian/Ubuntu: use groupadd/useradd
		dockerfile.WriteString("# Create dev user (Debian/Ubuntu)\n")
		dockerfile.WriteString(fmt.Sprintf("RUN groupadd -g %d %s 2>/dev/null || true\n", gid, user))
		dockerfile.WriteString(fmt.Sprintf("RUN useradd -m -u %d -g %s -s /bin/zsh %s 2>/dev/null || true\n\n", uid, user, user))
	}

	// Copy shell configuration files from staging area (only if they exist)
	stagingDir := g.effectiveStagingDir()
	hasZshrc := stagingDir != "" && fileExistsInDir(stagingDir, ".zshrc")
	hasStarship := stagingDir != "" && fileExistsInDir(stagingDir, filepath.Join(".config", "starship.toml"))

	if hasZshrc || hasStarship {
		dockerfile.WriteString("# Copy shell configuration\n")
		var chownPaths []string
		if hasZshrc {
			dockerfile.WriteString(fmt.Sprintf("COPY .zshrc /home/%s/.zshrc\n", user))
			chownPaths = append(chownPaths, fmt.Sprintf("/home/%s/.zshrc", user))
		}
		if hasStarship {
			dockerfile.WriteString(fmt.Sprintf("COPY .config/starship.toml /home/%s/.config/starship.toml\n", user))
			chownPaths = append(chownPaths, fmt.Sprintf("/home/%s/.config", user))
		}
		dockerfile.WriteString(fmt.Sprintf("RUN chown -R %s:%s %s\n\n", user, user, strings.Join(chownPaths, " ")))
	} else {
		dockerfile.WriteString("# Shell configuration files not found in staging — skipped\n\n")
	}
}

func (g *DefaultDockerfileGenerator) getDefaultPackages() []string {
	base := []string{
		"git",
		"curl",
		"wget",
		"zsh",
		"openssh-client",
	}

	// Add language-specific base packages
	switch g.language {
	case "python":
		if !g.isAlpineImage() {
			base = append(base, "build-essential", "python3-pip")
		}
	case "golang":
		// Go images usually have everything needed
	}

	return base
}

// isAlpineImage returns true if the base image is Alpine-based.
// This value is computed in generateBaseStage() based on the actual image chosen,
// not guessed from workspace metadata. It affects user creation commands,
// package manager selection, and binary compatibility.
func (g *DefaultDockerfileGenerator) isAlpineImage() bool {
	return g.isAlpine
}

// cacheID returns a workspace-scoped identifier for Docker cache mounts.
// When multiple workspaces build in parallel, shared cache mounts cause apt lock
// conflicts (E: Could not get lock /var/lib/apt/lists/lock). Scoping the cache
// mount ID to the workspace name ensures each parallel build uses its own cache,
// eliminating lock file contention. See issue #233.
func (g *DefaultDockerfileGenerator) cacheID() string {
	if g.workspace != nil && g.workspace.Name != "" {
		return g.workspace.Name
	}
	return "default"
}

// aptCacheMounts returns multi-line apt cache mount directives with workspace-scoped IDs
// and sharing=locked to prevent cache corruption during parallel builds. See issue #249.
// Format (for embedding into RUN instructions):
//
//	RUN --mount=type=cache,target=/var/cache/apt,id=apt-cache-<ws>,sharing=locked \
//	    --mount=type=cache,target=/var/lib/apt/lists,id=apt-lists-<ws>,sharing=locked \
func (g *DefaultDockerfileGenerator) aptCacheMounts() string {
	id := g.cacheID()
	return fmt.Sprintf("RUN --mount=type=cache,target=/var/cache/apt,id=apt-cache-%s,sharing=locked \\\n"+
		"    --mount=type=cache,target=/var/lib/apt/lists,id=apt-lists-%s,sharing=locked \\\n", id, id)
}

// aptCacheMountsLocked returns single-line apt cache mount directives with
// workspace-scoped IDs and sharing=locked (used in parallel builder stages).
func (g *DefaultDockerfileGenerator) aptCacheMountsLocked() string {
	id := g.cacheID()
	return fmt.Sprintf("RUN --mount=type=cache,target=/var/cache/apt,id=apt-cache-%s,sharing=locked "+
		"--mount=type=cache,target=/var/lib/apt,id=apt-lists-%s,sharing=locked \\\n", id, id)
}

// apkCacheMounts returns the apk cache mount directive with a workspace-scoped ID
// and sharing=locked to prevent cache corruption during parallel builds. See issue #249.
func (g *DefaultDockerfileGenerator) apkCacheMounts() string {
	return fmt.Sprintf("RUN --mount=type=cache,target=/var/cache/apk,id=apk-cache-%s,sharing=locked \\\n", g.cacheID())
}

// apkCacheMountsLocked returns the apk cache mount directive with a workspace-scoped
// ID and sharing=locked (used in parallel builder stages).
func (g *DefaultDockerfileGenerator) apkCacheMountsLocked() string {
	return fmt.Sprintf("RUN --mount=type=cache,target=/var/cache/apk,id=apk-cache-%s,sharing=locked \\\n", g.cacheID())
}

// effectiveStagingDir returns the staging directory to use for file existence checks.
// Prefers the explicitly-set stagingDir; falls back to the legacy PathConfig-based lookup
// using filepath.Base(appPath) for backward compatibility.
func (g *DefaultDockerfileGenerator) effectiveStagingDir() string {
	if g.stagingDir != "" {
		return g.stagingDir
	}
	// Legacy fallback: compute from pathConfig + appPath basename.
	// Only return the computed path if it actually exists on disk —
	// in unit tests the staging directory is never created, so we
	// must fall back to "" (which lets callers use appPath instead).
	pc := g.pathConfig
	if pc == nil {
		return ""
	}
	dir := pc.BuildStagingDir(filepath.Base(g.appPath))
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return ""
	}
	return dir
}

// fileExistsInDir checks whether a relative path exists within the given directory.
func fileExistsInDir(dir, relPath string) bool {
	_, err := os.Stat(filepath.Join(dir, relPath))
	return err == nil
}

func (g *DefaultDockerfileGenerator) getDefaultLanguageTools() []string {
	switch g.language {
	case "python":
		return []string{"python-lsp-server", "black", "isort", "pytest"}
	case "golang":
		return []string{"gopls", "delve", "golangci-lint"}
	case "nodejs":
		return []string{"typescript", "ts-node"}
	default:
		return []string{}
	}
}

func (g *DefaultDockerfileGenerator) installLanguageTools(dockerfile *strings.Builder, tools []string) {
	dockerfile.WriteString("# Install language-specific tools\n")

	switch g.language {
	case "python":
		toolsList := strings.Join(tools, " ")
		dockerfile.WriteString("RUN --mount=type=cache,target=/root/.cache/pip \\\n")
		dockerfile.WriteString(fmt.Sprintf("    pip install %s \\\n", toolsList))
		dockerfile.WriteString("    || (unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy \\\n")
		dockerfile.WriteString(fmt.Sprintf("    && pip install %s)\n\n", toolsList))

	case "golang":
		// Go tools are installed in the parallel go-tools-builder stage
		// Only golangci-lint (curl installer) is handled separately in generateDevStage
		dockerfile.WriteString("# Go tools installed via parallel builder stage\n\n")

	case "nodejs":
		toolsList := strings.Join(tools, " ")
		dockerfile.WriteString("RUN --mount=type=cache,target=/root/.npm \\\n")
		dockerfile.WriteString(fmt.Sprintf("    npm install -g %s || \\\n", toolsList))
		dockerfile.WriteString("    (unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy NPM_CONFIG_REGISTRY npm_config_registry \\\n")
		dockerfile.WriteString(fmt.Sprintf("    && npm install -g %s)\n\n", toolsList))
	}
}

// effectiveUser returns the configured container user, defaulting to "dev"
func (g *DefaultDockerfileGenerator) effectiveUser() string {
	if g.workspaceYAML.Container.User != "" {
		return g.workspaceYAML.Container.User
	}
	return "dev"
}

// collectAdditionalBuildArgs returns a de-duplicated, sorted list of additional ARG names
// from both AdditionalBuildArgs and WorkspaceSpec.Build.Args, excluding any already in requiredBuildArgs.
func (g *DefaultDockerfileGenerator) collectAdditionalBuildArgs(requiredBuildArgs []string) []string {
	skip := make(map[string]bool, len(requiredBuildArgs))
	for _, arg := range requiredBuildArgs {
		skip[arg] = true
	}

	seen := make(map[string]bool)
	var result []string

	// Collect from AdditionalBuildArgs
	for _, arg := range g.additionalBuildArgs {
		if !skip[arg] && !seen[arg] {
			seen[arg] = true
			result = append(result, arg)
		}
	}

	// Collect from WorkspaceSpec.Build.Args keys
	for key := range g.workspaceYAML.Build.Args {
		if !skip[key] && !seen[key] {
			seen[key] = true
			result = append(result, key)
		}
	}

	sort.Strings(result)
	return result
}

// emitAdditionalBuildArgs writes ARG declarations for additional build args to the Dockerfile.
// These are ARG-only (no ENV) to avoid leaking secrets into the final image.
func (g *DefaultDockerfileGenerator) emitAdditionalBuildArgs(dockerfile *strings.Builder, requiredBuildArgs []string) {
	args := g.collectAdditionalBuildArgs(requiredBuildArgs)
	if len(args) == 0 {
		return
	}

	dockerfile.WriteString("# Additional build arguments\n")
	for _, arg := range args {
		dockerfile.WriteString(fmt.Sprintf("ARG %s\n", arg))
	}
	dockerfile.WriteString("\n")
}

// emitCACertSection writes the CA certificate injection block to the Dockerfile.
// This installs corporate/custom CA certificates so tools (pip, curl, node) trust them.
// Only emits the COPY command if the certs directory actually exists in the staging directory.
func (g *DefaultDockerfileGenerator) emitCACertSection(dockerfile *strings.Builder) {
	if len(g.workspaceYAML.Build.CACerts) == 0 {
		return
	}

	// Verify certs/ exists in staging before emitting COPY (fixes #228)
	stagingDir := g.effectiveStagingDir()
	if stagingDir == "" || !fileExistsInDir(stagingDir, "certs") {
		dockerfile.WriteString("# CA certificates configured but certs/ not found in staging — skipped\n")
		dockerfile.WriteString("# Ensure vault is accessible and CA certs are resolved before build\n\n")
		return
	}

	dockerfile.WriteString("# Corporate CA certificates\n")
	dockerfile.WriteString("COPY certs/ /usr/local/share/ca-certificates/custom/\n")
	dockerfile.WriteString("RUN update-ca-certificates\n")
	dockerfile.WriteString("ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt\n")
	dockerfile.WriteString("ENV REQUESTS_CA_BUNDLE=/etc/ssl/certs/ca-certificates.crt\n")
	dockerfile.WriteString("ENV NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt\n\n")
}

// generateNvimSection copies nvim config and installs plugins (called after user creation).
// Returns an error if filesystem operations fail unexpectedly (NOT for missing nvim config,
// which is a normal condition handled gracefully with a skip comment).
func (g *DefaultDockerfileGenerator) generateNvimSection(dockerfile *strings.Builder) error {
	// Skip if explicitly disabled
	if g.workspaceYAML.Nvim.Structure == "none" {
		return nil
	}

	// Check if staging nvim config directory exists.
	// Use the explicit staging directory when available (fixes path mismatch
	// when appPath basename differs from the staging key — see issue #228).
	stagingDir := g.effectiveStagingDir()
	if stagingDir == "" {
		// No staging directory available — nvim config cannot be found.
		// This is normal for fresh workspaces or unit tests.
		dockerfile.WriteString("# Skipping Neovim configuration (no config generated)\n")
		dockerfile.WriteString("# Run 'dvm build' after setting up nvim plugins to enable nvim config\n\n")
		return nil
	}
	nvimConfigPath := filepath.Join(stagingDir, ".config", "nvim")
	if _, err := os.Stat(nvimConfigPath); err != nil {
		if os.IsNotExist(err) {
			// Nvim config doesn't exist - skip this section (normal for fresh workspace)
			dockerfile.WriteString("# Skipping Neovim configuration (no config generated)\n")
			dockerfile.WriteString("# Run 'dvm build' after setting up nvim plugins to enable nvim config\n\n")
			return nil
		}
		// Unexpected filesystem error (permission denied, etc.)
		return fmt.Errorf("checking nvim config at %s: %w", nvimConfigPath, err)
	}

	// Copy nvim configuration from staging directory
	user := g.effectiveUser()
	dockerfile.WriteString("# Copy Neovim configuration\n")
	dockerfile.WriteString(fmt.Sprintf("COPY .config/nvim /home/%s/.config/nvim\n", user))
	dockerfile.WriteString(fmt.Sprintf("RUN chown -R %s:%s /home/%s/.config\n\n", user, user, user))

	// tree-sitter CLI is already installed via parallel builder stage (COPY --from=treesitter-builder)

	// Install lazy.nvim and plugins as dev user with proper error handling
	dockerfile.WriteString(fmt.Sprintf("USER %s\n", user))
	dockerfile.WriteString("# Bootstrap lazy.nvim and install plugins\n")
	dockerfile.WriteString("RUN unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy NPM_CONFIG_REGISTRY npm_config_registry && \\\n")
	dockerfile.WriteString("    nvim --headless \"+Lazy! sync\" +qa 2>&1 | tee /tmp/nvim-install.log || \\\n")
	dockerfile.WriteString("    (cat /tmp/nvim-install.log && exit 1)\n\n")

	// Install Treesitter parsers at build time (reduces first-attach startup time)
	g.installTreesitterParsers(dockerfile)

	// Install Mason tools (LSPs, linters, formatters) at build time (reduces first-attach startup time)
	g.installMasonTools(dockerfile)

	dockerfile.WriteString("USER root\n\n")
	return nil
}

// getMasonToolsForLanguage returns Mason packages (LSPs, linters, formatters) for the detected language.
// This is the SINGLE AUTHORITY for language-specific Mason tool installation.
// The plugin YAML (06-mason.yaml) provides only framework setup, not tool lists.
func (g *DefaultDockerfileGenerator) getMasonToolsForLanguage() []string {
	switch g.language {
	case "python":
		return []string{"pyright", "ruff-lsp", "black", "isort", "pylint"}
	case "golang":
		return []string{"gopls", "golangci-lint-langserver", "goimports"}
	case "nodejs":
		return []string{"typescript-language-server", "eslint-lsp", "prettier"}
	case "rust":
		return []string{"rust-analyzer"}
	case "ruby":
		return []string{"solargraph", "rubocop"}
	case "java":
		return []string{"jdtls", "google-java-format"}
	case "gleam":
		return []string{"gleam"} // Gleam LSP is installed via Mason
	default:
		return []string{}
	}
}

// getBaseMasonTools returns Mason packages that should be installed in every workspace,
// regardless of the detected language. These support nvim config editing and shell scripts.
func (g *DefaultDockerfileGenerator) getBaseMasonTools() []string {
	return []string{"lua-language-server", "stylua"}
}

// installMasonTools installs language servers, linters, and formatters via Mason at build time.
// Uses synchronous Lua-based install with mason-registry and vim.wait() to ensure
// all tools are fully installed before the build layer completes.
//
// The Lua script force-loads mason.nvim via lazy.nvim to ensure Mason is available
// in headless mode (see issue #234). It includes per-tool logging, retry logic for
// transient network failures, and a verification step.
func (g *DefaultDockerfileGenerator) installMasonTools(dockerfile *strings.Builder) {
	// Check if Mason is installed via manifest
	if g.pluginManifest != nil && !g.pluginManifest.Features.HasMason {
		dockerfile.WriteString("# Mason not installed - skipping LSP pre-install\n\n")
		return
	}

	// Merge base tools (always installed) with language-specific tools
	tools := g.getBaseMasonTools()
	tools = append(tools, g.getMasonToolsForLanguage()...)

	// Append user-configured extra tools
	if g.workspaceYAML.Nvim.ExtraMasonTools != nil {
		tools = append(tools, g.workspaceYAML.Nvim.ExtraMasonTools...)
	}

	if len(tools) == 0 {
		return
	}

	// Build Lua tool list string: 'tool1','tool2',...
	luaTools := "'" + strings.Join(tools, "','") + "'"

	dockerfile.WriteString("# Install LSPs, linters, and formatters via Mason at build time\n")
	// Write the Lua install script using BuildKit COPY heredoc syntax.
	// The generated Dockerfile has `# syntax=docker/dockerfile:1` which enables
	// heredoc support. Using COPY <<EOF avoids the broken pattern of embedding
	// shell heredocs inside RUN continuations (which caused Docker to see Lua
	// code lines as unknown Dockerfile instructions — see issue #204).
	// Use --chown so the file is owned by the container user, allowing cleanup
	// in subsequent RUN steps that execute as non-root (see issue #222).
	user := g.effectiveUser()
	g.writeMasonLuaScript(dockerfile, user, luaTools)

	// Execute nvim with the Lua script.
	// Force-load mason.nvim via Lazy! so Mason is available in headless mode
	// (same pattern as treesitter fix — see issues #232, #234).
	// Proxy vars are unset to avoid interference with Mason's package downloads.
	// Note: we don't rm the temp file — it's in /tmp and gets cleaned up anyway,
	// and removing it caused permission errors when COPY ran as root (issue #222).
	dockerfile.WriteString("RUN unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy NPM_CONFIG_REGISTRY npm_config_registry && \\\n")
	dockerfile.WriteString("    nvim --headless \\\n")
	dockerfile.WriteString("      -c \"Lazy! load mason.nvim\" \\\n")
	dockerfile.WriteString("      +\"luafile /tmp/mason-install.lua\" +qa 2>&1 | tee /tmp/mason-install.log || \\\n")
	dockerfile.WriteString("    (echo '--- Mason install log ---' && cat /tmp/mason-install.log && exit 1)\n\n")

	// Verification: ensure Mason packages directory was populated
	g.writeMasonVerification(dockerfile, user, len(tools))
}

// writeMasonLuaScript writes the COPY heredoc for the Mason install Lua script.
// The script includes: error-handled registry refresh, per-tool logging with retry
// logic (3 attempts per tool), and a final verification that all tools installed.
func (g *DefaultDockerfileGenerator) writeMasonLuaScript(dockerfile *strings.Builder, user, luaTools string) {
	dockerfile.WriteString(fmt.Sprintf("COPY --chown=%s:%s <<'LUAEOF' /tmp/mason-install.lua\n", user, user))
	// Refresh registry with error handling
	dockerfile.WriteString("local registry = require('mason-registry')\n")
	dockerfile.WriteString("print('[Mason] Refreshing registry...')\n")
	dockerfile.WriteString("local refresh_ok, refresh_err = pcall(registry.refresh)\n")
	dockerfile.WriteString("if not refresh_ok then\n")
	dockerfile.WriteString("  print('[Mason] ERROR: Registry refresh failed: ' .. tostring(refresh_err))\n")
	dockerfile.WriteString("  vim.cmd('cq')\n")
	dockerfile.WriteString("end\n\n")
	// Tool list and install loop with retry logic
	dockerfile.WriteString(fmt.Sprintf("local tools = {%s}\n", luaTools))
	dockerfile.WriteString("local max_retries = 3\n")
	dockerfile.WriteString("local failed = {}\n\n")
	dockerfile.WriteString("for _, name in ipairs(tools) do\n")
	dockerfile.WriteString("  local ok, pkg = pcall(registry.get_package, name)\n")
	dockerfile.WriteString("  if not ok then\n")
	dockerfile.WriteString("    print('[Mason] ERROR: Package not found: ' .. name)\n")
	dockerfile.WriteString("    table.insert(failed, name)\n")
	dockerfile.WriteString("  elseif pkg:is_installed() then\n")
	dockerfile.WriteString("    print('[Mason] ' .. name .. ' already installed')\n")
	dockerfile.WriteString("  else\n")
	dockerfile.WriteString("    local installed = false\n")
	dockerfile.WriteString("    for attempt = 1, max_retries do\n")
	dockerfile.WriteString("      print('[Mason] Installing ' .. name .. ' (attempt ' .. attempt .. '/' .. max_retries .. ')')\n")
	dockerfile.WriteString("      local install_ok, install_err = pcall(function()\n")
	dockerfile.WriteString("        pkg:install()\n")
	dockerfile.WriteString("      end)\n")
	dockerfile.WriteString("      if not install_ok then\n")
	dockerfile.WriteString("        print('[Mason] ERROR starting ' .. name .. ': ' .. tostring(install_err))\n")
	dockerfile.WriteString("      end\n")
	// Wait for this specific tool with a per-tool timeout
	dockerfile.WriteString("      vim.wait(120000, function() return pkg:is_installed() end, 2000)\n")
	dockerfile.WriteString("      if pkg:is_installed() then\n")
	dockerfile.WriteString("        print('[Mason] OK: ' .. name .. ' installed')\n")
	dockerfile.WriteString("        installed = true\n")
	dockerfile.WriteString("        break\n")
	dockerfile.WriteString("      else\n")
	dockerfile.WriteString("        print('[Mason] WARN: ' .. name .. ' not installed after attempt ' .. attempt)\n")
	dockerfile.WriteString("      end\n")
	dockerfile.WriteString("    end\n")
	dockerfile.WriteString("    if not installed then\n")
	dockerfile.WriteString("      print('[Mason] FAILED: ' .. name .. ' after ' .. max_retries .. ' attempts')\n")
	dockerfile.WriteString("      table.insert(failed, name)\n")
	dockerfile.WriteString("    end\n")
	dockerfile.WriteString("  end\n")
	dockerfile.WriteString("end\n\n")
	// Final verification
	dockerfile.WriteString("-- Final verification\n")
	dockerfile.WriteString("local done = 0\n")
	dockerfile.WriteString("for _, name in ipairs(tools) do\n")
	dockerfile.WriteString("  local ok, pkg = pcall(registry.get_package, name)\n")
	dockerfile.WriteString("  if ok and pkg:is_installed() then done = done + 1 end\n")
	dockerfile.WriteString("end\n")
	dockerfile.WriteString("print('[Mason] Installed ' .. done .. '/' .. #tools .. ' tools')\n")
	dockerfile.WriteString("if #failed > 0 then\n")
	dockerfile.WriteString("  print('[Mason] FAILED tools: ' .. table.concat(failed, ', '))\n")
	dockerfile.WriteString("  vim.cmd('cq')\n")
	dockerfile.WriteString("end\n")
	dockerfile.WriteString("LUAEOF\n\n")
}

// writeMasonVerification writes a Dockerfile RUN step that verifies Mason packages
// were actually installed by checking the Mason packages directory.
func (g *DefaultDockerfileGenerator) writeMasonVerification(dockerfile *strings.Builder, user string, toolCount int) {
	dockerfile.WriteString("# Verify Mason tools were installed\n")
	dockerfile.WriteString(fmt.Sprintf("RUN pkg_count=$(ls -1d /home/%s/.local/share/nvim/mason/packages/*/ 2>/dev/null | wc -l) && \\\n", user))
	dockerfile.WriteString("    if [ \"$pkg_count\" -lt 1 ]; then \\\n")
	dockerfile.WriteString("      echo \"ERROR: No Mason packages found after install step\" && \\\n")
	dockerfile.WriteString("      cat /tmp/mason-install.log 2>/dev/null && \\\n")
	dockerfile.WriteString("      exit 1; \\\n")
	dockerfile.WriteString("    fi && \\\n")
	dockerfile.WriteString(fmt.Sprintf("    echo \"Mason: $pkg_count package(s) installed (expected %d)\"\n\n", toolCount))
}

// getTreesitterParsersForLanguage returns Treesitter parsers for the detected language
func (g *DefaultDockerfileGenerator) getTreesitterParsersForLanguage() []string {
	// Base parsers always included for a good editing experience
	base := []string{"lua", "vim", "vimdoc", "query", "markdown", "markdown_inline", "bash", "json", "yaml"}

	switch g.language {
	case "python":
		return append(base, "python", "toml", "dockerfile", "gitignore")
	case "golang":
		return append(base, "go", "gomod", "gosum", "gowork", "dockerfile", "gitignore")
	case "nodejs":
		return append(base, "javascript", "typescript", "tsx", "html", "css", "dockerfile", "gitignore")
	case "rust":
		return append(base, "rust", "toml", "dockerfile", "gitignore")
	case "ruby":
		return append(base, "ruby", "dockerfile", "gitignore")
	case "java":
		return append(base, "java", "xml", "dockerfile", "gitignore")
	case "gleam":
		return append(base, "gleam", "erlang", "elixir", "toml", "dockerfile", "gitignore")
	default:
		return base
	}
}

// installTreesitterParsers installs Treesitter parsers at build time
func (g *DefaultDockerfileGenerator) installTreesitterParsers(dockerfile *strings.Builder) {
	// Check if Treesitter is installed via manifest
	if g.pluginManifest != nil && !g.pluginManifest.Features.HasTreesitter {
		dockerfile.WriteString("# Treesitter not installed - skipping parser pre-install\n\n")
		return
	}

	parsers := g.getTreesitterParsersForLanguage()

	// Append user-configured extra parsers
	if g.workspaceYAML.Nvim.ExtraTreesitterParsers != nil {
		parsers = append(parsers, g.workspaceYAML.Nvim.ExtraTreesitterParsers...)
	}

	if len(parsers) == 0 {
		return
	}

	user := g.effectiveUser()

	// Write the Lua install script using BuildKit COPY heredoc syntax.
	// The old approach used -c "TSInstall ..." which is ASYNCHRONOUS — parsers
	// start compiling in the background but -c "qa" exits immediately (0.3s),
	// producing parser_count=0 (see issue #248).
	//
	// The Lua script calls TSInstall! per parser, then uses vim.wait() to poll
	// for each .so file to actually appear on disk before exiting. This ensures
	// the event loop runs long enough for compilation to complete in headless mode.
	// Note: +qa is NOT on the nvim command line — the script calls vim.cmd('qa!')
	// only after all parsers are verified compiled.
	g.writeTreesitterLuaScript(dockerfile, user, parsers)

	// Execute nvim with the Lua script.
	// Force-load nvim-treesitter via Lazy! so TSInstall! is available in headless mode
	// (see issues #232, #235, #248).
	// Proxy vars are unset to avoid interference with parser git clones.
	// NOTE: No +qa here — the Lua script exits nvim after verifying .so files exist.
	dockerfile.WriteString("# Install Treesitter parsers at build time\n")
	dockerfile.WriteString("RUN unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy NPM_CONFIG_REGISTRY npm_config_registry && \\\n")
	dockerfile.WriteString("    nvim --headless \\\n")
	dockerfile.WriteString("      -c \"Lazy! load nvim-treesitter\" \\\n")
	dockerfile.WriteString("      +\"luafile /tmp/treesitter-install.lua\" 2>&1 | tee /tmp/treesitter-install.log || \\\n")
	dockerfile.WriteString("    (cat /tmp/treesitter-install.log && exit 1)\n\n")

	// Verification: ensure at least one parser .so was actually installed
	dockerfile.WriteString("# Verify Treesitter parsers were installed\n")
	dockerfile.WriteString(fmt.Sprintf("RUN parser_count=$(find /home/%s/.local/share/nvim/lazy/nvim-treesitter/parser -name '*.so' 2>/dev/null | wc -l) && \\\n", user))
	dockerfile.WriteString("    if [ \"$parser_count\" -lt 1 ]; then \\\n")
	dockerfile.WriteString("      echo \"ERROR: No Treesitter parsers found after install step\" && \\\n")
	dockerfile.WriteString("      cat /tmp/treesitter-install.log 2>/dev/null && \\\n")
	dockerfile.WriteString("      exit 1; \\\n")
	dockerfile.WriteString("    fi && \\\n")
	dockerfile.WriteString(fmt.Sprintf("    echo \"Treesitter: $parser_count parser(s) installed (expected %d)\"\n\n", len(parsers)))
}

// writeTreesitterLuaScript writes the COPY heredoc for the Treesitter install Lua script.
// The script calls TSInstall! per parser, then polls for each .so file using vim.wait().
// This is critical in headless mode: TSInstall! dispatches async compilation jobs, and
// without polling, nvim exits before the .so files are produced (issue #248).
// The script calls vim.cmd('qa!') only after all .so files are verified on disk.
func (g *DefaultDockerfileGenerator) writeTreesitterLuaScript(dockerfile *strings.Builder, user string, parsers []string) {
	luaParsers := "'" + strings.Join(parsers, "','") + "'"

	dockerfile.WriteString(fmt.Sprintf("COPY --chown=%s:%s <<'LUAEOF' /tmp/treesitter-install.lua\n", user, user))
	dockerfile.WriteString(fmt.Sprintf("local parsers = {%s}\n", luaParsers))
	dockerfile.WriteString("local failed = {}\n")
	dockerfile.WriteString("local parser_dir = vim.fn.stdpath('data') .. '/lazy/nvim-treesitter/parser/'\n\n")
	dockerfile.WriteString("-- Dispatch all TSInstall! commands first\n")
	dockerfile.WriteString("for _, lang in ipairs(parsers) do\n")
	dockerfile.WriteString("  print('[Treesitter] Installing ' .. lang .. '...')\n")
	dockerfile.WriteString("  local ok, err = pcall(vim.cmd, 'TSInstall! ' .. lang)\n")
	dockerfile.WriteString("  if not ok then\n")
	dockerfile.WriteString("    print('[Treesitter] FAILED to dispatch: ' .. lang .. ' — ' .. tostring(err))\n")
	dockerfile.WriteString("    table.insert(failed, lang)\n")
	dockerfile.WriteString("  end\n")
	dockerfile.WriteString("end\n\n")
	dockerfile.WriteString("-- Poll for each .so file to verify compilation completed.\n")
	dockerfile.WriteString("-- TSInstall! dispatches async compilation; in headless mode the event loop\n")
	dockerfile.WriteString("-- must keep running until the .so files actually appear on disk.\n")
	dockerfile.WriteString("for _, lang in ipairs(parsers) do\n")
	dockerfile.WriteString("  local so_path = parser_dir .. lang .. '.so'\n")
	dockerfile.WriteString("  local found = vim.wait(120000, function()\n")
	dockerfile.WriteString("    return vim.fn.filereadable(so_path) == 1\n")
	dockerfile.WriteString("  end, 1000)\n")
	dockerfile.WriteString("  if found then\n")
	dockerfile.WriteString("    print('[Treesitter] OK: ' .. lang .. ' (.so verified)')\n")
	dockerfile.WriteString("  else\n")
	dockerfile.WriteString("    print('[Treesitter] TIMEOUT: ' .. lang .. ' — .so not found after 120s')\n")
	dockerfile.WriteString("    table.insert(failed, lang)\n")
	dockerfile.WriteString("  end\n")
	dockerfile.WriteString("end\n\n")
	dockerfile.WriteString("print('[Treesitter] Installed ' .. (#parsers - #failed) .. '/' .. #parsers .. ' parsers')\n")
	dockerfile.WriteString("if #failed > 0 then\n")
	dockerfile.WriteString("  print('[Treesitter] FAILED parsers: ' .. table.concat(failed, ', '))\n")
	dockerfile.WriteString("  vim.cmd('cq')\n")
	dockerfile.WriteString("else\n")
	dockerfile.WriteString("  vim.cmd('qa!')\n")
	dockerfile.WriteString("end\n")
	dockerfile.WriteString("LUAEOF\n\n")
}
