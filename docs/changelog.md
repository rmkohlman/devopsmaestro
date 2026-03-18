# Changelog

All notable changes to DevOpsMaestro are documented in the [CHANGELOG.md](https://github.com/rmkohlman/devopsmaestro/blob/main/CHANGELOG.md) file in the repository.

## v0.53.0 (2026-03-17)

**✨ List Format YAML/JSON Export for `dvm get all`**

`dvm get all -o yaml` and `-o json` now produce a kubectl-style `kind: List` document with full resource YAML per item, replacing the previous lossy `AllResources` summary struct. Output can be piped directly to `dvm apply -f -` for backup and restore.

**⚠️ Breaking change:** The JSON/YAML output structure for `dvm get all` has changed. Scripts parsing the old `AllResources` format (with top-level `ecosystems`, `domains`, `apps`, etc. arrays) will break. The new output is a `kind: List` document where each item is the full resource YAML.

- **`kind: List` output** — full resource YAML per item (not lossy summaries); identical to `dvm get <resource> <name> -o yaml` output; all 13 resource types included (was 9)
- **Round-trip fidelity** — `dvm get all -A -o yaml | dvm apply -f -` exports and restores all resources; enables infrastructure-as-code backup/restore workflows
- **`dvm apply -f` supports `kind: List`** — apply pipeline detects List documents and applies each item individually with continue-on-error; enables `dvm apply -f backup.yaml`
- **Dependency-ordered output** — items in the List follow apply-safe dependency order: Ecosystems → Domains → Apps → GitRepos → Registries → Credentials → Workspaces → NvimPlugins → NvimThemes → NvimPackages → TerminalPrompts → TerminalPackages
- **Scoped export excludes globals** — when using `-e`/`-d`/`-a` flags with `-o yaml/json`, global resources (registries, git repos, nvim plugins/themes/packages, terminal prompts/packages) are excluded; only hierarchical resources matching the scope are exported; table output is unaffected
- **`metadata.domain` on Workspace YAML** — Workspace YAML now includes `metadata.domain` for context-free apply; handler resolves domain from this field first, falling back to active context
- **`ResourceList` type** — new `pkg/resource/list.go` with `BuildList()` and `ApplyList()` functions
- 1 new production file (`pkg/resource/list.go`); 2 modified production files (`cmd/get_all.go`, `cmd/apply.go`)

## v0.52.0 (2026-03-17)

**✨ Scoped Hierarchical Views in `dvm get all`**

`dvm get all` now scopes output to the active context by default and supports explicit scope flags for filtering to a specific ecosystem, domain, or app.

- **Scoped filtering** — scopes to the active context (ecosystem, domain, app) by default; `-e/--ecosystem`, `-d/--domain`, `-a/--app` flags filter to a specific scope; `-A/--all` shows everything regardless of context
- **kubectl-style scope resolution** — priority chain: `-A` > explicit flags > active context > show all (discovery mode); no flags and no active context shows all resources
- **Hierarchical cascade** — filtering by ecosystem shows only that ecosystem's domains, apps, workspaces, and scoped credentials; global resources (registries, git repos, nvim plugins, nvim themes) always shown
- **Validation** — `-A` combined with scope flags produces an error; `-d` without ecosystem and `-a` without domain produce helpful error messages with hints
- **Context helpers** — new `getActiveEcosystemFromContext()` and `getActiveDomainFromContext()` functions with env var override support (`DVM_ECOSYSTEM`/`DVM_DOMAIN`)
- 34 new tests (+ ~23 context helper tests); 0 breaking changes

## v0.51.0 (2026-03-17)

**✨ Rich Columns in `dvm get all`**

`dvm get all` now displays the same rich column structure as individual `dvm get <resource>` commands — active markers, section counts, resource-specific columns, and `-o wide` support.

- **Rich columns** — each section now shows the same headers as the dedicated list command: ECOSYSTEM under domains, DOMAIN under apps, APP under workspaces, PATH for apps, CREATED timestamps, and all other resource-specific columns; previously `dvm get all` used a reduced column set
- **Section counts** — section headers now include the resource count (e.g., `=== Ecosystems (3) ===`)
- **Active markers (`●`)** — active ecosystem, domain, app, and workspace rows are marked with `●` in the NAME column, matching individual `dvm get` output
- **`-o wide` support** — adds ID, GITREPO, CONTAINER-ID, SLUG, REF, AUTO_SYNC, CREATED columns across all sections
- **Registries full column set** — NAME, TYPE, VERSION, PORT, LIFECYCLE, STATE, UPTIME
- **Refactoring** — 9 inline table-building sections in `get_all.go` replaced with the shared `tableBuilder` implementations from `cmd/table.go`; `dvm get all` and individual `dvm get <resource>` commands now use identical table builders; 0 breaking changes

## v0.50.1 (2026-03-17)

**🐛 Fix FOREIGN KEY Constraint Failed on Delete**

Deleting a workspace, app, domain, or ecosystem that was set as the active context previously failed with `FOREIGN KEY constraint failed`. The `context` table tracked active resource IDs via FK columns with no `ON DELETE SET NULL`, so SQLite blocked any delete that the context row still referenced.

- **Migration 016** — Rebuilt the `context` table with `ON DELETE SET NULL` on all 4 FK references (`active_ecosystem_id`, `active_domain_id`, `active_app_id`, `active_workspace_id`); SQLite now auto-clears the active context when the referenced resource is deleted
- **Credential orphan cleanup** — `DeleteWorkspace`, `DeleteApp`, `DeleteDomain`, `DeleteEcosystem` now remove credentials scoped to the deleted resource (and its children) before deletion; credentials use a polymorphic `scope_type`/`scope_id` pattern with no SQL FK constraint and must be cleaned up application-side
- **`cmd/delete.go` UX fix** — active context check moved to before the delete so the "Cleared active workspace context" message is displayed correctly
- 8 new regression tests in `db/store_delete_context_test.go` covering delete-while-active for all 4 resource types plus credential cleanup; 0 breaking changes

## v0.50.0 (2026-03-17)

**🏗️ GitRepo Resource Handler + Shared Table Helpers**

Foundation sprint for the Get All Enhancement initiative — no new CLI commands or flags; all changes are internal infrastructure.

- **GitRepo resource handler** — GitRepo was the only resource type that bypassed the `pkg/resource/` framework; added `GitRepoHandler` implementing the full `resource.Handler` interface (Apply, Get, List, Delete, ToYAML), plus `GitRepoYAML` / `GitRepoMetadata` / `GitRepoSpec` model types and a `GitRepoResource` wrapper; enables future `dvm apply -f gitrepo.yaml` and `dvm get all -o yaml` support for GitRepo resources
- **Shared table-building helpers** — extracted the repeated table-building pattern (headers → iterate → rows → wide → render) from 12+ duplicate implementations across `cmd/*.go` into a single `cmd/table.go`; introduced the unexported `tableBuilder` interface, generic `BuildTable[T any]()` function, `renderTable()` helper, and utility functions (`truncateLeft`, `truncateRight`, `activeMarker`, `splitStatusUptime`); 9 builder structs cover all resource types (ecosystem, domain, app, workspace, credential, registry, gitrepo, nvim-plugin, nvim-theme)
- 77 new tests (18 GitRepo handler tests + 59 table builder tests); 2 new production files, 2 new test files, 2 modified production files; 0 breaking changes

## v0.49.0 (2026-03-17)

**✨ Auto-Detect Git Default Branch**

`dvm create gitrepo` and `dvm create app --repo <url>` now auto-detect the remote repository's default branch instead of hardcoding `main` — repos using `master`, `develop`, `trunk`, or any custom default branch now work correctly on workspace creation.

- **Auto-detected via `git ls-remote --symref`** — 10-second timeout; falls back to `main` if detection fails (missing git, network error, unresolvable URL)
- **`--default-ref` flag** — explicit override on `dvm create gitrepo` when auto-detection is not desired; e.g., `--default-ref develop`
- **`dvm create app --repo <url>` included** — the auto-create-GitRepo path now detects the default branch instead of hardcoding `main`
- 10 new test cases (8 parser unit tests + 2 CLI flag tests); 1 new production file, 2 modified; 0 breaking changes

## v0.48.0 (2026-03-17)

**✨ Python System Dependency Auto-Detection**

`dvm build` now automatically scans `requirements.txt` for Python packages that require C extension headers and injects the corresponding system libraries into the Dockerfile `apt-get install` command — no manual configuration needed for common packages.

- **Auto-detected at build time** — `detectPythonSystemDeps()` runs alongside the existing private-repo scan; no separate command required
- **11 supported packages** — `psycopg2` → `libpq-dev`; `mysqlclient` → `default-libmysqlclient-dev`; `pillow` → `libjpeg-dev zlib1g-dev libfreetype6-dev`; `lxml` → `libxml2-dev libxslt1-dev`; `cryptography`/`cffi` → `libffi-dev libssl-dev`; `pyyaml` → `libyaml-dev`; `python-ldap` → `libldap2-dev libsasl2-dev`; `gevent` → `libev-dev libevent-dev`; `pycairo` → `libcairo2-dev pkg-config`; `h5py` → `libhdf5-dev`
- **Binary wheels excluded** — `psycopg2-binary` and other pre-compiled variants are not matched; they ship without header requirements
- **PEP 503 normalization** — package names are lowercased and `[-_.]` collapsed before lookup; handles `PyYAML`, `Pillow`, `python_ldap`, etc. correctly
- **Dockerfile comment** — each auto-detected mapping is documented inline (e.g., `# Auto-detected: psycopg2 -> libpq-dev`); detected deps also logged via `render.Info()` during build
- **`baseStage.packages` YAML field** — manual escape hatch for packages not in the auto-detect map; merged and deduplicated alongside auto-detected packages
- 45 new test subtests across `utils/private_repo_detector_test.go` and `builders/dockerfile_generator_test.go`; 4 production files changed; 0 breaking changes

## Latest Releases

### v0.47.0 (2026-03-17)

**🏗️ Improved Credential Output**

`dvm get credentials`, `dvm get credential <name>`, and `dvm get all` now display richer, more readable credential information.

- **`dvm get credentials` list** — now renders a 5-column table (NAME, SCOPE, SOURCE, TARGET, DESCRIPTION) instead of plain-text `name  (scope: x, source: y)`; TARGET shows the env var(s) injected at build/attach time; DESCRIPTION shows the credential description
- **`dvm get credential <name>` scope display** — scope now resolves to a human-readable name (e.g., `app: my-api`) instead of a raw numeric ID (e.g., `app (ID: 5)`)
- **`dvm get all` credentials table** — updated from 3 columns (NAME, SCOPE, SOURCE) to 4 columns (NAME, SCOPE, SOURCE, TARGET)

### v0.46.0 (2026-03-17)

**✨ Terminal Tab Title on Attach**

`dvm attach` now sets the terminal tab/window title via OSC 0 escape sequences when attaching to a workspace.

- **Title set on attach** — writes `\x1b]0;[dvm] appName/workspaceName\x07` to stderr before the attach handoff; e.g., attaching to workspace `dev` in app `myapp` sets the title to `[dvm] myapp/dev`
- **Title reset on detach and error** — writes `\x1b]0;\x07` to stderr when the session exits or if attach returns an error; the tab title is never left in a stale state
- **Stderr convention** — terminal control sequences go to stderr to avoid interfering with piped stdout
- **Universal OSC 0 support** — works with any terminal that supports the xterm OSC 0 standard: WezTerm, iTerm2, Kitty, Alacritty, macOS Terminal.app, GNOME Terminal, Windows Terminal, etc. No terminal-specific configuration needed. WezTerm users can optionally customize the display via the `format-tab-title` event in `wezterm.lua`.
- 1 production file changed (`cmd/attach.go`); no new tests

### v0.45.6 (2026-03-17)

**🐛 Auto-Session Restore Fix**

Fixed two bugs in the Neovim plugin YAML configuration that prevented auto-session restore from working in containers.

- **`auto_restore_enabled` was `false`** (`pkg/nvimops/library/plugins/25-auto-session.yaml`) — `rmagatti/auto-session` never auto-restored the previous session on `VimEnter`; users had to manually press `<leader>wr` every time they entered the container. Changed to `true` — Neovim now auto-restores the last session for the current working directory on startup.
- **Alpha dashboard button used deprecated command** (`pkg/nvimops/library/plugins/15-alpha.yaml`) — the "Restore Session" button used `<cmd>SessionRestore<CR>` (old/deprecated) instead of `<cmd>AutoSession restore<CR>` (correct). Changed to match the keymap already in `25-auto-session.yaml`.
- No unit tests (YAML plugin configs verified via manual testing in container)

### v0.45.5 (2026-03-17)

**🐛 Pip Install Proxy Fallback**

Fixed pip install hanging and failing during Docker builds when dvm's Squid HTTP proxy registry is enabled but unreachable from inside the build context (`builders/dockerfile_generator.go`).

Root cause: Squid proxy env vars are set unconditionally in the build context. When the proxy is unreachable (e.g., `host.docker.internal:3128` doesn't resolve, firewall blocks the proxy port, or Squid crashed mid-build), pip has no built-in fallback — it retries for ~110 seconds then fails hard with `ProxyError('Cannot connect to proxy.', TimeoutError('_ssl.c:999: The handshake operation timed out'))`.

- **Fix** — all 5 pip install sites in the generated Dockerfile now use a `|| fallback` pattern: if pip install fails, it retries with proxy env vars unset (`unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy`), allowing direct PyPI access. Affected sites: `generateBaseStage()` default, "https", "ssh", and "mixed" cases, plus `installLanguageTools()` Python dev tools (ruff, mypy, etc.). Follows the same `|| fallback` strategy used for NodeSource in v0.45.3.
- 1 new test function (`TestPipInstall_ProxyFallback`) with 5 table-driven subtests covering all pip install sites

### v0.45.4 (2026-03-17)

**🐛 Mason Package Name Fix**

Fixed `MasonInstall` failure during container builds caused by a wrong package name in `builders/dockerfile_generator.go`.

Root cause: `getBaseMasonTools()` returned `"lua_ls"`, which is the nvim-lspconfig identifier. Mason's registry uses the hyphenated name `"lua-language-server"`. The generated `MasonInstall lua_ls` command failed immediately with `"lua_ls" is not a valid package`.

- **Fix** — changed `"lua_ls"` to `"lua-language-server"` in `getBaseMasonTools()`; the generated Dockerfile now emits `MasonInstall lua-language-server`, which resolves correctly against the Mason registry
- 1 new test function (`TestGetBaseMasonTools_UsesRegistryNames`) validates that no Mason package names contain underscores and explicitly checks for `lua-language-server`; 2 existing test functions updated with corrected assertion values

### v0.45.3 (2026-03-17)

**🐛 NodeSource Install Ordering and Fallback**

Fixed two bugs in Debian Node.js workspace builds (`builders/dockerfile_generator.go`):

- **Ordering bug** — The `RUN curl ... nodesource.com/setup_22.x | bash` step ran before the merged `apt-get install` that installs `curl`. On a fresh build without cache, `curl` was not yet available, producing `curl: not found` and aborting the build. Fixed by moving the NodeSource block to after the merged `apt-get install`.
- **No network fallback** — If NodeSource (`deb.nodesource.com`) was unreachable (corporate firewalls, Colima DNS issues), the build failed hard. Added a `|| apt-get install -y --no-install-recommends nodejs npm` fallback so the image build can succeed with Debian's default Node 18 when NodeSource is unavailable. Mason works with either version.
- 2 new test functions: `TestGenerateDevStage_DebianNodeSource_OrderAfterMergedInstall` and `TestGenerateDevStage_DebianNodeSource_Fallback`

### v0.45.2 (2026-03-17)

**🐛 IsRunning Health Probe Fallback**

Fixed `dvm get registries` showing "stopped" for Athens, Zot, and Devpi registries that were adopted as already-running by `Start()`.

Root cause: `Start()` adopted a running instance (health probe succeeded on a busy port) and returned `nil` without writing a PID file. A fresh `RegistryManager` (created on every `dvm get registries` call) then called `IsRunning()` → read PID file → not found → returned `false` → displayed "stopped".

- **`IsRunning()` overridden on all 3 managers** — adds a health probe fallback when the PID file check returns false; reuses the existing `ProbeServiceHealth()` from `utils.go` with the same endpoints and accepted status codes as the adoption path in `Start()`
- **Health endpoints** — Athens: `GET /healthz` → 200; Zot: `GET /v2/` → 200 or 401; Devpi: `GET /` → 200 or 302
- 8 new test functions across 3 test files

### v0.45.1 (2026-03-16)

**🐛 Zot Checksum Manifest Parsing Fix**

The v0.45.0 manifest parser compared the raw entry against the plain filename, but `sha256sum` binary-mode prefixes filenames with `*` (e.g., `*zot-darwin-arm64`). Added `strings.TrimPrefix` to strip the `*` before comparison.

### v0.45.0 (2026-03-16)

**🐛 Registry Startup Resilience**

Fixed 3 bugs that caused only 2 of 5 registries (squid and verdaccio) to start successfully. Athens, Zot, and Devpi all failed to start when they were already running because `Start()` returned `ErrPortInUse` without checking whether the service was healthy.

- **Port-in-use probe** (Athens, Zot, Devpi) — `Start()` now calls `ProbeServiceHealth()` before returning `ErrPortInUse`; if the service is healthy, it is adopted as already-running. Health endpoints: Athens `GET /healthz` → 200; Zot `GET /v2/` → 200 or 401; Devpi `GET /` → 200 or 302.
- **Zot checksum URL** — `fetchChecksum()` was appending `.sha256` directly to the binary URL (404). Rewrote to fetch the `{baseURL}/checksums.sha256.txt` manifest and parse the matching filename entry.
- **Devpi pip fallback** — `ensurePipxInstalled()` had no fallback when `pipx` was absent. Added `fallbackPipInstall()` using `python3 -m pip install --user devpi-server==6.2.0` and a `getPythonUserBase()` helper.
- 11 new test functions across 5 test files covering all three fixes.

### v0.44.0 (2026-03-16)

**🐛 Container Neovim Environment Fixes**

Fixed 6 cascading bugs that caused Neovim failures inside workspace containers. The root cause chain: Node.js 18 (too old) → Mason can't install some tools → pylint missing → BufEnter ENOENT error → auto-session restore fails.

- **Node 22 on Debian** — Debian/Node.js builds now install Node 22 from NodeSource; Alpine builds are unchanged. `effectiveVersion()` default for `nodejs` updated from `"18"` to `"20"`.
- **Mason LSP config `ensure_installed` removed** (`06-mason.yaml`) — the hardcoded list of 19 language servers is gone; Mason is now a framework only; server installation is driven by `getMasonToolsForLanguage()` at build time.
- **Mason tool installer `ensure_installed` removed** (`06-mason.yaml` + `builders/dockerfile_generator.go`) — the hardcoded list of 9 tools is gone; `getMasonToolsForLanguage()` + `getBaseMasonTools()` + `installMasonTools()` are the single authority for build-time Mason tool installation.
- **Python tools expanded** — `pylint` added to the Python Mason tool list; **Go tools expanded** — `goimports` added to the Go Mason tool list.
- **Defensive linting guards** (`24-linting.yaml`) — `vim.fn.executable()` checks added for `shellcheck` and `pylint`; missing binaries produce no ENOENT error.
- **auto-session restore** — resolved automatically once the cascade root cause (Node 18) was fixed.
- Plugin YAML `06-mason.yaml` reduced from 64 to 34 lines. 6 new test functions (9 subtests); 2 pre-existing tests updated (node:18 → node:20).

### v0.43.2 (2026-03-16)

**🔒 Build Output Secret Redaction**

Package managers (`pip install`, `npm install`, `go get`) print build URLs in plain text during Docker/BuildKit builds. When credentials are passed as `--build-arg` values, those values appeared verbatim in terminal output (e.g., `Downloading https://ghp_WJ0M3T...@github.com/org/repo/archive/main.tar.gz`). A new `RedactingWriter` intercepts all build output and replaces known credential values with `***` before they reach the terminal.

- **`builders/redacting_writer.go`** (new) — `RedactingWriter` type wrapping any `io.Writer`; cross-boundary buffering catches secrets split across chunk boundaries; longest-first secret matching prevents partial-match artifacts
- **Minimum secret length: 8 characters** — avoids false positives on version strings and short flags; values shorter than 8 bytes are not redacted
- **Zero-overhead fast path** — when no qualifying secrets exist, `NewRedactingWriter` returns the inner writer directly; no wrapping overhead
- **`docker_builder.go`** — `cmd.Stdout`/`cmd.Stderr` both wrapped with `NewRedactingWriter`; `defer Flush()` drains buffer on all exit paths
- **`buildkit_builder.go`** — `progressui.NewDisplay` writer wrapped with `NewRedactingWriter`; `Flush()` called after progress completes
- 15 new tests in `builders/redacting_writer_test.go` (basic redaction, multiple secrets, split-across-writes, min-length boundary, pip output simulation)

Note: This is Layer 1 (immediate mitigation). Layer 2 (`--mount=type=secret` for BuildKit) is tracked for a future release.

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
| **0.45.6** | 2026-03-17 | Auto-session restore fix — `auto_restore_enabled` set to `true`; Alpha dashboard "Restore Session" button updated from deprecated `SessionRestore` to `AutoSession restore` |
| **0.45.5** | 2026-03-17 | Pip install proxy fallback — `ProxyError` / ~110 s hang fix when Squid proxy unreachable; all 5 pip install sites now retry with proxy env vars unset |
| **0.45.3** | 2026-03-17 | NodeSource install ordering and fallback — `curl: not found` fix (NodeSource moved after merged apt-get install); `\|\|` fallback to Debian default nodejs when NodeSource unreachable |
| **0.45.2** | 2026-03-17 | IsRunning health probe fallback — `dvm get registries` no longer shows "stopped" for adopted Athens/Zot/Devpi instances; `IsRunning()` probes health endpoint when PID file is absent |
| **0.45.0** | 2026-03-16 | Registry startup resilience — port-in-use probe for Athens/Zot/Devpi, Zot checksum URL rewrite, Devpi pip fallback |
| **0.44.0** | 2026-03-16 | Container Neovim environment fixes — Node 22 on Debian, Mason ensure_installed removed from plugin YAML, build-time tool authority centralized, pylint/shellcheck executable guards |
| **0.43.2** | 2026-03-16 | Build output secret redaction — `RedactingWriter` intercepts `pip`/`npm`/`go get` output; replaces credential values with `***`; cross-boundary buffering; zero-overhead fast path |
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
