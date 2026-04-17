# Changelog

All notable changes to DevOpsMaestro are documented in the [CHANGELOG.md](https://github.com/rmkohlman/devopsmaestro/blob/main/CHANGELOG.md) file in the repository.

## v0.101.5 (2026-04-16)

**Bug Fixes**
- **BuildKit cache export/import fails with HTTPS→HTTP protocol mismatch against local zot registry** — `cmd/build_phases.go` adds `registry.insecure=true` to both the cache-from and cache-to BuildKit format strings, allowing BuildKit to communicate with the local zot registry over HTTP ([#388](https://github.com/rmkohlman/devopsmaestro/issues/388))

---

## v0.101.4 (2026-04-16)

**Bug Fixes**
- **Registry state detection broken due to stale PID files — connect-based port check and stale PID cleanup** — `pkg/registry/utils.go` changes `IsPortAvailable` from a bind-based check to a connect-based check, correctly detecting whether a port is in use by an active listener; `pkg/registry/process_manager.go` adds stale PID file cleanup so dead processes no longer cause false "already running" reports ([#387](https://github.com/rmkohlman/devopsmaestro/issues/387))

**Tests**
- Updated `pkg/registry/utils_test.go` — test cases updated to reflect the new connect-based `IsPortAvailable` behavior ([#387](https://github.com/rmkohlman/devopsmaestro/issues/387))

---

## v0.101.3 (2026-04-16)

**Bug Fixes**
- **zot networking: BuildKit uses host.docker.internal for VM-accessible registry URLs** — `pkg/registry/buildkit_config.go` and `pkg/registry/build_support.go` updated so BuildKit resolves the zot registry via `host.docker.internal` instead of `localhost`, fixing connectivity inside Docker VMs ([#386](https://github.com/rmkohlman/devopsmaestro/issues/386))
- **Squid state detection: port-based fallback when PID file check fails** — `pkg/registry/squid_manager.go` now falls back to a port liveness check when the PID file is missing or stale, preventing false "not running" reports ([#386](https://github.com/rmkohlman/devopsmaestro/issues/386))
- **Verdaccio state detection: port-based fallback when PID file check fails** — `pkg/registry/verdaccio_manager.go` applies the same port-based fallback as Squid for consistent state detection ([#386](https://github.com/rmkohlman/devopsmaestro/issues/386))

**Tests**
- Added `pkg/registry/issue386_test.go` — tests covering zot `host.docker.internal` URL generation and port-based state detection fallback ([#386](https://github.com/rmkohlman/devopsmaestro/issues/386))

---

## v0.101.2 (2026-04-16)

**Bug Fixes**
- **BuildKit crashes with EOF errors during parallel builds — add retry logic and registry proxy resilience** — `builders/build_errors.go` adds BuildKit connection error detection; `cmd/build_phases.go` adds retry logic for BuildKit EOF crashes; `builders/buildkit_builder.go` implements graceful cache import degradation; `pkg/registry/process_manager.go` fixes stale PID race; `pkg/registry/devpi_manager.go` and `pkg/registry/verdaccio_manager.go` add graceful "already running" handling ([#385](https://github.com/rmkohlman/devopsmaestro/issues/385))

**Tests**
- Added `builders/build_errors_test.go` — tests for BuildKit connection error detection ([#385](https://github.com/rmkohlman/devopsmaestro/issues/385))
- Added `builders/buildkit_builder_test.go` — tests for registry reachability ([#385](https://github.com/rmkohlman/devopsmaestro/issues/385))

---

## v0.101.1 (2026-04-16)

**Bug Fixes**
- **Registry cache not leveraged during builds — fix cache ref format and startup race condition** — `cmd/build_phases.go` strips the `http://` prefix from registry endpoints so the `type=registry` cache ref is correctly formatted and accepted by BuildKit; adds post-build light cache pruning; `pkg/registry/verdaccio_manager.go` fixes a startup race condition with a health probe; `builders/buildkit_builder.go` adds `PruneBuildKitCacheLight()` for targeted post-build pruning ([#384](https://github.com/rmkohlman/devopsmaestro/issues/384))

---

## v0.101.0 (2026-04-16)

**Features**
- **`--clean-cache` now aggressively leverages registries for Docker layer caching and minimises Colima/BuildKit disk footprint** — new pre/post build cleanup helpers prune BuildKit cache, delete old workspace images, and prune dangling images; registry layer cache (`type=registry`) is injected into `CacheFrom`/`CacheTo` when a local zot registry is available; both BuildKit and docker CLI now support multiple cache sources via newline-separated `CacheFrom` values ([#383](https://github.com/rmkohlman/devopsmaestro/issues/383))

**Tests**
- Added `build_clean_cache_test.go` — 11 tests covering nil-platform guards, registry ref format, multi-value `CacheFrom` splitting, and local+registry cache combination ([#383](https://github.com/rmkohlman/devopsmaestro/issues/383))

---

## v0.100.5 (2026-04-16)

**Bug Fixes**
- **Lazygit "Git version must be at least 2.32.0" — install git unconditionally from backports** — `builders/dockerfile_generator.go` now adds the Debian backports apt source before the merged package install and pins git to the backports target, guaranteeing git >= 2.32.0 is always installed without a conditional post-install step; replaces the version-check approach from #380 with a simpler, more reliable unconditional install ([#382](https://github.com/rmkohlman/devopsmaestro/issues/382))

---

## v0.100.4 (2026-04-16)

**Bug Fixes**
- **Lazygit inside container fails with "Git version must be 2.32.0 please upgrade your git version"** — `builders/dockerfile_generator.go` now conditionally upgrades git from Debian backports when the installed version is below 2.32.0; older Debian base images (e.g., `python:3.9-slim` / Bullseye) ship git 2.30 which is too old for lazygit v0.60+; Alpine images are unaffected ([#380](https://github.com/rmkohlman/devopsmaestro/issues/380))

---

## v0.100.3 (2026-04-16)

**Bug Fixes**
- **Lazygit inside container tries to connect to remote git server instead of using local bare git repositories** — `cmd/attach.go` now mounts git mirror volumes and rewrites remotes so bare git repos are accessible inside containers, allowing lazygit to function correctly without network access ([#379](https://github.com/rmkohlman/devopsmaestro/issues/379))

---

## v0.100.2 (2026-04-16)

**Bug Fixes**
- **Auto-detect and recover from BuildKit cache corruption** — `BuildKitBuilder` now detects BuildKit cache corruption errors during build and automatically clears the cache and retries; `build_phases.go` surfaces actionable diagnostics when corruption is detected; `build_errors.go` expanded with BuildKit-specific error classification and recovery guidance ([#378](https://github.com/rmkohlman/devopsmaestro/issues/378))

**Tests**
- Added `build_errors_test.go` coverage for BuildKit cache corruption detection and error classification ([#378](https://github.com/rmkohlman/devopsmaestro/issues/378))

---

## v0.100.1 (2026-04-16)

**Bug Fixes**
- **Squid proxy cache corruption recovery causing build failures with network timeouts** — added cache corruption detection and recovery with automatic retry logic in `squid_manager.go`; enhanced build error messages in new `build_errors.go` module to surface actionable diagnostics for network timeout failures; `BuildKitBuilder` and `DockerBuilder` now use `EnhanceBuildError` for consistent error enrichment ([#377](https://github.com/rmkohlman/devopsmaestro/issues/377))

**Tests**
- Added `squid_cache_recovery_test.go` — covers cache corruption detection, recovery, and retry logic ([#377](https://github.com/rmkohlman/devopsmaestro/issues/377))
- Added `build_errors_test.go` — covers enhanced build error message generation for network timeout scenarios ([#377](https://github.com/rmkohlman/devopsmaestro/issues/377))

---

## v0.100.0 (2026-04-16)

**Features**
- **Auto-prune old workspace images after successful rebuild** — after every successful `dvm build`, automatically removes all older tagged images for the workspace, keeping only the freshly-built image; prune failures are non-fatal and never block the build result ([#376](https://github.com/rmkohlman/devopsmaestro/issues/376))

---

## v0.99.4 (2026-04-15)

**Bug Fixes**
- **`dvm build status` showed stale "running" state and incorrect counters after build completion** — recompute succeeded/failed counters from `GetBuildSessionStats` (source of truth) instead of denormalized session fields; detect sessions stuck in "running" when all workspace entries are terminal and derive corrected status (`completed`/`partial`/`failed`) from actual workspace data; self-heal by persisting corrected state via `UpdateBuildSession`; estimate duration from latest workspace `CompletedAt` when session `CompletedAt` is missing due to interrupted finalization ([#375](https://github.com/rmkohlman/devopsmaestro/issues/375))

---

## v0.99.3 (2026-04-15)

**Bug Fixes**
- **Git credential mounting for workspace containers** — mounts `~/.ssh` (read-only) and `~/.gitconfig` (read-only) into workspace containers at runtime; developers can now use their host SSH keys and git identity inside containers without manual setup or credential copy ([#374](https://github.com/rmkohlman/devopsmaestro/issues/374))

---

## v0.99.2 (2026-04-15)

**Bug Fixes**
- **Squid registry incorrectly reported as stopped when running** — set `pid_filename none` in the generated `squid.conf` template; squid running in foreground mode (`-N`) would otherwise skip writing or actively remove the PID file that dvm's `ProcessManager` writes, causing `dvm` to report the process as `stopped` even though it was running ([#373](https://github.com/rmkohlman/devopsmaestro/issues/373))

---

## v0.99.1 (2026-04-15)

**Bug Fixes**
- **Squid proxy silent fallback on fresh machine** — added `IsAvailable()` pre-flight check to `BrewBinaryManager`; builds now emit a clear, prominent actionable message ("Squid proxy not available. Install with: `brew install squid` for faster builds") instead of silently continuing without proxy caching ([#371](https://github.com/rmkohlman/devopsmaestro/issues/371))
- **Distinguish binary-not-installed from binary-failed-to-start** — introduced `ErrBinaryNotInstalled` sentinel error; `Prepare()` now distinguishes missing-binary errors from binary-present-but-failed-to-start errors with different log levels and messages ([#371](https://github.com/rmkohlman/devopsmaestro/issues/371))
- **CacheReadiness.Unhealthy warnings more prominent** — elevated registry warning rendering levels in build output; added emoji and color to `FormatSummary()` so proxy cache failures are visually distinct ([#371](https://github.com/rmkohlman/devopsmaestro/issues/371))

---

## v0.99.0 (2026-04-15)

**Bug Fixes**
- **CRITICAL: Build state persistence** — workspace image tags are now written to the database after a successful build; previously a successful build left workspaces in `:pending` state, making them unattachable via `dvm attach` ([#367](https://github.com/rmkohlman/devopsmaestro/issues/367))
- **Squid xcalloc integer overflow** — reduced L2 cache `cache_dir` directive from 256 to 64 MB; a 256 MB L2 allocation triggered an xcalloc integer overflow on some systems, causing a fatal Squid crash at startup ([#363](https://github.com/rmkohlman/devopsmaestro/issues/363))
- **Neovim GLIBC fallback on older base images** — added a runtime GLIBC version check; if the base image's glibc is too old for the pre-built Neovim binary, the build automatically falls back to compiling Neovim from source via CMake ([#342](https://github.com/rmkohlman/devopsmaestro/issues/342))
- **Session state logging** — persistence failures during build sessions now emit structured error logs with full context, improving visibility into state-write errors ([#366](https://github.com/rmkohlman/devopsmaestro/issues/366))
- **Mason poll loop** — replaced the fixed 5-second sleep with a 30-second poll loop that checks for Mason package installation completion; eliminates false success/failure counts caused by fixed sleep timing ([#365](https://github.com/rmkohlman/devopsmaestro/issues/365))
- **Zot PID race condition** — added `os.IsNotExist()` guards around PID file read/unlink operations during Zot registry startup; prevents `ENOENT` errors when concurrent goroutines race to clean up or check the PID file ([#364](https://github.com/rmkohlman/devopsmaestro/issues/364))

---

## v0.98.0 (2026-04-15)

**Bug Fixes**
- **Fix APT timeout accumulation** — reduced per-package APT install timeout from 60s to 10s with proxy health check fail-fast; prevents multi-package installs from stalling for minutes on a dead proxy ([#354](https://github.com/rmkohlman/devopsmaestro/issues/354))
- **Fix `--timeout` flag not respected** — user-supplied timeout is now plumbed through from the CLI flag down to the build watchdog; previously the watchdog always used the default timeout regardless of what was passed ([#252](https://github.com/rmkohlman/devopsmaestro/issues/252))
- **Add auto-sync of embedded library to DB at build time** — library contents are now fingerprinted via SHA-256 and synced to the database automatically on every `dvm build`; eliminates stale-library errors after upgrades ([#255](https://github.com/rmkohlman/devopsmaestro/issues/255))
- **Scope filters now auto-build all matching workspaces** — `--ecosystem`, `--domain`, and `--app` filters on `dvm build` now resolve and build all matching workspaces in batch ([#215](https://github.com/rmkohlman/devopsmaestro/issues/215))

**Added**
- **Theme discoverability docs** — new quick-start guide `docs/nvp/quick-start-themes.md` covering `dvm library get themes`, `dvm library describe theme <name>`, `dvm set theme` at all hierarchy levels, cascade visualization, and common workflows ([#191](https://github.com/rmkohlman/devopsmaestro/issues/191))
- **Built-in themes reference** — rewrote `docs/nvp/themes.md` with complete catalog of all 34+ built-in themes organized by family, discovery commands, and custom theme creation ([#191](https://github.com/rmkohlman/devopsmaestro/issues/191))
- **Built-in packages reference** — rewrote `docs/nvp/packages.md` with built-in package catalog (`core`, `lazyvim`, `maestro-python`, `maestro-go`), full 38+ plugin library listing, discovery commands, and custom package authoring ([#191](https://github.com/rmkohlman/devopsmaestro/issues/191))

---

## v0.97.0 (2026-04-14)

**Bug Fixes**
- **Fix Mason package installation race condition** — packages now fully finalize before Neovim exits; removed `+qa` from the Mason nvim command, added a 5-second settle period and an explicit `vim.cmd('qa!')` call in the Lua install script ([#358](https://github.com/rmkohlman/devopsmaestro/issues/358))
- **Fix image copy failure under concurrency** — temp files now use unique slug+UUID names instead of a shared PID-based name, eliminating collisions when multiple workspaces copy images concurrently ([#359](https://github.com/rmkohlman/devopsmaestro/issues/359))

**Tests**
- Added `TestMasonInstallation_SettlePeriodAndQuit` — asserts settle period and `vim.cmd('qa!')` are present and that `+qa` is absent from the Mason nvim invocation ([#358](https://github.com/rmkohlman/devopsmaestro/issues/358))
- Added `TestBuildTarName_UniquePerWorkspace`, `TestBuildTarName_ContainsSlug`, and `TestBuildTarName_NoSharedPID` — assert that temp tar files use unique slug+UUID names and do not collide across concurrent builds ([#359](https://github.com/rmkohlman/devopsmaestro/issues/359))

---

## v0.96.3 (2026-04-14)

**Bug Fixes**
- **Neovim switched from AppImage to tarball download** — `neovim-builder` stage now downloads the official Neovim Linux tarball and extracts it with `tar xzf --strip-components=1`; removes the `squashfs-tools` dependency and `unsquashfs` extraction step; binary path is now `/opt/nvim/bin/nvim` instead of `/opt/nvim/usr/bin/nvim` ([#356](https://github.com/rmkohlman/devopsmaestro/issues/356))

**Tests**
- Updated `TestNeovimInstallation_PythonSlim` and `TestDockerfileGenerator` to assert tarball download and `tar xzf` extraction; binary path assertions updated to `/opt/nvim/bin/nvim`; asserts that `squashfs-tools` and `unsquashfs` are absent ([#356](https://github.com/rmkohlman/devopsmaestro/issues/356))

---

## v0.96.2 (2026-04-14)

**Bug Fixes**
- **Squid proxy binding**: Use explicit `0.0.0.0` binding address instead of ambiguous bare port form that defaulted to localhost on some platforms ([#353](https://github.com/rmkohlman/devopsmaestro/issues/353))
- **Stale proxy config**: Detect and restart squid processes running with outdated configuration from previous dvm versions ([#353](https://github.com/rmkohlman/devopsmaestro/issues/353))

---

## v0.96.1 (2026-04-14)

**Bug Fixes**
- **Neovim AppImage extraction fixed on ARM64** — replaced direct AppImage execution (`--appimage-extract`) with `unsquashfs` extraction; ARM64 AppImages fail with exit code 127 in minimal containers (`debian:bookworm-slim`) that lack FUSE and compatible dynamic linkers; `squashfs-tools` is now installed in the `neovim-builder` stage and the squashfs payload offset is detected via the `hsqs` magic bytes before extraction ([#351](https://github.com/rmkohlman/devopsmaestro/issues/351))

**Tests**
- Updated `TestNeovimInstallation_PythonSlim` to assert `unsquashfs` extraction and `squashfs-tools` installation, and assert that `--appimage-extract` is absent; added `does_not_execute_appimage_directly` and `detects_squashfs_offset` sub-tests ([#351](https://github.com/rmkohlman/devopsmaestro/issues/351))

---

## v0.96.0 (2026-04-14)

**Bug Fixes**
- **Squid proxy binding fixed for container network access** — squid now listens on `0.0.0.0` instead of `127.0.0.1`; BuildKit containers connect via `host.docker.internal` which resolves to the VM gateway IP (e.g. `192.168.5.2`), not loopback; RFC1918 ACL comment documents the security model ([#346](https://github.com/rmkohlman/devopsmaestro/issues/346))
- **Default build timeout increased from 30m to 45m** — reduces timeout-related build failures on slower machines or large workspaces; per-workspace timeout is now logged at parallel build start for visibility ([#347](https://github.com/rmkohlman/devopsmaestro/issues/347))
- **tree-sitter-cli switched to pre-built GitHub binary on Debian** — eliminates the Rust/Cargo compilation step (~85s) by downloading the pre-built binary from GitHub releases with SHA256 checksum verification; Alpine still builds from source via Cargo since pre-built binaries require glibc ([#348](https://github.com/rmkohlman/devopsmaestro/issues/348))
- **Neovim AppImage extraction validated** — confirmed already correct; no changes needed ([#349](https://github.com/rmkohlman/devopsmaestro/issues/349))

**Tests**
- Updated tree-sitter builder tests to assert pre-built binary download path for Debian and cargo-based path for Alpine ([#348](https://github.com/rmkohlman/devopsmaestro/issues/348))
- Updated squid manager tests to reflect `0.0.0.0` binding and RFC1918 ACL comment ([#346](https://github.com/rmkohlman/devopsmaestro/issues/346))
- Updated timeout flag test to reflect new 45m default ([#347](https://github.com/rmkohlman/devopsmaestro/issues/347))

---

## v0.95.0 (2026-04-14)

**New Features**
- **Ruby language support** — `generateBaseStage` now supports `ruby:{ver}-slim` as the base image; Gemfile/bundler dependency installation is generated automatically when a Gemfile is present ([#344](https://github.com/rmkohlman/devopsmaestro/issues/344))
- **Neovim uses AppImage extraction** — Neovim is now installed via AppImage extraction (`--appimage-extract`) rather than a pre-built binary download, eliminating GLIBC incompatibility across all base images and architectures ([#342](https://github.com/rmkohlman/devopsmaestro/issues/342))

**Bug Fixes**
- **Build summary reports incorrect success/failure counts** — the build summary now correctly counts successes and failures using unique `app/workspace` identifiers, preventing double-counting across retries and reporting the accurate totals ([#343](https://github.com/rmkohlman/devopsmaestro/issues/343))
- **Deprecated `ruff-lsp` in Mason Python tools** — replaced `ruff-lsp` with `ruff` in the Mason Python tool list; `ruff-lsp` was removed from Mason in favour of the bundled LSP in `ruff` itself ([#326](https://github.com/rmkohlman/devopsmaestro/issues/326))
- **Mason verification uses wrong package count** — Mason verification now counts executable entries in `mason/bin/` instead of the top-level `mason/` directory, giving an accurate installed-package count that matches what Mason actually reports ([#327](https://github.com/rmkohlman/devopsmaestro/issues/327))
- **Squid `dns_v4_first` directive removed** — the obsolete `dns_v4_first on` directive (removed in Squid 6) is no longer emitted; additionally, the PID file adoption error for already-running Squid processes is now suppressed via `--foreground` startup ([#335](https://github.com/rmkohlman/devopsmaestro/issues/335))
- **Verdaccio upgraded from deprecated v5.28.0 to v6.1.2** — Verdaccio is now pinned to v6.1.2; the install step runs `npm uninstall -g verdaccio` before installing the new version to clear stale v5 artifacts that caused startup failures ([#336](https://github.com/rmkohlman/devopsmaestro/issues/336))

**Tests**
- **Failed-build image tag propagation** — new regression tests assert that build results (including image tags) are preserved on the failure path and correctly surfaced by `dvm build status` ([#339](https://github.com/rmkohlman/devopsmaestro/issues/339))
- **Single-workspace session creation** — new tests verify that a build session is created and torn down correctly when only one workspace is in scope ([#339](https://github.com/rmkohlman/devopsmaestro/issues/339))
- **Ruby LanguageVersionTable and isAlpine consistency** — new table-driven tests cover Ruby version/image resolution and assert `isAlpine` returns correct results for all supported base image patterns ([#316](https://github.com/rmkohlman/devopsmaestro/issues/316))
- **Neovim AppImage extraction** — new tests assert the AppImage extraction path is used (and the old binary-download path is absent) across both Alpine and Debian Neovim builder variants ([#342](https://github.com/rmkohlman/devopsmaestro/issues/342))

---

## v0.94.1 (2026-04-14)

**Bug Fixes**
- **tree-sitter builder uses hardcoded `/root/.cargo` path** — replaced `/root/.cargo/bin/tree-sitter` with `$CARGO_HOME/bin/tree-sitter` in both the Alpine (`rust:1-alpine3.20`) and Debian (`rust:1-slim-bookworm`) tree-sitter builder stages; the fix respects non-root and custom Cargo installs where `CARGO_HOME` differs from `/root/.cargo` ([#338](https://github.com/rmkohlman/devopsmaestro/issues/338))

**Tests**
- Added `TestGenerateTreeSitterBuilder_UsesCargoHomeEnvVar` — regression test asserting `$CARGO_HOME/bin/tree-sitter` is used (and `/root/.cargo` is absent) for both Alpine and Debian builder variants ([#338](https://github.com/rmkohlman/devopsmaestro/issues/338))

---

## v0.94.0 (2026-04-14)

**Bug Fixes**
- **Go tools builder enforces Go >= 1.25 minimum** — The Go tools builder stage now enforces a minimum Go version of 1.25 for gopls and delve compatibility, regardless of the project's configured Go version. Older project versions are silently upgraded to 1.25 to prevent toolchain incompatibilities ([#220](https://github.com/rmkohlman/devopsmaestro/issues/220))
- **Tab completion broken after running dvm commands** — Fixed zsh completion by switching to the correct `autoload -Uz compinit && compinit` format and adding a guard to skip database initialization for `__complete` subcommands, preventing side-effects during shell completion ([#292](https://github.com/rmkohlman/devopsmaestro/issues/292))
- **`dvm build status` shows stale image tags** — Fixed tag propagation so build results (including tags) are preserved on failure paths, and added a session lifecycle for single-workspace builds so the latest tag is always reflected in `dvm build status` output ([#323](https://github.com/rmkohlman/devopsmaestro/issues/323))
- **Concurrent builds race on staging directories** — Builds for apps sharing a workspace no longer collide on staging directories; each build run now appends a UUID suffix to its staging path, providing full isolation under concurrent load ([#256](https://github.com/rmkohlman/devopsmaestro/issues/256))

**New Features**
- **`dvm system` maintenance commands (Phase 1)** — New top-level command group for system inspection and cleanup ([#257](https://github.com/rmkohlman/devopsmaestro/issues/257)):
  - `dvm system info` — displays platform details, current version, and disk usage overview
  - `dvm system df` — Docker-style breakdown of resource usage (BuildKit cache, workspace images, volumes)
  - `dvm system prune` — cleans up BuildKit cache and unused workspace images; supports `--dry-run`, `--force`, `--buildkit`, `--images`, and `--all` flags

---

## v0.93.0 (2026-04-14)

**Bug Fixes**
- **BuildKit cache mounts exhaust disk during concurrent builds** — `cacheID()` now derives the cache mount ID from `workspace.Slug`, guaranteeing globally unique cache IDs across parallel builds and preventing cross-workspace collisions that filled the disk ([#332](https://github.com/rmkohlman/devopsmaestro/issues/332))
- **APT GPG signature corruption in generated Dockerfiles** — added `rm -rf /var/lib/apt/lists/*` before every `apt-get update` to ensure a clean package list state and eliminate stale or corrupt signature errors ([#333](https://github.com/rmkohlman/devopsmaestro/issues/333))
- **tree-sitter GLIBC 2.39 incompatibility** — replaced pre-built `tree-sitter-cli` binary download with a Cargo-based source build (`cargo install tree-sitter-cli@0.24.7`), resolving GLIBC version mismatch on Ubuntu 22.04 and earlier ([#334](https://github.com/rmkohlman/devopsmaestro/issues/334))

**Enhancements**
- Added `sharing=locked` to the Neovim cache mount for consistency with all other BuildKit cache mounts

**Tests**
- 11 new tests covering cache ID isolation, APT list cleanup, and cargo-based tree-sitter installation

---

## v0.92.0 (2026-04-14)

**Bug Fixes**
- Fix ENOSPC during parallel Docker builds — added BuildKit cache mounts (`--mount=type=cache`) for Neovim Lazy sync, Treesitter parser compilation, and Mason tool installation ([#251](https://github.com/rmkohlman/devopsmaestro/issues/251))
- Fix APT lists cache path inconsistency in locked cache mounts (`/var/lib/apt` → `/var/lib/apt/lists`) ([#321](https://github.com/rmkohlman/devopsmaestro/issues/321))
- Replace deprecated `apt-key add` with modern `signed-by` GPG keyring pattern for Scala/sbt ([#321](https://github.com/rmkohlman/devopsmaestro/issues/321))

**Enhancements**
- Bump default `--concurrency` from 4 to 8, with adaptive upper bound guard at 2× CPU cores ([#330](https://github.com/rmkohlman/devopsmaestro/issues/330))

---

## v0.91.0 (2026-04-14)

### Features
- Integrate Zot registry as BuildKit mirror for faster image pulls (#328)
- BuildKit now configured to use Zot as a pull-through cache for docker.io and ghcr.io
- Containerd hosts.toml generated for nerdctl pull-through cache
- Buildx builder manager with hash-based config change detection

### Fixes
- Process manager now adopts already-running registry processes (no more 'address in use' errors)

---

## v0.90.1 (2026-04-14)

**Bug Fixes**

- **Python base image tag corrected for older versions** — Changed Python base image from `python:<ver>-slim-bookworm` to `python:<ver>-slim` to fix build failures with Python 3.9.x and other older versions where the `slim-bookworm` variant tag does not exist ([#324](https://github.com/rmkohlman/devopsmaestro/issues/324)).

---

## v0.90.0 (2026-04-14)

**New Features**

- **UV as default Python package manager in generated Dockerfiles** — `ghcr.io/astral-sh/uv:0.7.2` is now installed in Python Docker builds, providing up to 10x faster dependency installation compared to pip ([#273](https://github.com/rmkohlman/devopsmaestro/issues/273)).

- **3-tier UV fallback strategy** — Python dependency installation uses a robust fallback chain: `uv pip install` → `uv pip install` (no proxy) → `pip install`, ensuring compatibility across all network environments ([#273](https://github.com/rmkohlman/devopsmaestro/issues/273)).

- **UV environment variables** — Generated Dockerfiles set `UV_LINK_MODE=copy`, `UV_COMPILE_BYTECODE=1`, and `UV_SYSTEM_PYTHON=1` for optimal performance and compatibility in containerized environments ([#273](https://github.com/rmkohlman/devopsmaestro/issues/273)).

- **UV cache mount** — Docker layer cache mount updated from `/root/.cache/pip` to `/root/.cache/uv` to maximize cache reuse with UV-based builds ([#273](https://github.com/rmkohlman/devopsmaestro/issues/273)).

- **Python and Jupyter sandbox presets updated for UV** — Both the `python` and `jupyter` sandbox presets now use UV for dependency installation, consistent with Dockerfile generation ([#273](https://github.com/rmkohlman/devopsmaestro/issues/273)).

---

## v0.89.0 (2026-04-14)

**New Features**

- **euporie.nvim plugin for Jupyter notebook support** — New plugin (`54-euporie.yaml`) added to the MaestroNvim plugin library. Lazy-loads on `.ipynb` files, provides an Alt+n toggle for a floating terminal window running `euporie-notebook`, and auto-opens `.ipynb` buffers directly in Euporie ([#284](https://github.com/rmkohlman/devopsmaestro/issues/284)).

- **euporie added to `maestro-python` package** — Python workspaces automatically get Jupyter notebook support via the `maestro-python` Neovim package, which now includes the euporie plugin ([#284](https://github.com/rmkohlman/devopsmaestro/issues/284)).

- **`jupyter` sandbox preset** — New sandbox preset for Jupyter/notebook workspaces with aliases `notebook` and `ipynb`. Uses Python 3.13 slim base image and installs `euporie jupyter` via pip ([#274](https://github.com/rmkohlman/devopsmaestro/issues/274)).

- **euporie added to Python default dev tools** — `euporie` is now included in the default Python language tools generated in Dockerfiles, so Python workspaces automatically include Jupyter notebook support in their dev environment ([#274](https://github.com/rmkohlman/devopsmaestro/issues/274)).

---

## v0.88.0 (2026-04-14)

**New Features**

- **Multi-language support expansion (7 → 22 languages)** — Comprehensive expansion of language detection, Dockerfile generation, sandbox presets, and Neovim integration across 22 languages. Adds 15 new languages while significantly improving support for all existing ones ([#295](https://github.com/rmkohlman/devopsmaestro/issues/295)).

- **Tier 1 languages: Rust, Ruby, Java, .NET/C#, PHP, Kotlin** — Full Dockerfile generation, sandbox presets, private repository detection, and Neovim package mappings ([#299](https://github.com/rmkohlman/devopsmaestro/issues/299), [#300](https://github.com/rmkohlman/devopsmaestro/issues/300), [#301](https://github.com/rmkohlman/devopsmaestro/issues/301), [#304](https://github.com/rmkohlman/devopsmaestro/issues/304), [#305](https://github.com/rmkohlman/devopsmaestro/issues/305), [#306](https://github.com/rmkohlman/devopsmaestro/issues/306)).

- **Tier 2 languages: Elixir, Scala, Swift, Zig** — Enhanced detection, Dockerfile generation, and sandbox presets ([#307](https://github.com/rmkohlman/devopsmaestro/issues/307), [#308](https://github.com/rmkohlman/devopsmaestro/issues/308), [#309](https://github.com/rmkohlman/devopsmaestro/issues/309), [#310](https://github.com/rmkohlman/devopsmaestro/issues/310)).

- **Tier 3 languages: Dart, Lua, R, Haskell, Perl** — Detection, Dockerfile generation, sandbox presets, and Neovim integration ([#311](https://github.com/rmkohlman/devopsmaestro/issues/311), [#312](https://github.com/rmkohlman/devopsmaestro/issues/312), [#313](https://github.com/rmkohlman/devopsmaestro/issues/313), [#314](https://github.com/rmkohlman/devopsmaestro/issues/314), [#315](https://github.com/rmkohlman/devopsmaestro/issues/315)).

**Language Improvements**

- **Go**: default updated to 1.24, pinned digest added ([#297](https://github.com/rmkohlman/devopsmaestro/issues/297)).
- **Node.js**: default updated to 22, `.node-version` detection, package manager detection (yarn/pnpm/bun) ([#298](https://github.com/rmkohlman/devopsmaestro/issues/298)).
- **Python**: default updated to 3.13, `pyproject.toml`/`Pipfile` version detection, Poetry/Pipenv support, `ruff` replaces `ruff-lsp` ([#296](https://github.com/rmkohlman/devopsmaestro/issues/296)).
- **Gleam**: version detection from `gleam.toml`, sandbox preset added ([#302](https://github.com/rmkohlman/devopsmaestro/issues/302)).
- **C/C++**: language detection, Mason tools (clangd), Treesitter parsers, nvim package ([#303](https://github.com/rmkohlman/devopsmaestro/issues/303)).
- **Sandbox generator**: Alpine detection for correct `apk` vs `apt` package manager selection ([#294](https://github.com/rmkohlman/devopsmaestro/issues/294)).

**Bug Fixes**

- **Python base image uses pinned `slim-bookworm` tag** — Fixed floating `slim` tag for reproducible builds ([#293](https://github.com/rmkohlman/devopsmaestro/issues/293)).
- **Sandbox presets use pinned `-bookworm` tags** — Fixed floating tags in sandbox preset base images ([#294](https://github.com/rmkohlman/devopsmaestro/issues/294)).

---

## v0.87.4 (2026-04-13)

**Bug Fixes**

- **Parallel build output corruption (`dvm build -A`)** — Implemented per-workspace output buffering with mutex-protected atomic flush to prevent garbled interleaved output when building multiple workspaces simultaneously ([#290](https://github.com/rmkohlman/devopsmaestro/issues/290)).

---

## v0.87.3 (2026-04-13)

**Bug Fixes**

- **System column missing from get command table outputs** — Added SYSTEM column to `get apps`, `get app`, `get all`, `get workspaces --all`, and `get workspace` detail views so the hierarchy context is complete ([#288](https://github.com/rmkohlman/devopsmaestro/issues/288)).

---

## v0.87.2 (2026-04-13)

**Bug Fixes**

- **`dvm create app --system` now resolves systems globally** — When the `--system` flag is explicitly provided, the command now searches across all domains instead of only the active domain. This fixes workflows where the system exists in a different domain than the current active context. Also fixes `get app`, `get apps`, `delete app`, and YAML apply with the same issue ([#287](https://github.com/rmkohlman/devopsmaestro/issues/287)).

---

## v0.87.1 (2026-04-13)

**Bug Fixes**

- **Database initialization panic when using `-v` flag** — The `-v` (verbose) flag was incorrectly treated as `--version`, skipping database initialization and causing a panic on commands like `dvm admin init`. Fixed by only matching actual command names (`completion`, `version`, `help`) in the skip check. Added defensive nil-interface guard to prevent panics ([#285](https://github.com/rmkohlman/devopsmaestro/issues/285)).

---

## v0.87.0 (2026-04-13)

**New Features**

- **System hierarchy layer** — New organizational level between Domain and App in the resource hierarchy (`ecosystem → domain → system → app → workspace`). Systems group related apps within a domain and support theme, build args, CA certs, and credential cascading. All intermediate hierarchy levels (ecosystem, domain, system) are now optional — only workspace is required (#261).

---

## v0.86.1 (2026-04-13)

**Bug Fixes**

- **BuildKit builder used for sandbox image builds on containerd** — Sandbox image builds now correctly select a BuildKit builder instance when running on a containerd runtime, resolving build failures that occurred on containerd-based environments (#260).

---

## v0.86.0 (2026-04-12)

**New Features**

- **Ephemeral sandbox workspaces (`dvm sandbox`)** — New top-level command (alias: `sb`) with `create`, `get`, `attach`, and `delete` subcommands for spinning up short-lived language environments in containers. Containers are auto-cleaned on exit and tracked at runtime via container labels only — no database records. Image caching enables fast re-creation (#259).

- **5 built-in language presets** — Python, Go, Rust, Node.js, and C++ presets ship out of the box with sensible base images and toolchain defaults (#259).

- **Interactive version picker** — Tab/arrow navigation for selecting a language runtime version without typing exact strings (#259).

- **`--deps` flag for dependency file injection** — Supports `requirements.txt` (Python), `go.mod` (Go), `package.json` (Node.js), `Cargo.toml` (Rust), and `CMakeLists.txt` (C++) so project dependencies are available inside the sandbox immediately (#259).

- **`--version`, `--name`, `--repo`, `--no-cache` flags** — Fine-grained control over sandbox runtime version, container name, image repository, and cache bypass (#259).

- **Extended `ContainerRuntime` interface** — Adds `RemoveContainer`, `RemoveImage`, `ListContainers`, and `ListImages` methods to support sandbox lifecycle management (#259).

---

## v0.85.2 (2026-04-12)

**New Features**

- **Column width constraints for `dvm get git-repos` table** — The NAME column is capped at 40 characters (truncated at end) and the URL column at 50 characters (truncated in the middle) to prevent long values from breaking table layout. Introduces `constrainedTableBuilder` interface and `gitRepoTableBuilder.Constraints()`. Powered by MaestroSDK v0.1.10's new `ColumnConstraint`, `TruncateStrategy` enum, and `TruncateMiddle()` helper (#258).

**Dependency Updates**

- MaestroSDK bumped to v0.1.10 (`ColumnConstraint`, `TruncateStrategy`, `TruncateMiddle()`) (#258).

---

## v0.85.1 (2026-04-12)

**Bug Fixes**

- **Table colors now adapt to terminal color scheme** — Default table styling uses standard ANSI codes that respect the terminal's configured palette instead of hardcoded Catppuccin dark theme colors. Truecolor is only used when a MaestroTheme is active (#230).

---

## v0.85.0 (2026-04-12)

**New Features**

- **Styled table output with box-drawing borders and zebra-striping** — All table-producing commands now render with box-drawing borders, alternating row backgrounds, and a themed header by default. `NO_COLOR` support is respected. Powered by MaestroSDK v0.1.8's styled table renderer (#230).

- **Stress test infrastructure for parallel builds** — New test suite validates staging directory uniqueness and same-domain parallel build behavior (#245).

**Bug Fixes**

- **Treesitter parsers not installing in Docker builds** — Pinned nvim-treesitter to `master` branch, using legacy `configs` API with `sync_install = true`, explicit runtimepath prepend, and `-c "luafile"` post-init execution. Requires MaestroNvim v0.2.7 (#246).

- **gopls requires Go 1.25+ in Docker builds** — Fixed by using `golang:1.25-alpine` as the base image for Go tools installation (#247).

- **`apt-get` exit 100 in parallel builds** — Fixed by assigning per-workspace unique Docker cache mount IDs to prevent concurrent `apt-get` lock contention (#249).

**Dependency Updates**

- MaestroSDK bumped to v0.1.8 (styled table renderer) (#230).
- MaestroNvim bumped to v0.2.7 (treesitter master branch pin) (#246).

---

## v0.84.0 (2026-04-10)

**Bug Fixes**

- **Treesitter TSInstallSync command not recognized in headless Neovim** — Fixed by forcing lazy.nvim to load nvim-treesitter before TSInstall runs, ensuring the plugin is available in headless mode (#232).

- **Treesitter parser installation uses incorrect Lua API method** — Fixed by using the correct `-c "Lazy! load nvim-treesitter" -c "TSInstall ..."` sequence with a verification step to confirm parser installation (#235).

- **Mason tool installation fails during Docker build** — Fixed with per-tool logging, retry logic, Lazy! force-load, and a verification step to confirm each tool is installed before proceeding (#234).

- **Apt package manager lock conflict during parallel Docker builds** — Fixed by assigning per-workspace unique Docker cache mount IDs to prevent concurrent `apt` processes from contending on the same lock (#233).

---

## v0.83.4 (2026-04-10)

**Bug Fixes**

- **ANSI-aware table column alignment** — Table renderers (colored, compact, plain) now strip ANSI escape sequences before measuring column widths, so cells with colored text (e.g. theme swatches) align correctly. Previously, escape-code bytes inflated width calculations causing visible misalignment. Requires MaestroSDK v0.1.7 (#243).

---

## v0.83.3 (2026-04-10)

**New**

- **Shell tab-completion for `dvm set theme`** — Tab-completion now includes all 34+ built-in library themes in addition to user-defined themes, enabling full tab-completion for `dvm set theme <name>` across bash, zsh, and fish shells (#238, #239, #240, #241, #242).

---

## v0.83.2 (2026-04-10)

**New**

- **Color preview swatches (●) in `dvm get themes`** — Theme list now renders ANSI true-color dot swatches showing primary palette colors inline with each theme entry (#237).

---

## v0.83.1 (2026-04-10)

**New**

- **`dvm get themes` and `dvm get theme` as top-level commands** — Both commands are now available directly at the top level with a `t` alias on `get`. Previously only accessible under `dvm nvim get themes` / `dvm nvim get theme` (#231).

**Bug Fixes**

- **Data race in `TestBuildSessionPersistence_FailedBuildTracked`** — Replaced direct struct-field reads/writes with atomic operations (`sync/atomic`) to eliminate the race condition flagged by `-race` (#236).

---

## v0.83.0 (2026-04-10)

**New**

- **Unified theme management with cascading resolution** — `dvm get themes` now merges library and user themes, displays the effective theme with resolution path (global → local → default), and supports YAML/JSON output via `--output`. Theme discovery walks `~/.devopsmaestro/themes/` and the bundled library (#231).

- **WezTerm theme integration** — `dvm build` now resolves the active theme and injects it as environment variables into the WezTerm configuration. A new `themeToWeztermColors()` mapping function translates MaestroTheme color roles to WezTerm's color palette format (#231).

- **THEME column in all table builders** — `dvm get apps`, `dvm get workspaces`, and related table views now include a THEME column showing the effective theme for each resource (#231).

**Changed**

- **`dvm set theme` defaults to `--global`** — The `set theme` command now defaults to global scope with shorthand flag `-g`. Explicit `--local` / `-l` targets workspace-scoped theme overrides (#231).

- **`dvm nvim get` uses `MaximumNArgs(1)`** — The effective theme query now accepts an optional argument for filtering, replacing the previous positional requirement (#231).

---

## v0.82.2 (2026-04-10)

**Bug Fixes**

- **CI test failures: table rendering tests updated to match actual unbordered renderer output** — Two tests in `cmd/pretty_table_test.go` expected bordered output with box-drawing characters for a planned feature that was never shipped. Updated to match the actual MaestroSDK v0.1.6 renderer behavior (unbordered format with horizontal dividers, column headers, and data rows) (#229).

---

## v0.82.1 (2026-04-10)

**Bug Fixes**

- **Build staging directory collision when building multiple workspaces in parallel** — Changed the staging directory key from `filepath.Base(sourcePath) + "-" + workspaceName` to `appName + "-" + workspaceName` so each workspace gets a unique staging directory. Staging cleanup failure is now logged as a warning instead of a fatal error (#227).

- **Dockerfile COPY commands failing for files not copied to staging directory** — `dockerfile_generator.go` now tracks the staging directory, validates that required files are present in staging before emitting `COPY` instructions, and skips `.dockerignore` during the staging copy phase. A new Phase 6b (`validateStagingDirectory`) runs before the Docker build to verify staging is populated (#228).

---

## v0.82.0 (2026-04-10)

**New**

- **Local directory Docker layer cache (`type=local`)** — `dvm build` now persists Docker build layers to disk at `~/.devopsmaestro/build-cache/<app>-<workspace>/` using BuildKit's `type=local` cache. Layers survive `docker system prune` and speed up rebuilds even after Docker's internal cache is cleared. Disabled automatically when `--no-cache` is set (#225).

- **`dvm cache clear` command** — Fully implemented (was previously a stub). Clears persistent build caches by type and reports space freed. Flags: `--all` (clear everything), `--buildkit` (build cache + Docker BuildKit cache), `--staging` (build staging directories), `--dry-run` (preview without deleting) (#225).

**Bug Fixes**

- **Build status incorrectly reports failure for successful builds** — A race condition in the watchdog loop caused `cmdDone` to be processed before the next ticker poll. Because `killedByWatchdog` was never set, the non-zero exit code Docker buildx emits on Colima after a successful export was treated as a failure. Fixed with two layers of defense: the watchdog now performs a final success-condition check before declaring failure when the process exits with error and was not killed by the watchdog; and `buildImage()` adds a fallback that verifies the image exists if the builder returns an error, treating the build as successful when it does (#224).

- **Treesitter parser installation fails with "attempt to index field 'list'"** — The generated Dockerfile called `require('nvim-treesitter').install({...}):wait()`, which broke when nvim-treesitter updated its Lua API. Switched to the stable `TSInstallSync` ex command: `nvim --headless +"TSInstallSync lua vim vimdoc ..." +qa` (#222).

- **Mason install cleanup fails with "Operation not permitted"** — The `mason-install.lua` temp file was `COPY`'d as root but the `rm -f` ran as the non-root `dev` user, causing a permission error. Fixed by adding `--chown=dev:dev` to the `COPY` heredoc so the file is owned by the container user. The unnecessary `rm -f` step was also removed (#222).

---

## v0.81.0 (2026-04-10)

**New**

- **`dvm describe gitrepo <name>`** — Rich status view for a bare git mirror. Displays mirror health, disk usage, branch and tag counts, last sync time, credential status, and the apps and workspaces linked to the mirror (#223).

- **`dvm get branches --repo <name>`** — Lists all branches in a bare git mirror using `git for-each-ref refs/heads/` (#223).

- **`dvm get tags --repo <name>`** — Lists all tags in a bare git mirror using `git for-each-ref refs/tags/` (#223).

- **`MirrorInspector` interface** — New read-only interface in `pkg/mirror/interfaces.go` with `ListBranches`, `ListTags`, `DiskUsage`, and `Verify` methods. Separates read-only inspection from the mutable `MirrorManager` interface (#223).

- **`ListAppsByGitRepoID` / `ListWorkspacesByGitRepoID`** — New `DataStore` methods for querying all apps and workspaces associated with a given git repo by ID. Used by `dvm describe gitrepo` to report relationship visibility (#223).

**Changed**

- **`NewGitMirrorManager` returns interface** — Factory now returns `MirrorManager` (interface) instead of `*GitMirrorManager` (concrete type), consistent with the project's interface-first factory pattern (#223).

- **`getMirrorManager(cmd)` DI helper** — New context helper in `cmd/gitrepo.go` that constructs the `MirrorManager` from a Cobra command context, replacing inline instantiation across gitrepo commands (#223).

---

## v0.80.0 (2026-04-10)

**New**

- **Cache mounts on all builder stages** — `neovim-builder`, `lazygit-builder`, `starship-builder`, and `treesitter-builder` now use `--mount=type=cache` for `apt`/`apk` package caches. The lock contention concern that previously blocked adding these mounts was incorrect for BuildKit (each mount target is isolated per stage). Package caches are preserved across builds even when layers change, preventing re-downloads of apt/apk packages (#221).

- **Cache readiness reporting** — Build output shows `Cache: X/5 registries active (failures listed)` before the Docker build, giving users visibility into which registry caches are healthy. Implemented via `CacheReadiness` struct and `EnsureCachesReady()` / `FormatSummary()` in `pkg/registry/build_support.go` (#221).

**Changed**

- **`docker build` → `docker buildx build`** — All image builds now use `docker buildx build` (via `DockerBuilder`), enabling `--cache-from`/`--cache-to` registry layer cache support in a future phase. Docker's built-in layer cache still provides ~2s warm rebuild times (31 steps, all cached) (#221).

## v0.79.1 (2026-04-10)

**Bug Fixes**

- **`dvm build` parallel path executed a no-op placeholder (0s duration)** — The `buildFn` in the parallel worker pool was a stub that returned immediately without running any Docker build. Replaced with `buildSingleWorkspaceForParallel()`, which executes the full 7-phase build pipeline per workspace. Builds now produce real images with accurate durations (#218).

- **Parallel builds sharing a staging directory caused `Dockerfile.dvm: no such file or directory`** — Multiple workspaces within the same app used the same staging directory (keyed by app name only), causing all but one to fail when run in parallel. Fixed by including the workspace name in the staging directory key (`appName-workspaceName`) so each parallel build gets an isolated staging directory (#218).

- **Build success/failure counter always reported 0 failures** — The final summary line always showed `N succeeded, 0 failed` because the counter never reflected actual results. Replaced with `getBuildSessionCounts()`, which reads succeeded/failed counts from the most recently persisted build session in the database (#218).

## v0.79.0 (2026-04-09)

**Bug Fixes**

- **Workspace images no longer stuck on `:pending`** — After a successful build, the workspace record's image field is updated with the actual built image tag. Previously the tag remained `:pending` indefinitely after the build completed (#217).

**New**

- **Build session persistence** — Every `dvm build` run now creates a session record in the database with a UUID, start/end timestamps, status (`in_progress` / `succeeded` / `failed`), and per-workspace results (status, duration, image tag, error). Sessions older than 30 days are automatically cleaned up. New migration `022_add_build_sessions` adds the `build_sessions` and `build_session_workspaces` tables (#217).

- **`dvm build status` is fully operational** — Previously returned `"no active build session"`. Now shows the latest session with a per-workspace table. New flags: `--session-id <uuid>` to query a specific session, `--history` to list the 10 most recent sessions (#217).

- **`dvm generate template`** — Outputs annotated, copy-paste-ready YAML templates for any resource kind to stdout. 15 kinds supported: `ecosystem`, `domain`, `app`, `workspace`, `infra`, `build-arg`, `ca-cert`, `color`, `credential`, `env`, `mirror`, `nvim-plugin`, `registry`, `source`, `terminal-prompt`. Use `--output json` for JSON, `--all` / `-A` for all kinds as a multi-document YAML stream. Shell completion built in for kind names. Templates are embedded in the binary via `pkg/templates/` (#210).

- **Man page generation** — All three CLIs (`dvm`, `nvp`, `dvt`) now ship a hidden `generate-docs` developer command that generates section-1 man pages and markdown reference docs for every command and subcommand via `cobra/doc`. Run `dvm generate-docs --man-pages --output-dir ./docs/man/` (#35).

- **Parallel batch builds** — Full parallel build orchestration engine. `dvm build --all` discovers and builds every workspace across all apps without requiring an active workspace. Scope flags (`-e`, `-d`, `-a`, `-w`) compose additively with `--all` to narrow the build scope (e.g., `dvm build --all --ecosystem beans-modules`). A semaphore-based worker pool runs builds in parallel with configurable concurrency (`--concurrency`, default: `4`). `--detach` launches the pool in the background and returns a session ID immediately; monitor with `dvm build status`. Failure isolation: one workspace failing does not block others; exit code is non-zero if any workspace fails. `dvm build status` now shows a hint to `dvm build --all` when no active session exists (#213).

**Changed**

- **`dvm build` scope flags are additive** — `--all` and scope flags (`-e/-d/-a/-w`) no longer conflict. They compose: `--all` selects all workspaces, scope flags filter that set. Previously combining `--all` with any scope flag returned an error (#213).

## v0.60.7 (2026-04-06)

**Bug Fixes**

- **Credential export ordering** — `dvm get all -o yaml` now emits credentials after apps and workspaces, fixing restore of app-scoped and workspace-scoped credentials via `dvm apply -f` (#195).

**New**

- **Full multi-ecosystem round-trip test** — `TestFullSystemRoundTrip` in `pkg/resource/handlers/full_roundtrip_test.go` covers 29 resources across 15 kinds and 2 ecosystems. Validates export → wipe → restore → compare fidelity, idempotency, and cross-ecosystem leakage prevention (#173).
- **`KindTerminalPrompt` constant** — Exported from `pkg/resource/handlers/terminal_prompt.go` for handler consistency.

**Dependencies**

- MaestroSDK updated to v0.1.4.

## v0.60.6 (2026-04-06)

**Testing**

- **Progressive stacking integration test** — New test file `pkg/resource/handlers/progressive_stacking_test.go` validates the GitOps round-trip contract by building a full DVM hierarchy incrementally across 5 steps (ecosystem → domain → app → workspace → registries + globals). Verifies YAML export at each step is a strict superset of the previous — no earlier resources disappear as the hierarchy grows. Covers full round-trip fidelity (export → wipe DB → `ApplyList` → re-export → verify parity; 10/10 resources restored) and confirms all handlers are idempotent upserts by applying the same YAML twice and asserting no duplicates (#172).

## v0.60.5 (2026-04-06)

**Bug Fixes**

- **SSH agent forwarding in `dvm attach`** — When a workspace has `ssh_agent_forwarding=true` in the database, the SSH agent socket is now properly forwarded into the container. Previously, `SSHAgentForwarding` was read from the workspace record but never passed into `StartOptions`, so the setting was silently ignored at runtime (#133).
- **Remove `testify/mock` from production binary** — Removed dead `MockExecutor` struct and the `testify/mock` import from `cmd/executor.go`. A test dependency was leaking into the production binary, adding unnecessary bloat (#70).

## v0.60.4 (2026-04-06)

**Bug Fixes**

- **`dvm apply` filesystem provisioning** — `dvm apply -f backup.yaml` now creates actual filesystem directories for GitRepo bare mirrors and Workspace directories after saving DB records. Previously, records were written to the database but no directories were created, causing `dvm build` and `dvm attach` to fail after restoring from a YAML backup. Clone failures are non-fatal (warnings only) and operations are idempotent — re-applying is safe (#193).

## v0.60.3 (2026-04-06)

**Bug Fixes**

- **Container UID/GID** — Runtime no longer hardcodes UID/GID 1000:1000. `StartOptions` and `AttachOptions` now carry dynamic UID/GID fields populated from workspace configuration, with 1000 as the default when unset (#98).

**Security**

- **Exec sessions** — `nerdctl exec` and Docker `ContainerExecCreate` now explicitly set `--user` for defense-in-depth (#97).

## v0.60.2 (2026-04-06)

**Bug Fixes**

- **Version detection** — `detectPythonVersion()` and `detectNodeVersion()` now extract semantic version from `.python-version` and `.nvmrc` files that contain prefixed or suffixed strings (e.g., `daa-api-3.9.9` → `3.9.9`, `v18.17.0` → `18.17.0`). Non-numeric aliases like `lts/*` fall back to defaults (#189).

## v0.60.1 (2026-04-06)

**Bug Fixes**

- **SSH client in workspace containers** — Workspace containers now include `openssh-client` in the default package list. SSH-based git operations (`fetch`, `pull`, `push`) and lazygit now work out of the box. Previously, slim base images lacked the `ssh` binary, causing "No such file or directory" errors (#190).

## v0.60.0 (2026-04-06)

**Enhancements**

- **Shell completion for all scoped commands** — Tab completion now works for `--ecosystem`, `--domain`, `--app`, and `--workspace` flags across all commands that accept them. Added 27 missing flag completions across 7 commands: `set build-arg`, `delete build-arg`, `get build-args`, `set ca-cert`, `delete ca-cert`, `get ca-certs`, and `get all`. Fixed an init ordering bug where completion registration ran before command flags were defined by moving registration to `zz_completion_init.go` (#187).

## v0.59.21 (2026-04-06)

**Bug Fixes**

- **Python build without requirements.txt** — `dvm build` no longer fails for Python projects that have no `requirements.txt`. The Dockerfile generator now checks for `requirements.txt` existence before emitting `COPY`/`pip install` commands; projects without it receive a comment `# No requirements.txt found — skipping pip install` instead (#186).

## v0.59.20 (2026-04-06)

**Bug Fixes**

- **Export resource ordering** — `dvm get all -o yaml` now emits resources in dependency order so that `dvm apply -f` can restore a full backup without cross-reference failures. Registries, Credentials, and GitRepos are emitted before Apps (which reference GitRepos), and GitRepos are emitted after Credentials (which GitRepos reference). Correct emit order: `GlobalDefaults → Ecosystem → Domain → Registry → Credential → GitRepo → App → Workspace → (nvim/terminal/CRD resources)` (#184).

## v0.59.19 (2026-04-06)

**Bug Fixes**

- **Singular get YAML for credential and gitrepo** — `dvm get credential <name> -o yaml` and `dvm get gitrepo <name> -o yaml` now produce proper resource format (`apiVersion/kind/metadata/spec`) instead of plain text or flat map. Output is `dvm apply`-compatible (#183).

## v0.59.18 (2026-04-06)

**Bug Fixes**

- **Terminal plugin handler** — Terminal plugins are now exported by `dvm get all -A -o yaml` and restorable via `dvm apply -f`. Created new `TerminalPluginHandler` with full Apply/Get/List/Delete/ToYAML support, registered in the resource system, and added to the export pipeline and table output (#182).

## v0.59.17 (2026-04-06)

**Bug Fixes**

- **CRD instance export** — `dvm get all -A -o yaml` now exports custom resource instances alongside their CRD definitions. Previously only definitions were exported — all instance data was silently lost on backup/restore. Also fixed `toResourceMap()` to include `apiVersion` field (#180).

## v0.59.16 (2026-04-06)

**Bug Fixes**

- **Nvim theme/package export filter** — `NvimTheme.List()` and `NvimPackage.List()` no longer merge embedded library items into YAML export. Previously, 34 library themes and 12 library packages were exported alongside user-configured items, causing YAML bloat and DB pollution on restore. Only user-configured items are now exported (#181).

## v0.59.15 (2026-04-06)

**Bug Fixes**

- **GlobalDefaults full key coverage** — `kind: GlobalDefaults` now exports and restores all defaults table keys: `nvim-package`, `terminal-package`, `plugins`, 5 registry type defaults (`registry-oci`, `registry-pypi`, `registry-npm`, `registry-go`, `registry-http`), and `registry-idle-timeout`. Previously only `buildArgs`, `caCerts`, and `theme` were handled — 9 keys were silently lost on backup/restore. Also fixed `Delete()` to clear all 12 keys (#177).

## v0.59.14 (2026-04-06)

**Bug Fixes**

- **Registry fields round-trip** — `dvm get registry -o yaml` now exports `enabled`, `storage`, and `idleTimeout` fields. Previously all three were silently dropped, causing registries to restore as disabled with default storage/timeout after backup/restore. Uses `*bool` for Enabled to distinguish "omitted" (defaults to `true`) from "explicitly false" (#178).

## v0.59.13 (2026-04-06)

**Bug Fixes**

- **GitRepo credential round-trip** — git repos created with `--auth-type https --credential <name>` now preserve their credential association across YAML export/restore. Added `credential` field to `GitRepoSpec`; `ToYAML()` resolves `CredentialID` → credential name, `FromYAML()` reads the name from spec, `Apply()` resolves name → `CredentialID`, and `List()`/`Get()` resolve `CredentialID` → name on output (#179).

## v0.59.12 (2026-04-06)

**Bug Fixes**

- **Workspace env round-trip** — `dvm get all -o yaml` now always exports the `env` field for workspaces (previously omitted when empty due to `omitempty` tag on `WorkspaceSpec.Env`). `dvm apply -f` defaults a missing or null env to an empty map before the DB write, fixing `NOT NULL constraint failed: workspaces.env` errors on backup/restore cycles (`models/workspace.go`, `pkg/resource/handlers/workspace.go`) (#185).

## v0.59.11 (2026-03-30)

**Bug Fixes**

- **`kind: GlobalDefaults` missing global default theme on export** — `globalDefaultsSpec` now includes a `Theme` field. `loadGlobalDefaults()`, `ToYAML()`, `Apply()`, and `List()` updated so the theme set via `dvm set theme X` is exported and fully restored after a wipe-and-restore cycle (#174).
- **App–GitRepo association lost on YAML export** — `AppSpec` now includes a `GitRepo` field. `ToYAML()`, `Apply()`, `Get()`, and `List()` resolve the `GitRepoID` foreign key to/from the repo name so the association survives export → wipe → restore without data loss. MaestroSDK updated to v0.1.3 to reorder `DependencyOrder` so `GitRepo` is applied before `App` (#175).
- **`dvm get app <name> -o yaml` missing `metadata.ecosystem`** — singular app get now sets ecosystem from the active context, matching the enrichment applied by the plural list handler (#176).

## v0.59.10 (2026-03-30)

**Bug Fixes**

- **`dvm create credential` flags not mapping to YAML spec** — `--username-var` and `--password-var` now build a `VaultFields` JSON map instead of setting deprecated separate DB columns. CLI-created credentials are now structurally identical to YAML-applied credentials (`cmd/credential.go`) (#157).
- **`ToUsernameConfig()` / `ToPasswordConfig()` missing vault field** — both methods now set `VaultField` on the output `CredentialConfig`, enabling field-level vault secret access instead of a whole-secret fetch (`models/credential.go`) (#157).

## v0.59.9 (2026-03-24)

**opencode Integration (CLI Tool + nvim Plugin)**

Two independent tracks for using the opencode AI coding assistant in DevOpsMaestro workspaces.

- **`tools.opencode: true`** — New opt-in workspace field that installs the opencode TUI binary in the container image at build time. API keys (`ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, etc.) are injected at runtime via `dvm create credential` — never baked into the image.
- **opencode.nvim plugin** — `nickjvandyke/opencode.nvim` (v0.5.2) added to the nvp plugin library with default keybindings: `<C-a>` ask, `<C-x>` select action, `<C-.>` toggle, `go`/`goo` operator. Communicates with opencode over HTTP on port 4096.
- **snacks.nvim plugin** — `folke/snacks.nvim` (v2.30.0) added as the recommended companion for picker integration (`<A-a>` sends picker results to opencode).
- **rmkohlman package** updated to include both `snacks` and `opencode`.
- Either track works standalone — the plugin does not require the CLI binary, and the CLI binary does not require the plugin.

**Bug Fixes**

- **Apply error reporting** — `ApplyList` now reports the resource kind, name, and failure reason for each failed item instead of just a total count. When `dvm apply -f backup.yaml` processes a `kind: List`, actionable errors like `"Item 3 (App 'foo') failed: <reason>"` are surfaced for every failure (#152).
- **Workspace domain disambiguation** — workspace YAML now includes `metadata.ecosystem`, and apply scopes domain resolution to the correct ecosystem. Cross-ecosystem domain name collisions (e.g., two ecosystems each with a domain named "library") are detected and error instead of silently resolving to the wrong one (#153).
- **Plural get YAML format** — `dvm get ecosystems -o yaml`, `dvm get apps -o yaml`, and all other plural get commands now wrap output in a `kind: List` envelope. Output is directly consumable by `dvm apply -f` for backup/restore workflows (#154).
- **Registry trusted host detection** — `isLocalHost()` now recognizes `host.docker.internal`, ensuring `PIP_TRUSTED_HOST` is correctly set when a local devpi registry is used during Docker builds. Prevents pip from rejecting the local registry URL as untrusted (#148).
- **Scoped YAML export ecosystem filtering** — `filterApps()` now correctly filters apps by ecosystem domain membership when using `--ecosystem` scope. Previously, apps from other ecosystems leaked into the export with empty `domain` and `ecosystem` metadata fields (#149).
- **Ecosystem-scoped export credential filtering** — `filterCredentials()` in `cmd/get_all.go` now walks the full hierarchy using filtered domain/app/workspace slices (same pattern as `filterApps`/`filterWorkspaces`). Previously, only top-level credentials were exported; credentials attached to any domain, app, or workspace within the scoped ecosystem are now included (#155).
- **CRDs missing from `dvm get all` export** — `getAll()` now queries CRD resources for both YAML/JSON and table output paths. CRD definitions were silently omitted from backup exports despite having a registered handler; they now appear in `dvm get all -A -o yaml` output and can be re-applied via `dvm apply -f` (#156).

## v0.57.1 (2026-03-18)

**Bug Fixes from Local Testing**

Three bug fixes and one enhancement surfaced during local testing. `nvp theme create` help text listed incorrect preset name examples (`synthwave, matrix, arctic`) — corrected to `coolnight-synthwave, coolnight-matrix, coolnight-arctic` to match the actual MaestroTheme parametric engine keys. `dvt prompt library install` now syncs to the database in addition to the file store, fixing "not found" errors from `dvt prompt get/generate/set` (which read from the database) for library-installed prompts. `dvt prompt delete` now removes from the file store in addition to the database, fixing stale entries in `dvt prompt list`. Added `--short` flag to `dvm version` for consistency with `nvp version --short` and `dvt version --short` — outputs just the version string (e.g., `v0.57.1`).

## v0.57.0 (2026-03-18)

**Package Extraction & Docs Cleanup**

Five packages extracted from dvm into standalone versioned Go modules (`MaestroPalette`, `MaestroSDK`, `MaestroNvim`, `MaestroTheme`, `MaestroTerminal`). Four thin bridge packages (`colorbridge`, `nvimbridge`, `themebridge`, `terminalbridge`) wire the external modules back to dvm's DataStore. 486 files changed, ~46,000 lines removed from the dvm repository. 0 breaking changes — all CLI commands, YAML schemas, and database schemas are identical to v0.56.0. Documentation cleanup: MkDocs nav fix (`projects.md → apps.md`), removed hardcoded version pin from install docs, deleted 4 duplicate root-level doc files, rewrote `architecture.md` to reflect the post-extraction module boundaries.

## v0.56.0 (2026-03-18)

**Hierarchical CA Certificate Cascade**

CA certificates now cascade down the full `global → ecosystem → domain → app → workspace` hierarchy, matching the build args cascade introduced in v0.55.0. The most specific level wins by cert name.

- **`dvm set ca-cert NAME`** — set a CA cert at any level via `--global`, `--ecosystem`, `--domain`, `--app`, or `--workspace`; requires `--vault-secret <name>`; optional `--vault-env` and `--vault-field` flags; `--dry-run` previews without applying; names validated against `^[a-zA-Z0-9][a-zA-Z0-9_-]*$` with a 64-character limit
- **`dvm get ca-certs`** — list CA certs at a specified level; `--effective` (requires `--workspace`) shows the fully merged cascade with a SOURCE provenance column indicating which level each cert originates from; cascade order: `global < ecosystem < domain < app < workspace`; `-o yaml`/`-o json` for scripting
- **`dvm delete ca-cert NAME`** — delete a CA cert at any level using the same hierarchy flags; `-f/--force` skips confirmation; deleting a non-existent cert is a no-op
- **Five-level cascade resolver** — `pkg/cacerts/resolver/` (`HierarchyCACertsResolver`, parallel to `pkg/buildargs/resolver/`); resolved certs fetched from MaestroVault at build time; missing or invalid certs are a fatal build error
- **Strengthened PEM validation** — `crypto/x509` verifies `IsCA` flag; leaf certs, private keys, and non-certificate PEM blocks are rejected
- **New YAML fields** — `spec.caCerts` added to Ecosystem and Domain; `spec.build.caCerts` added to App (Workspace already had this from v0.54.0)
- 0 breaking changes; 1 new migration (`018_add_ca_certs`); 4 new production files; 4 modified production files; 36+ new test cases

## v0.55.0 (2026-03-18)

**Hierarchical Build Args**

Build args now cascade down the full `global → ecosystem → domain → app → workspace` hierarchy. The most specific level wins.

- **`dvm set build-arg KEY VALUE`** — set a build arg at any level via `--global`, `--ecosystem`, `--domain`, `--app`, or `--workspace`; keys validated as legal env var names; `DVM_`-prefixed and dangerous system keys are rejected
- **`dvm get build-args`** — list build args at a specified level using `--global`, `--ecosystem`, `--domain`, `--app`, or `--workspace`; `--effective` shows the fully merged cascade with a provenance column indicating which level each arg originates from; `--output yaml`/`--output json` for scripting
- **`dvm delete build-arg KEY`** — delete a build arg at any level using the same hierarchy flags
- **Five-level cascade resolver** — `global < ecosystem < domain < app < workspace`; resolved args are injected as `--build-arg KEY=VALUE` at build time; values are not persisted in image layers (complements `ARG` declarations from v0.54.0)
- **New YAML fields** — `spec.build.args` added to Ecosystem and Domain resources (already existed on App and Workspace)
- 0 breaking changes; 1 new migration (`017_add_build_args`); 4 new production files; 2 modified production files

## v0.54.0 (2026-03-17)

**✨ Corporate Build Configuration**

Three improvements for building container images in corporate network environments.

- **CA certificate injection** — `spec.build.caCerts` accepts a list of `CACertConfig` objects; certificates are fetched from MaestroVault at build time and injected via `COPY certs/ /usr/local/share/ca-certificates/custom/` + `RUN update-ca-certificates`; `SSL_CERT_FILE`, `REQUESTS_CA_BUNDLE`, and `NODE_EXTRA_CA_CERTS` are set automatically; Alpine images auto-receive the `ca-certificates` package; missing or invalid certs are a fatal build error
- **Build args as `ARG` declarations** — `spec.build.args` keys are emitted as `ARG` declarations (not `ENV`) in both stages of the generated Dockerfile; credentials such as `PIP_INDEX_URL` are available during the build but not persisted in image layers
- **USER directive follows `container.user`** — the `USER` directive in generated Dockerfiles now reads `container.user` instead of being hardcoded to `"dev"`; defaults to `"dev"` when unset
- 0 breaking changes; 1 new type (`CACertConfig`); 2 modified production files

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
