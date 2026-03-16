# Changelog

All notable changes to DevOpsMaestro are documented in the [CHANGELOG.md](https://github.com/rmkohlman/devopsmaestro/blob/main/CHANGELOG.md) file in the repository.

## Latest Releases

### v0.43.1 (2026-03-16)

**🐛 Fix Tree-Sitter Builder for Debian-Based Builds**

The tree-sitter CLI builder stage was hardcoded to `alpine:3.20` regardless of workspace language. For Python/Node.js builds (Debian-based), Alpine's `apk` would fail on networks where the Alpine CDN is unreachable (e.g., corporate proxies intercepting SSL). The lazygit builder already had dual Alpine/Debian paths — the tree-sitter builder now matches this pattern.

- **`generateTreeSitterBuilder()` accepts `isAlpine bool`** — matches `generateLazygitBuilder()` parameter pattern
- **Python/Node.js (Debian) path** — `FROM debian:bookworm-slim`, `apt-get` with `ca-certificates`, `dpkg --print-architecture` for arch detection
- **Go (Alpine) path** — unchanged; `FROM alpine:3.20`, `apk add --no-cache curl sed`
- **`activeBuilderStages()`** updated to pass `isAlpine` to the tree-sitter builder (matching lazygit wiring)
- 1 new test function (`TestDockerfileGenerator_TreeSitterBuilder_DebianPath`) with 2 subtests

### v0.43.0 (2026-03-16)

**✨ Auto-Token Creation for MaestroVault**

The build pipeline now automatically resolves a MaestroVault token through a priority chain, eliminating the need to manually run `mav token create` and export `MAV_TOKEN` before every build.

- **Token resolution chain** — first non-empty value wins: `MAV_TOKEN` env var → `vault.token` viper config → `~/.devopsmaestro/.vault_token` file → auto-create via `mav token create --name "dvm-auto" --scope read -o json` (persisted to `.vault_token` for reuse)
- **Graceful degradation** — if auto-creation fails (e.g., `mav` not in PATH, no vault database), build continues without vault; no hard failures
- **New `config/vault_token.go`** — `ResolveVaultToken()`, `ResolveVaultTokenFromDir()`, injectable `TokenCreator` type, `persistToken()` with 0600 permissions
- **`Config` struct gains `Vault VaultConfig`** — supports `vault.token` config file path
- **`mav token create` subprocess** uses env allowlist (PATH, HOME only); token values never logged
- 21 new tests in `config/vault_token_test.go`

### v0.42.1 (2026-03-16)

**🐛 Fix Python Private Repo Credential Injection**

Two bugs that prevented Docker build args from reaching pip during Python builds with private git dependencies in `requirements.txt`:

- **ARG before FROM** — `ARG GITHUB_USERNAME` and `ARG GITHUB_PAT` were declared before `FROM`; Docker only makes pre-FROM ARGs available to the `FROM` instruction, not to subsequent `RUN` commands. Fixed by moving `ARG` declarations into `generateBaseStage()` after the `FROM` line.
- **Unnecessary sed substitution removed** — the generated Dockerfile used a `sed` pipeline to substitute `${VAR}` placeholders in `requirements.txt`, but pip natively expands `${VAR}` from environment variables. The sed pipeline was broken (incorrect syntax) and redundant. Removed entirely.

### v0.42.0 (2026-03-16)

**✨ Dynamic Completions & Get All**

- **~48 dynamic tab-completions wired** — all resource types (ecosystems, domains, apps, workspaces, credentials, registries, git repos, nvim plugins/themes, terminal packages/prompts) now have intelligent positional and flag completions
- **Flag completions** added for `--ecosystem`, `--domain`, `--app`, `--workspace`, `--repo`, and `--credential` across all relevant commands
- **Bug fixed** — removed stub registration in `cmd/nvim.go` that was overwriting central completions with empty stubs
- **New `dvm get all` command** — kubectl-style overview showing all 9 resource types; supports `-o json` and `-o yaml`; empty sections show `(none)`
- Shell completion docs updated with full dynamic completion reference

### v0.39.1 (2026-03-12)

**🐛 Change Default Keychain Type to "internet"**

Fixes silent lookup failures for credentials stored via the macOS Passwords app (macOS Sequoia's primary password manager), which exclusively creates `inet` (internet password) entries.

- **Default changed** — `--keychain-type` now defaults to `internet` (`find-internet-password`); credentials stored by the Passwords app, Safari, or iCloud Keychain resolve without extra flags
- **Help text updated** — `--keychain-type` flag description reflects the new default; `generic` must now be specified explicitly for Keychain Access (non-Passwords-app) entries
- **DB migration 012** — Migrates all existing credentials with `keychain_type = "generic"` to `"internet"` so credentials created under v0.39.0's old default are automatically updated

**⚠️ Breaking change:** Users on v0.39.0 who explicitly needed `generic`-type entries and did not pass `--keychain-type generic` must now do so. Since `--keychain-type` was introduced in v0.39.0 (released the same day), no users had a release window to depend on the old default.

### v0.39.0 (2026-03-12)

**✨ Keychain Label-Based Lookup**

Users can now look up macOS keychain entries by **label** (the display name visible in Keychain Access / Passwords app) instead of service name — enabling portable, team-shareable credential configs.

- **`--keychain-label` flag (`-l`)** — identifies keychain entries by display label instead of service/domain string
- **`--keychain-type generic|internet` flag** — explicitly selects Generic Passwords (Keychain Access) or Internet Passwords (Passwords app / Safari); no silent fallback between types
- **`keychainLabel:` in YAML** — portable label reference; team members only need a Keychain entry with a matching label; no machine-specific service URLs in shared configs
- **`keychainType:` in YAML** — explicit keychain class in credential YAML specs
- **`--service` deprecated** — still works with a cobra deprecation warning; migrate to `--keychain-label`
- **`service:` YAML field deprecated** — still parsed for backward compatibility; `keychainLabel:` is now canonical
- **DB migration 011** — adds `label` + `keychain_type` columns, backfills label from existing service values
- 37 new tests across 4 test files

### v0.38.2 (2026-03-12)

**🐛 Credential Resolution Robustness**

Three bugs that caused silent credential failures during `dvm build` and `dvm attach`:

- **(P0) Invisible errors** — `loadBuildCredentials()` now returns warnings displayed via `render.Warning()`; credential resolution failures are surfaced immediately instead of being silently logged
- **(P1) Env var fallback gap** — `ResolveCredentialsWithErrors()` now rescues failed keychain lookups with matching env vars; the "env vars always win" contract is enforced correctly
- **(P1) Missing `-a $USER` filter** — `GetAccountFromKeychain()` now filters by current system user (consistent with all other keychain functions); prevents returning wrong account on multi-account machines

14 new tests across 3 test files.

### v0.38.1 (2026-03-12)

**🐛 Fix Python HTTPS Token Substitution**

Two bugs that broke Dockerfile credential injection for Python workspaces with private repos:

- **SSH regex false positive** — `sshGitPattern` matched `.git@v1.0` version pins in HTTPS URLs, misclassifying them as SSH; fixed to require `git@` not preceded by a dot
- **Dispatch chain priority error** — replaced `if NeedsSSH / else if RequiredBuildArgs` chain with a `switch` on `GitURLType`; HTTPS, SSH, Mixed, and plain cases are now independent and cannot shadow each other

12 new subtests across 2 test files.

### v0.38.0 (2026-03-12)

**✨ Dockerfile Generator Purity & Robustness**

Seven structural improvements eliminating heuristics, injecting dependencies, and propagating errors:

- **(D2 Bug)** `isAlpineImage()` uses computed `isAlpine` field (set during base stage generation) instead of `language == "golang"` heuristic
- **(A2 Bug)** `getDefaultPackages()` uses `isAlpineImage()` instead of direct `strings.Contains()` — both callers are now consistent
- **(A1)** `effectiveVersion()` method centralizes language-default version resolution across all 3 language branches
- **(A4)** `activeBuilderStages()` computed once and passed to both `emitBuilderStages()` and `emitCopyFromBuilders()` — eliminates stage mismatch window
- **(A5 Internal)** `NewDockerfileGenerator()` now accepts `DockerfileGeneratorOptions` struct (replaces 6 positional parameters)
- **(D3)** `PathConfig` injection via options struct — `generateNvimSection()` uses injected path, no more `os.UserHomeDir()` calls in production
- **(D1)** `generateNvimSection()` returns `error`; `Generate()` propagates it — no more silent nvim config failures

6 new tests.

### v0.37.5 (2026-03-12)

**✨ BuildKit Structural Improvements**

- **Tree-sitter dynamic versioning** — version fetched from GitHub API at build time (same pattern as lazygit); replaces hardcoded `v0.24.6`
- **Nil workspace guard** — `Generate()` returns `fmt.Errorf` instead of panicking on nil workspace
- **`builderStage` struct** — single source of truth for which builder stages are emitted and copied; eliminates the mirrored-conditional window
- **Fail-fast on unknown architecture** — all 4 builder stages `exit 1` on unsupported arch instead of silently falling back to x86_64
- **`effectiveUser()` helper** — `generateNvimSection()` uses configured user instead of hardcoded `/home/dev/`
- **Lazygit download consolidation** — Alpine and Debian paths share common download logic section

6 new tests.

### v0.37.4 (2026-03-12)

**🔨 BuildKit Builder Stage Robustness**

Hardened all 5 Dockerfile builder stages to eliminate silent download failures (e.g., `"/usr/local/bin/starship": not found`):

- **`curlFlags` constant** — centralized `−fsSL --retry 3 --connect-timeout 30` across all builder stages
- **`set -e` in all builder `RUN` commands** — any command failure immediately aborts the stage
- **Download-to-file pattern** — starship and golangci-lint replaced pipe-to-shell with download-then-execute
- **`test -x` binary verification** — every builder stage asserts its binary exists and is executable before completing
- **Lazygit `[ -n "$LAZYGIT_VERSION" ]` guard** — fails explicitly if the GitHub API returns empty

6 new tests.

### v0.37.3 (2026-03-12)

**🔒 Security Hardening**

- **(HIGH)** Checksum verification for downloaded registry binaries using `crypto/subtle.ConstantTimeCompare()`; `io.LimitReader` caps body reads at 500 MB
- **(MEDIUM)** Defensive 5-minute timeout on Athens binary downloads
- **(MEDIUM)** Config and log file permissions hardened from 0644 → 0600 in all 5 registry managers
- **(MEDIUM)** Storage path validation (`validateStoragePath()`) in `pkg/registry/strategy.go` — rejects path-traversal attempts
- **(LOW)** Minimum idle timeout enforcement (60 s minimum for on-demand registries)

12 new tests.

### v0.37.2 (2026-03-12)

**🐛 Registry Bug Fixes**

- **(P0)** `downloadBinary()` applies a defensive 5-minute context timeout — prevents indefinite hangs on slow networks
- **(P1)** Log file handle stored in struct and closed in `Stop()` — prevents premature close while child Zot process still writes
- **(P1)** All 4 registry creation paths set version from strategy's `GetDefaultVersion()` — fixes empty version string causing `EnsureBinary()` to skip reconciliation
- **(P2)** Idle timer callback wrapped in goroutine — eliminates latent deadlock across all 5 registry manager types

4 new tests.

### v0.37.1 (2026-03-11)

**✨ Keychain Dual-Field Credentials**

A single `dvm create credential` can now extract both the account (username) and password from one macOS Keychain entry:

- **`--username-var <VAR>`** — env var name for the Keychain account field
- **`--password-var <VAR>`** — env var name for the Keychain password field
- **`CredentialsToMap()` fanout** — 1 DB row → 2 `CredentialConfig` entries
- **DB migration 010** — adds `username_var` + `password_var` columns
- **YAML apply support** — `usernameVar` / `passwordVar` in credential YAML specs

Example:
```bash
dvm create credential github-creds \
  --source keychain --service github.com \
  --username-var GITHUB_USERNAME \
  --password-var GITHUB_PAT \
  --ecosystem myorg
```

42 new tests.

### v0.37.0 (2026-03-11)

**✨ Runtime Credential & Env Injection**

- **Runtime credential injection** — `dvm attach` injects credentials from the 5-layer hierarchy into the running container via `loadBuildCredentials()`
- **Runtime registry env injection** — `loadRegistryEnv()` generates `PIP_INDEX_URL`, `GOPROXY`, `NPM_CONFIG_REGISTRY`, etc. at attach time
- **4-map merge policy** — `buildRuntimeEnv` merges theme → registry → credentials → workspace env; metadata vars (`DVM_*`) always applied last
- **`--env KEY=VALUE` flag** on `dvm create workspace` — inline env var setting at creation time; validates POSIX names and blocks `DVM_*` reserved prefix
- **Bootstrap all 5 default registries** on `dvm admin init` — OCI (zot:5001), PyPI (devpi:5002), npm (verdaccio:5003), Go (athens:5004), HTTP (squid:5005)
- **Shell-escape Colima env values** — single-quote wrapping prevents shell injection
- **Expanded `NO_PROXY`** — adds RFC1918 ranges (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16)

~25 new tests across 5 test files.

### v0.36.1 (2026-03-11)

**✨ Default OCI Registry on Init**

`dvm admin init` now auto-creates a default Zot OCI registry (port 5001, on-demand lifecycle) so `dvm build` works out-of-the-box without manual `dvm registry enable oci`. Idempotent — re-running init skips creation if registry already exists.

5 new tests.

### v0.36.0 (2026-03-11)

**✨ Credential Injection & Bug Fixes**

- **SQLite foreign key enforcement** — `_foreign_keys=on` DSN parameter + PRAGMA verification; cascade deletes now propagate correctly
- **GitRepo single-item render fix** — `dvm get gitrepo <name>` now renders with `KeyValueData` instead of raw `map[string]interface{}`
- **GitRepo `--credential` flag wired** — previously registered but discarded as dead code; now validates and stores the credential reference
- **Runtime env injection** — `dvm attach` injects workspace env vars and theme env vars into the container
- **Workspace `env` field** — DB migration 009; `GetEnv()`/`SetEnv()` helpers; round-trips through `ToYAML()`/`FromYAML()`
- **Build-arg credential redaction** — `redactBuildArgs()` replaces credential values with `***REDACTED***` in logged Docker build-arg strings
- **`pkg/envvalidation/` package** — POSIX env var name validation, dangerous env var denylist (`LD_PRELOAD`, `DYLD_INSERT_LIBRARIES`, etc.)

### v0.35.2 (2026-03-12)

**🐛 Registry Binary Version Reconciliation Fix**

- **`GetVersion()` reads Zot stderr correctly** — previously ran `zot version` (invalid subcommand) and always returned empty string; now reads `--version` JSON from stderr
- **Backup/rollback preserved** — failed binary downloads restore the previous working binary so the registry remains runnable
- **Defensive timeout on Athens downloads** — mirrors the Zot fix

### v0.35.1 (2026-03-12)

**✨ Declarative Registry Version Management**

- **`--version` flag on `dvm create registry`** — specifies desired binary version at creation time; validated as semver
- **VERSION column** in `dvm get registries` table output
- **Version in `dvm get registry <name>`** detail view and `-o yaml` output
- **`GetDefaultVersion()` interface method** on `RegistryStrategy` — Zot returns `"2.1.15"`, others `""`
- **`EnsureBinary()` version reconciliation** — checks installed vs. desired version on every start; downloads correct version if mismatched
- **DB migration 008** — adds `version` column to `registries` table

8 new manual test scenarios (Scenarios 32–39).

### v0.35.0 (2026-03-11)

**✨ Credential CLI Feature**

User-facing CLI commands for credential management. The credential backend (database, config resolution, keychain integration, build pipeline) already existed; this release adds the missing CLI surface.

- **create credential** - `dvm create credential <name> --source keychain|env [--service svc] [--env-var var]` with scope flags; aliases `cred`/`creds`
- **get credentials** - `dvm get credentials` lists credentials in active context; `-A/--all` lists all scopes
- **get credential** - `dvm get credential <name>` shows a single credential with scope flags
- **delete credential** - `dvm delete credential <name>` with interactive confirmation; `--force/-f` skips prompt
- **apply credential** - `dvm apply -f credential.yaml` — `Credential` kind now fully supported

New model structs (`CredentialYAML`, `CredentialMetadata`, `CredentialSpec`) with `ToYAML()`/`FromYAML()`, `ScopeInfo()`, and `ValidateCredentialYAML()`. Full `CredentialHandler` registered in the apply pipeline. 69 new tests (42 CLI + 13 model + 14 handler). Manual Test Plan Part 6 added (Scenarios 17–31).

### v0.32.8 (2026-03-05)

**🐛 Bug Fixes & Integration Confirmations**

Targeted fixes for library commands, `set theme` flag handling, and workspace branch creation:

- **library list** - `dvm library list "nvim packages"` now works; quoted space-separated type args are normalised to hyphenated equivalents alongside `nvim-packages`
- **library show** - `dvm library show nvim-package <name>` and `dvm library show terminal-package <name>` now display correctly in table format
- **set theme** - `--workspace` and `--app` can now be combined on `dvm set theme`; overly restrictive Cobra mutual exclusivity replaced with manual flag validation
- **create branch** - New `dvm create branch <name>` command and `--create-branch` flag for `dvm create workspace`; clone vs checkout errors differentiated via `ClonePhaseError` sentinel types
- **language detection** - `getLanguageFromApp` correctly uses workspace source path for accurate language detection
- **library import** - `dvm library import` confirmed working (no code changes required)

New test files: `cmd/library_test.go`, `cmd/set_theme_test.go`, `cmd/create_test.go`, `cmd/build_language_test.go`, `cmd/library_import_test.go`, `utils/language_detector_test.go`

### v0.32.4 (2026-03-04)

**🐛 Bug Fixes & Features**

Critical fixes for workspace builds, database errors, and CLI routing:

- **Build fallback** - Workspaces with no plugins configured now fall back to embedded core package (treesitter, telescope, lspconfig, mason)
- **nvp error handling** - `nvp package install` shows clear error when database not initialized (guides user to run `dvm admin init`)
- **CLI routing** - Fixed `dvm get app <name>` to return single app (removed alias conflict)
- **Branch selection** - New `--branch` flag for `dvm create workspace --repo <name> --branch <branch>` enables feature branch workflows

### v0.32.3 (2026-03-04)

**✨ Build Optimization**

Pre-install development tools at Docker image build time for faster workspace startup:

- **Mason LSP installation** - Language servers (pyright, gopls, rust-analyzer, etc.) now installed during build
- **Treesitter parsers** - Syntax parsers pre-installed per language during build
- **Documentation fixes** - Corrected `dvm init` → `dvm admin init` across all docs

### v0.32.2 (2026-03-04)

**🐛 Bug Fixes**

- **GitRepo inheritance** - Workspaces now inherit GitRepo from parent app
- **Nvim config path** - Fixed nvim configuration not being copied for GitRepo-backed workspaces

### v0.32.1 (2026-03-04)

**🐛 Bug Fixes & Refactoring**

Fixed error handling in `--repo` flag and extracted watchdog helper:

- **Error handling** - Fixed error message check causing "not found" errors to be misclassified as database errors
- **Error messages** - Improved slug conflict and GitRepo not found error messages with helpful examples
- **Watchdog refactor** - Extracted watchdog helper to dedicated module with `WatchdogConfig` and injection pattern
- **Test coverage** - Added 22 new tests (11 for watchdog, 11 for GitRepo resolution)
- **Test fixes** - Unskipped 4 previously skipped tests after error handling improvements

### v0.32.0 (2026-03-04)

**✨ Feature - `--repo` Flag for App Creation**

Streamlined GitRepo-backed app creation with new `--repo` flag:

- **Accept URL** - `dvm create app my-app --repo https://github.com/user/repo.git`
- **Accept name** - `dvm create app my-app --repo my-existing-repo`
- **Auto-create GitRepo** - Automatically creates GitRepo resource when given a URL
- **Detect duplicates** - Reuses existing GitRepos by URL to avoid duplicates
- **Mutually exclusive** - Cannot use with `--path` or `--from-cwd` flags

**🐛 Bug Fix - Docker Build Hang on Colima**

Fixed Docker buildx + Colima hang where build completes but process doesn't exit:

- **Watchdog mechanism** - Polls for image existence during build
- **Context cancellation** - Terminates hung docker process when image is detected
- **Goroutine execution** - Runs docker build in background for parallel monitoring
- **Timeout protection** - 30-minute overall timeout as fallback
- **Reliable Colima builds** - Enables reliable builds on Colima with Docker runtime (non-containerd)

### v0.31.0 (2026-03-03)

**✨ Feature - Lazygit in Containers**

All development containers now include lazygit for terminal-based git operations:

- **Automatic installation** - Lazygit installed from GitHub releases during container builds
- **Multi-architecture** - Supports x86_64 and ARM64 architectures
- **Multi-distro** - Works with Alpine (musl) and Debian-based images
- **Stable release** - Downloads latest stable version from jesseduffield/lazygit

Every workspace now has `lazygit` available at `/usr/local/bin/lazygit`.

### v0.30.4 (2026-03-03)

**🐛 Bug Fix - Attach Mount Path for GitRepo-backed Workspaces**

Fixed `dvm attach` mounting wrong directory for GitRepo-backed workspaces. Container now correctly mounts workspace repo path instead of empty app path.

### v0.30.3 (2026-03-03)

**🐛 Bug Fix - Database Schema Drift**

Fixed schema drift where `apps.git_repo_id` column existed in database but not in Go code. All 7 App DataStore methods now include `git_repo_id` field.

### v0.30.2 (2026-03-03)

**🐛 Bug Fixes - Theme & Terminal Package Integration**

Four critical fixes for workspace theme inheritance and terminal package prompt generation:

**Theme Inheritance Hierarchy Fix:**
- `dvm build` now properly walks Workspace → App → Domain → Ecosystem → Global hierarchy when resolving themes
- Previously bypassed hierarchy and always used global default theme
- Workspaces now correctly inherit themes from their parent resources

**Terminal Package Prompt Composition Fix:**
- `generateShellConfig()` now loads terminal packages from library and composes prompts from style + extensions
- Previously ignored workspace's `terminal-package` setting, always generating default prompts
- Custom terminal packages like `rmkohlman` now generate rich prompts as designed

**CoolNight Theme Prompt Colors:**
- Added monochromatic `promptColors` gradients to all 21 CoolNight theme variants (except ocean)
- Previously missing `promptColors` section caused harsh ANSI fallback colors
- Terminal prompts now have smooth, cohesive color gradients matching the workspace theme

**Neovim Colorscheme Generation:**
- `dvm build` now generates `theme/colorscheme.lua` containing `vim.api.nvim_set_hl()` calls
- Previously only generated 3 theme files, missing the colorscheme file
- Neovim inside workspaces now displays correct theme colors matching the terminal prompt

**Testing:**
- Added `cmd/build_terminal_package_test.go` (9 tests)
- Added `cmd/build_theme_test.go` (5 tests)
- Added `cmd/set_theme_test.go` (theme setting tests)

### v0.30.1 (2026-03-03)

**🐛 Bug Fixes - Database Schema & YAML Completeness**

Seven critical fixes for database schema completeness and YAML field parsing:

- **Registry database queries** - Fixed missing `storage`, `enabled`, `idle_timeout` columns (GitHub Issue #5)
- **Workspace database queries** - Fixed missing terminal configuration columns (GitHub Issue #8)
- **Terminal package validation** - Now checks database + embedded library (GitHub Issue #7)
- **Library show commands** - Fixed table output for library resources (GitHub Issue #6)
- **Git checkout** - Fixed workspace creation with `--repo` flag (GitHub Issue #9)
- **YAML apply workspace fields** - Fixed incomplete field parsing (GitHub Issue #10)
- **Workspace YAML serialization** - Fixed incomplete YAML output (GitHub Issue #11)

### v0.24.0 (2026-03-01)

**🔄 Registry Resources - Multi-Registry Support**

Major refactor of registry system to support multiple registry types with database-backed resources.

**⚠️ Breaking Changes:**

- **Registry commands now require name as positional argument**:
  - OLD: `dvm registry start` → NEW: `dvm registry start myregistry`
  - OLD: `dvm registry stop` → NEW: `dvm registry stop myregistry`
- **Removed `--name` flag** from all registry runtime commands
- **Removed config-based registry** - Must use Registry Resources instead
- **Must create registry resource first**: `dvm create registry <name> --type <type> --port <port>`

**New Features:**

- **Registry Resource Type** - Database-backed registry management:
  - `dvm create registry <name> --type <type> --port <port>` - Create registry
  - `dvm get registries` - List all registries
  - `dvm get registry <name>` - Show specific registry
  - `dvm delete registry <name>` - Delete registry
  
- **Multi-Registry Support** - Run multiple registry types simultaneously:
  - `zot` - OCI container images (full support)
  - `athens` - Go module proxy (stub)
  - `devpi` - Python package index (stub)
  - `verdaccio` - npm registry (stub)
  - `squid` - HTTP proxy cache (stub)

- **ServiceFactory Pattern** - Extensible architecture for registry services:
  - Each registry type has dedicated service implementation
  - Database persistence for all registry configurations
  - Independent lifecycle management per registry

**Updated Commands:**

- `dvm registry start <name>` - Start specific registry (name REQUIRED)
- `dvm registry stop <name>` - Stop specific registry (name REQUIRED)
- `dvm registry status` - List all registries (no name = show all)
- `dvm registry status <name>` - Show specific registry status

**Migration Guide:**

```bash
# OLD approach (no longer works):
dvm registry start

# NEW approach (required):
# Step 1: Create registry resource
dvm create registry myregistry --type zot --port 5000

# Step 2: Start the registry
dvm registry start myregistry

# Check status
dvm registry status myregistry
# or list all:
dvm registry status
```

**Technical Details:**

- New Registry Resource model with database persistence
- ServiceFactory pattern for multi-registry support
- Zot service implementation with full lifecycle management
- Stub implementations for Athens, Devpi, Verdaccio, Squid
- Full test coverage for CRUD operations and runtime commands

### v0.22.0 (2026-02-28)

**🔗 Integration - Consolidated Library & Terminal Configuration**

Complete integration of nvp/dvt functionality into dvm. Users can now browse libraries and configure terminal settings without leaving the dvm CLI.

**New Features:**

- **Library Browsing Commands** - Browse all libraries from dvm:
  - `dvm library list plugins` - 38+ nvim plugins
  - `dvm library list themes` - 34+ nvim themes
  - `dvm library list nvim packages` - Nvim bundles
  - `dvm library list terminal prompts` - 5 terminal prompts
  - `dvm library list terminal plugins` - 8 shell plugins
  - `dvm library list terminal packages` - Terminal bundles
  - `dvm library show <resource> <name>` - Detailed info
  - **Aliases**: `lib` → `library`, `ls` → `list`, `np` → `plugins`, `nt` → `themes`, `tp` → `terminal prompts`

- **Terminal Configuration** - Configure terminal per-workspace:
  - `dvm set terminal prompt -w <workspace> <name>` - Set prompt
  - `dvm set terminal plugin -w <workspace> <plugins...>` - Set plugins
  - `dvm set terminal package -w <workspace> <name>` - Set package bundle
  - Workspace-specific terminal configuration stored in database
  - Validation ensures resources exist in library

**Example Workflow:**
```bash
# Browse available themes
dvm lib ls themes

# Set terminal prompt for workspace
dvm set terminal prompt -w dev starship-minimal

# Set shell plugins
dvm set terminal plugin -w dev zsh-autosuggestions fzf

# Verify workspace config
dvm get workspace dev -o yaml
```

**Technical Details:**
- New database migration `004_add_terminal_fields`
- 104 new integration tests following TDD
- Workspace model extended with TerminalPrompt, TerminalPlugins, TerminalPackage fields

### v0.21.0 (2026-02-28)

**🚀 Local OCI Registry (Zot) - Container Image Caching**

> **Note**: This version has been superseded by v0.24.0 which introduces Registry Resources. See v0.24.0 for current registry usage.

**Legacy Features (replaced in v0.24.0):**

- Local OCI Registry with pull-through caching
- Registry CLI commands (now require positional arguments in v0.24.0)
- Build integration with `--no-cache`, `--push`, `--registry` flags

See [Full CHANGELOG](https://github.com/rmkohlman/devopsmaestro/blob/main/CHANGELOG.md) for complete v0.21.0 details.

### v0.20.1 (2026-02-28)

**🔗 GitRepo-Workspace Integration**

**New Features:**

- **Workspace Creation with GitRepo** - `--repo` flag for `dvm create workspace`
  - Associate workspaces with existing GitRepo resources
  - Automatically clones from local mirror to workspace's `repo/` directory
  - Each workspace gets independent clone for isolated development
  - Example: `dvm create workspace dev --repo my-project`
  
- **Auto-Sync Control** - `--no-sync` flag for `dvm attach`
  - Skip automatic mirror sync before attaching to workspace
  - Default: Syncs mirror if GitRepo has AutoSync=true
  - Use `--no-sync` for faster attach or offline work
  - Sync failures are warnings, not fatal errors
  - Example: `dvm attach --no-sync`

**Workflow:**
```bash
# 1. Create git mirror
dvm create gitrepo my-project --url https://github.com/myorg/my-project

# 2. Create workspace with repo
dvm create workspace dev --repo my-project

# 3. Attach (auto-syncs by default)
dvm attach

# Or skip sync for faster attach
dvm attach --no-sync
```

### v0.19.0 (2026-02-28)

**🚀 Full Workspace Isolation**

**🚀 Full Workspace Isolation**

**⚠️ BREAKING CHANGES - Fresh Database Required**

This is a major breaking release that requires a fresh database. **You must backup and delete your existing database before upgrading.**

- Fresh database schema required - existing databases incompatible
- Removed `projects` table - use Ecosystem → Domain → App hierarchy
- Removed credential `value` field - plaintext storage no longer supported
- Removed SSH key auto-mounting - use SSH agent forwarding instead

**New Features:**

- **Workspace Isolation** - Each workspace has dedicated isolated directories:
  - `~/.devopsmaestro/workspaces/{slug}/repo/` - Git repository clone
  - `~/.devopsmaestro/workspaces/{slug}/volume/` - Persistent data (nvim-data, cache)
  - `~/.devopsmaestro/workspaces/{slug}/.dvm/` - Generated configs
- **Workspace Slug** - Unique identifier format: `{ecosystem}-{domain}-{app}-{workspace}`
- **SSH Agent Forwarding** - Opt-in via `--ssh-agent` flag or `ssh_agent_forwarding: true` in YAML
- **Enhanced Security** - SSH keys never mounted, credentials limited to keychain/env only

**Migration Steps:**

1. Backup: `dvm get <resources> -o yaml > backup.yaml`
2. Delete database: `rm ~/.devopsmaestro/devopsmaestro.db`
3. Upgrade: `brew upgrade devopsmaestro`
4. Re-initialize: `dvm admin init`
5. Re-apply resources: `dvm apply -f backup.yaml`

See the [full migration guide](https://github.com/rmkohlman/devopsmaestro/blob/main/CHANGELOG.md#migration-guide) for detailed instructions.

### v0.18.25 (2026-02-28)

**Fix Coolnight Theme Git Clone Errors**

- Fixed git clone failures for all 21 coolnight parametric themes by converting them to standalone mode
- Coolnight themes no longer require cloning external `rmkohlman/coolnight.nvim` repo (which doesn't exist)
- Standalone themes apply colors directly via `nvim_set_hl()` for better reliability
- **Breaking change**: Existing workspaces using coolnight themes need `dvm build` to regenerate config

### v0.18.24 (2026-02-23)

**Hierarchical Container Naming & Starship Theme Colors**

- Container names now include full hierarchy path: `dvm-{ecosystem}-{domain}-{app}-{workspace}` for better identification
- Added ecosystem/domain labels (`io.devopsmaestro.ecosystem`, `io.devopsmaestro.domain`) and environment variables to containers
- Starship prompts now automatically use active workspace theme colors via new ColorToPaletteAdapter
- Backward compatible fallback to `dvm-{app}-{workspace}` naming when hierarchy unavailable

### v0.18.23 (2026-02-23)

**Theme Database Persistence & Output Formatting**

- Fixed theme persistence bug where `dvm set theme --workspace X` wasn't saving to database
- Fixed theme output formatting that was showing raw struct instead of clean key-value format
- Theme values now properly appear in `dvm get workspace -o yaml` and all workspace queries

### v0.18.22 (2026-02-23)

**Shell Completion Enhancements**

- Added comprehensive tab completion for all resource commands (`dvm get ecosystem <TAB>`, etc.)
- All `--ecosystem`, `--domain`, `--app`, `--workspace` flags now complete resource names on TAB
- Fixed `dvm nvim sync <TAB>` and `dvm nvim push <TAB>` to complete workspace names correctly
- Provides kubectl-style CLI experience with real-time completion based on current database state

### v0.18.21 (2026-02-23)

**Theme Visibility Fix**

- Fixed bug where `dvm set theme` stored themes correctly but theme values were missing from YAML output
- Theme values now properly appear in `dvm get workspace -o yaml` and similar commands for all resource types
- Added dedicated theme column to workspaces table and updated all model ToYAML/FromYAML methods

### v0.18.20 (2026-02-23)

**Container Label Compatibility Fix**

- Fixed "name already in use" error when attaching to workspaces after upgrading from pre-v0.18.18 versions
- Containers created before the image tracking feature (v0.18.18) are now automatically detected and recreated with proper labels
- Provides seamless upgrade experience without requiring manual container cleanup

### v0.18.19 (2026-02-23)

**Mason Toolchain Auto-Installation**

- When Neovim is configured for a workspace, generated containers now automatically include npm, cargo, and pip toolchains
- Enables Mason to install language servers, formatters, and linters (stylua, prettier, pyright) without "executable not found" errors
- Supports both Alpine and Debian-based images with cross-platform package management

### v0.18.18 (2026-02-23)

**Containerd Runtime Image Change Detection**

- Fixed critical bug where containerd/Colima runtime was reusing running containers without checking if image had changed
- `dvm build --force --no-cache` now properly recreates containers with new images instead of reusing stale ones
- Added `io.devopsmaestro.image` label tracking to detect when underlying image changes
- Brings containerd runtime behavior in line with Docker runtime for consistent experience

### v0.18.17 (2026-02-21)

**Docker Build Context Fix**

- Fixed critical bug where `Dockerfile.dvm` was saved to the original app directory but Docker build used staging directory
- Docker COPY commands now work correctly with generated config files like `.config/starship.toml`
- Eliminates container configuration issues where generated files weren't found during build

### v0.18.16 (2026-02-21)

**Shell Configuration Fix**

- Fixed critical bug where shell configuration (starship.toml, .zshrc) was only generated when nvim was configured
- `dvm build` now always generates shell config regardless of nvim configuration, eliminating TOML parse errors in containers
- Refactored build flow to separate shell config generation from nvim config for better reliability

### v0.18.14 (2026-02-20)

**Plugin Storage Compatibility Fix**

- Fixed critical bug where `nvp package install` only saved to FileStore but `dvm build` reads from database
- `nvp package install` now saves plugins to BOTH FileStore and database for full compatibility
- Added plugin library fallback in `dvm build` when plugins not found in database
- Plugins installed via `nvp` are now immediately available to `dvm build`

### v0.18.6 (2026-02-20)

**Terminal Emulator Management (Phase 3 - Build Integration)**

- Added embedded emulator library with 6 curated configurations (rmkohlman, minimal, developer, alacritty-minimal, kitty-poweruser, iterm2-macos)
- New CLI commands: `dvt emulator install <name>`, `dvt emulator apply -f <file>`, `dvt emulator library list/show`
- Build integration: `dvm build` now generates WezTerm config and loads terminal plugins automatically
- Complete WezTerm configuration mapping from database (fonts, colors, keybindings, etc.)
- Terminal plugin loading in `.zshrc` with support for manual, oh-my-zsh, and zinit managers

### v0.18.5 (2026-02-20)

**Terminal Emulator Management (Phase 2)**

- Added complete terminal emulator database infrastructure with `terminal_emulators` table
- Support for wezterm, alacritty, kitty, and iterm2 emulator types with JSON configuration storage
- New `dvt emulator` CLI commands: `list`, `get`, `enable`, `disable` with filtering and formatting options
- Created `pkg/terminalops/emulator/` domain layer with proper interfaces and error handling
- Database adapter `DBEmulatorStore` following established DevOpsMaestro patterns
- Theme and workspace association support for coordinated terminal styling

### v0.18.4 (2026-02-20)

**Terminal Plugin Database Support**

- Added terminal plugin database support with `terminal_plugins` table for persistent storage
- Created database adapter for terminal plugins following nvimops pattern (`pkg/terminalops/store/db_adapter.go`)
- `dvt package install` now persists plugins to database for consistency with nvp
- `dvt plugin list` and `dvt plugin get` commands now read from database instead of file storage
- Enhanced cross-command integration - all terminal plugin operations use shared DevOpsMaestro database
- **Migration**: Plugin data moved from files to database; existing users may need to reinstall plugins from library

### v0.18.3 (2026-02-20)

**Build & Configuration Fixes**

- Fixed Starship TOML parsing error in containerized environments
- Fixed `dvm build` to respect user's default nvim package when workspace has no explicit plugins
- Fixed `dvm get defaults` to show actual user-configured values instead of hardcoded defaults
- Renamed nvp package `rkohlman-full` → `rmkohlman` for consistent naming

### v0.18.2 (2026-02-20)

**ARM64/Apple Silicon Stability**

- Fixed `dvm build` failing on ARM64 (Apple Silicon) with dpkg errors  
- Replaced gcc + python3-dev with build-essential for better ARM64 compatibility
- Added --fix-broken flag to apt-get commands for robust package installation
- Pinned Python base images to bookworm variant for reproducible builds
- Enhanced Docker cleanup process for better ARM64 system stability

### v0.18.1 (2026-02-20)

**Neovim Installation Fix**

- Fixed `dvm build` failures on slim Docker images (python:3.11-slim, etc.)
- Neovim now installed from GitHub releases instead of apt-get
- Works on all base images: Debian, Ubuntu, Alpine, and slim variants
- Multi-architecture support (amd64/arm64) with automatic detection

### v0.18.0 (2026-02-20)

**DVT Package Management**

- `dvt package list/get/install` commands for terminal package management
- Embedded terminal package library (core, developer, rmkohlman packages)
- Package inheritance support with automatic resolution
- Dry-run installation support for safe previews
- Parity with NvimPackage system for consistent package management

### v0.17.0 (2026-02-20)

**DVT Binary Release & TerminalPackages**

- DVT (TerminalOps) binary now included in releases alongside dvm and nvp
- New TerminalPackage resource type for terminal configuration bundles
- Terminal defaults management: `dvm use terminal package` and `dvm get terminal` commands  
- Auto-create NvimPackage after `nvp source sync` operations
- Mandatory test gate requirement for all releases (100% test success)

### v0.16.0 (2026-02-20)

**Package Management System**

- New NvimPackage resource type with kubectl-style CRUD operations
- Package defaults management for new workspaces
- External source sync (LazyVim integration)
- Auto-migration on startup for seamless upgrades

### v0.15.0 (2026-02-19)

**GitHub Directory & Secret Providers**

- Apply all YAML files from GitHub directories: `dvm apply -f github:user/repo/plugins/`
- Secret provider system with Keychain and environment support
- Inline secret syntax: `${secret:name}` in YAML resources
- DirectorySource interface for extensible directory-based sources

### v0.12.0 (2026-02-19)

**Hierarchical Theme System**

- Multi-level theme configuration: Workspace → App → Domain → Ecosystem → Global
- `dvm set theme` command with hierarchy flags
- 21 CoolNight theme variants with parametric generation
- kubectl-style theme IaC with file/URL/GitHub support

### v0.10.0 (2026-02-19)

**Plugin Packages & Enhanced Features**

- Plugin packages system with inheritance support
- Enhanced keymap generation for vim.keymap.set() calls
- Terminal theme integration with workspace context
- Smart workspace resolution with hierarchy flags

---

## Version History

| Version | Date | Highlights |
|---------|------|------------|
| **0.43.1** | 2026-03-16 | Fix tree-sitter builder for Debian — dual Alpine/Debian paths matching lazygit builder pattern; fixes corporate proxy SSL failures on Python/Node.js builds |
| **0.43.0** | 2026-03-16 | Auto-token creation for MaestroVault — priority chain resolves token automatically; no manual `mav token create` required |
| **0.42.1** | 2026-03-16 | Fix Python private repo credential injection — ARG moved after FROM; sed pipeline removed (pip handles `${VAR}` natively) |
| **0.39.1** | 2026-03-12 | Default keychain type changed to "internet" — fixes Passwords app / Safari / iCloud Keychain silent failures; DB migration 012 |
| **0.39.0** | 2026-03-12 | Keychain label-based lookup — `--keychain-label`, `--keychain-type`, `keychainLabel:` YAML, internet password support; `--service` deprecated |
| **0.38.2** | 2026-03-12 | Credential resolution robustness — visible warnings, env var rescue, keychain `-a $USER` filter |
| **0.38.1** | 2026-03-12 | Python HTTPS token substitution fix — SSH regex false positive + dispatch chain priority error |
| **0.38.0** | 2026-03-12 | Dockerfile generator purity — computed Alpine detection, options struct, PathConfig injection, error propagation |
| **0.37.5** | 2026-03-12 | BuildKit structural improvements — dynamic tree-sitter versioning, builderStage struct, arch fail-fast, custom user nvim fix |
| **0.37.4** | 2026-03-12 | Builder stage robustness — curl retry, binary verification, set -e, download-to-file pattern |
| **0.37.3** | 2026-03-12 | Security hardening — checksum verification, file permissions 0600, storage path validation, idle timeout minimum |
| **0.37.2** | 2026-03-12 | Registry bug fixes — download timeout, log file lifecycle, version defaults, idle timer deadlock |
| **0.37.1** | 2026-03-11 | Keychain dual-field credentials — single keychain entry yields username + password env vars |
| **0.37.0** | 2026-03-11 | Runtime credential & env injection — credentials, registry env, workspace env injected into containers at attach time |
| **0.36.1** | 2026-03-11 | Default OCI registry auto-created on `dvm admin init` |
| **0.36.0** | 2026-03-11 | Credential injection & bug fixes — cascade deletes, gitrepo render, build arg redaction, env validation package |
| **0.35.2** | 2026-03-12 | Registry binary version reconciliation fix — Zot `GetVersion()` stderr parsing, backup/rollback |
| **0.35.1** | 2026-03-12 | Declarative registry version management — `--version` flag, VERSION column, `EnsureBinary()` reconciliation |
| **0.35.0** | 2026-03-11 | Credential CLI feature — `create`, `get`, `list`, `delete`, `apply` commands for credentials |
| **0.32.8** | 2026-03-05 | library list quoted args fix, library show nvim/terminal-package, set theme flag fix, create branch command |
| **0.32.1** | 2026-03-04 | Error handling fixes, watchdog refactor, improved GitRepo resolution |
| **0.32.0** | 2026-03-04 | `--repo` flag for app creation, Docker build hang fix on Colima |
| **0.31.0** | 2026-03-03 | Lazygit in containers, multi-architecture support |
| **0.30.4** | 2026-03-03 | Attach mount path fix for GitRepo-backed workspaces |
| **0.30.3** | 2026-03-03 | Database schema drift fix (git_repo_id column) |
| **0.30.2** | 2026-03-03 | Theme inheritance & terminal package prompt fixes, colorscheme generation |
| **0.30.1** | 2026-03-03 | Database schema completeness, YAML field parsing fixes |
| **0.24.0** | 2026-03-01 | Registry resource refactor, multi-registry support (BREAKING) |
| **0.22.0** | 2026-02-28 | Library browsing (dvm library), terminal configuration (dvm set terminal), integration |
| **0.21.0** | 2026-02-28 | Local OCI registry (Zot), pull-through cache, offline builds, registry CLI commands |
| **0.20.1** | 2026-02-28 | GitRepo-Workspace integration (--repo flag, --no-sync flag, auto-sync on attach) |
| **0.20.0** | 2026-02-28 | Git repository mirror management, bare mirrors, MirrorManager package, security validation |
| **0.19.0** | 2026-02-28 | Full workspace isolation, SSH agent forwarding, security hardening (BREAKING) |
| **0.18.17** | 2026-02-21 | Docker build context fix for generated config files |
| **0.17.0** | 2026-02-20 | DVT binary release, TerminalPackages, test gate requirement |
| **0.16.0** | 2026-02-20 | Package management system, auto-migration |
| **0.15.1** | 2026-02-19 | NvimPlugin opts field support fix |
| **0.15.0** | 2026-02-19 | GitHub directory support, secret providers |
| **0.14.0** | 2026-02-19 | TerminalPrompt resources, dvt prompt CLI |
| **0.13.1** | 2026-02-19 | `dvm get defaults` command |
| **0.13.0** | 2026-02-19 | Container build improvements, staging directory |
| **0.12.3** | 2026-02-19 | Comprehensive YAML reference documentation |
| **0.12.2** | 2026-02-19 | WezTerm CLI commands with theme integration |
| **0.12.1** | 2026-02-19 | Default nvim configuration for new workspaces |
| **0.12.0** | 2026-02-19 | Hierarchical theme system, 21 CoolNight variants |
| **0.11.0** | 2026-02-19 | Terminal theme integration, environment variables |
| **0.10.0** | 2026-02-19 | Plugin packages system, keymap generation fixes |
| **0.9.7** | 2026-02-18 | Colima SSH command fix |
| **0.9.6** | 2026-02-18 | Colima path lookup fix |
| **0.9.5** | 2026-02-18 | Container detached mode fix |
| **0.9.4** | 2026-02-18 | Colima containerd SSH operations fix |
| **0.9.3** | 2026-02-18 | Container attach consistency fix |
| **0.9.2** | 2026-02-18 | ColorProvider architecture, dynamic completions |
| **0.9.1** | 2026-02-17 | `dvm get workspaces -A` flag, Colima mount fix |
| **0.9.0** | 2026-02-17 | Smart workspace resolution with hierarchy flags |
| **0.8.0** | 2025-01-06 | New object hierarchy (Ecosystem/Domain/App/Workspace) |
| **0.7.1** | 2026-02-04 | Unified resource pipeline |
| **0.7.0** | 2026-02-03 | Terminal resize, timestamp tags |
| **0.6.0** | 2026-02-03 | Status command, aliases |
| **0.5.1** | 2026-02-02 | BuildKit socket validation fix |
| **0.5.0** | 2026-01-30 | Theme system |
| **0.4.1** | 2026-01-29 | URL support for nvp apply |
| **0.4.0** | 2026-01-29 | nvp standalone CLI |
| **0.3.3** | 2026-01-29 | Pre-generated shell completions |
| **0.3.1** | 2026-01-29 | Multi-platform support |
| **0.2.0** | 2026-01-24 | Theme system + YAML highlighting |
| **0.1.0** | 2026-01-23 | Initial release |

---

## Links

- [Full Changelog](https://github.com/rmkohlman/devopsmaestro/blob/main/CHANGELOG.md)
- [GitHub Releases](https://github.com/rmkohlman/devopsmaestro/releases)
- [Release Process](./development/release-process.md)

---

## Semantic Versioning

DevOpsMaestro follows [Semantic Versioning](https://semver.org/):

- **MAJOR** (X.0.0) - Breaking changes
- **MINOR** (0.X.0) - New features (backward compatible)
- **PATCH** (0.0.X) - Bug fixes (backward compatible)
