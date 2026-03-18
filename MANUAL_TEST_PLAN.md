# DevOpsMaestro Manual Test Plan

> **Version**: v0.54.0  
> **Last Updated**: March 17, 2026

---

## Quick Start

```bash
# Run Part 1 (cleanup → build)
cd ~/Developer/tools/devopsmaestro
source tests/manual/part1-setup-and-build.sh

# Then do interactive testing (see below)
dvm attach

# After exiting container, run Part 2
source tests/manual/part2-post-attach.sh

# For nvp (NvimOps) testing, see NVIMOPS_TEST_PLAN.md
./tests/manual/nvp/test-nvp.sh
```

---

## Test Structure

| Part | Type | Script/Section |
|------|------|----------------|
| 1 | Automated | `tests/manual/part1-setup-and-build.sh` |
| - | Manual | Interactive container testing (below) |
| 2 | Automated | `tests/manual/part2-post-attach.sh` |
| 19 | Manual | Auto-Token Creation (v0.43.0) |
| 20 | Manual | Build Output Secret Redaction (v0.43.2) |
| 21 | Manual | Container Neovim Environment (v0.44.0) |
| 22 | Manual | Registry Startup Resilience (v0.45.0) |
| 23 | Manual | Auto-Detect Git Default Branch (v0.49.0) |
| 24 | Manual | Scoped Hierarchical Views (v0.52.0) |
| 25 | Manual | List Format Export (v0.53.0) |
| 26 | Manual | Corporate Build Configuration (v0.54.0) |

---

## Part 1: Setup and Build (Automated)

**Script:** `tests/manual/part1-setup-and-build.sh`

**What it tests:**
- Cleanup and initialization
- Platform detection (table, yaml, json formats)
- App creation from GitHub repo
- Workspace management (create, list, switch)
- Image build without nvim config

**Run:**
```bash
cd ~/Developer/tools/devopsmaestro
source tests/manual/part1-setup-and-build.sh
```

**Expected outcome:** Image built, ready for attach.

---

## Interactive Testing: Container Environment

After Part 1 completes, you need to manually test the container environment.

### Step 1: Attach to Workspace

```bash
builtin cd ~/Developer/sandbox/dvm-test-fastapi
dvm attach
```

**Expected:** Interactive zsh shell opens inside container.

---

### Step 2: Basic Environment Verification

**Run these INSIDE the container:**

```bash
# Check shell and user
echo "Shell: $SHELL"
whoami
id

# Check working directory
pwd
ls -la

# Check dev tools
git --version
curl --version | head -1
rg --version
fdfind --version || fd --version
```

**Expected:**
- Shell: `/bin/zsh`
- User: `dev` (uid=1000, gid=1000)
- Working directory: `/workspace`
- All dev tools installed

| Test | Expected | Status |
|------|----------|--------|
| Shell is zsh | /bin/zsh | |
| User is dev | uid=1000 | |
| pwd is /workspace | /workspace | |
| git installed | version 2.x | |
| ripgrep installed | version 13+ | |
| fd-find installed | version 8+ | |

---

### Step 3: Python Tools Verification

**Run INSIDE the container:**

```bash
# Python
python3 --version
pip --version

# LSP and tools
pylsp --version 2>&1 || which pylsp
black --version
isort --version
pytest --version
```

**Expected:** All Python tools installed.

| Test | Expected | Status |
|------|----------|--------|
| Python3 | 3.11+ | |
| pylsp | available | |
| black | 23+ | |
| isort | 5+ | |
| pytest | 7+ | |

---

### Step 4: Neovim Verification

**Run INSIDE the container:**

```bash
# Check neovim
nvim --version | head -3

# Check config (may not exist if build skipped nvim)
ls -la ~/.config/nvim/ 2>/dev/null || echo "No nvim config"

# Test headless startup
nvim --headless +qa
echo "Exit code: $?"
```

**Expected:**
- Neovim 0.9+ installed
- Starts without errors (exit code 0)

| Test | Expected | Status |
|------|----------|--------|
| Neovim version | 0.9+ | |
| Headless startup | exit code 0 | |

---

### Step 5: Interactive Neovim Test (Optional)

**Run INSIDE the container:**

```bash
nvim app/main.py
```

**Inside Neovim, check:**
1. `:LspInfo` - Is pylsp attached?
2. Press `K` on a symbol - Does hover work?
3. `:q` to exit

---

### Step 6: Exit Container

```bash
exit
```

You're now back on the host.

---

## Part 2: Post-Attach Tests (Automated)

**Script:** `tests/manual/part2-post-attach.sh`

**What it tests:**
- Plugin management (create, list, get, delete)
- Error handling (invalid inputs)
- Output formats (table, yaml, json)
- Version and help commands
- Unit tests

**Run:**
```bash
source tests/manual/part2-post-attach.sh
```

---

## Part 3: Namespaced Nvim Commands (v0.6.0+)

### kubectl-style Nvim Commands

These are the current nvim commands using kubectl-style namespacing:

| Command | Description |
|---------|-------------|
| `dvm get nvim plugins` | List all nvim plugins |
| `dvm get nvim plugin <name>` | Get specific plugin |
| `dvm apply nvim plugin -f <file>` | Apply plugin from YAML |
| `dvm delete nvim plugin <name>` | Delete a plugin |
| `dvm edit nvim plugin <name>` | Edit plugin in $EDITOR |

### Test: New Commands Work

```bash
# Test listing plugins (empty if no plugins yet)
dvm get nvim plugins
dvm get nvim plugins -o yaml
dvm get nvim plugins -o json

# Apply a test plugin
cat <<EOF | dvm apply nvim plugin -f -
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: test-plugin
  description: Test plugin for manual testing
spec:
  repo: test/test-plugin
  lazy: true
EOF

# Then get it
dvm get nvim plugin test-plugin
dvm get nvim plugin test-plugin -o yaml

# Delete it
dvm delete nvim plugin test-plugin --force

# Verify it's gone
dvm get nvim plugins
```

**Expected:** All commands work without errors.

### Test: Help Output

```bash
# All namespaced commands should have complete help
dvm get nvim --help
dvm apply nvim --help
dvm delete nvim --help
dvm edit nvim --help
```

**Expected:** Help shows all available subcommands (plugin, theme).

### Test: Old Commands Removed

```bash
# These should NOT work (shows parent command help instead)
dvm get plugins      # Should show 'dvm get' help, not plugin list
dvm get plugin foo   # Should show 'dvm get' help, not plugin details
dvm plugin list      # Should show error: unknown command
```

**Expected:** Old commands are removed, users must use new namespaced commands.

---

## Part 4: Registry System

These scenarios cover the full lifecycle of the registry subsystem introduced in v0.33.x. Run them from a clean state (`rm -rf ~/.devopsmaestro/registries` or equivalent) unless a scenario explicitly depends on prior state.

**Prerequisites:**
- `dvm` binary built from v0.33.3+
- `zot` binary available in PATH (for zot-type registry tests)
- Ports 5000, 5001, 5100 available

---

### Scenario 1: Multi-Registry Lifecycle

Full create → start → stop → delete sequence for a single registry.

```bash
# Create
dvm create registry my-registry --type zot --port 5000

# List (should show status: stopped)
dvm get registries

# Start
dvm start registry my-registry

# List (should show status: running, PID populated)
dvm get registries

# Stop
dvm stop registry my-registry

# List (should show status: stopped)
dvm get registries

# Delete
dvm delete registry my-registry

# List (should be empty)
dvm get registries
```

| Test | Expected | Result |
|------|----------|--------|
| Create succeeds | No error, registry appears in list | |
| Start succeeds | Status changes to "running", PID shown | |
| Stop succeeds | Status changes to "stopped", PID cleared | |
| Delete succeeds | Registry no longer appears in list | |

---

### Scenario 2: Concurrent Registries

Two zot registries running simultaneously on different ports.

```bash
dvm create registry reg-a --type zot --port 5000
dvm create registry reg-b --type zot --port 5001

dvm start registry reg-a
dvm start registry reg-b

# Both should appear as running
dvm get registries

dvm stop registry reg-a
dvm stop registry reg-b

dvm delete registry reg-a
dvm delete registry reg-b
```

| Test | Expected | Result |
|------|----------|--------|
| Both start without conflict | Two entries, both status "running" | |
| Both appear in `get registries` | Table shows two rows | |
| Both stop cleanly | Both status "stopped" | |

---

### Scenario 3: Force Stop

Verify `--force` sends SIGKILL instead of SIGTERM.

```bash
dvm create registry force-test --type zot --port 5000
dvm start registry force-test

# Force kill
dvm stop registry force-test --force

# Should be stopped
dvm get registries
```

| Test | Expected | Result |
|------|----------|--------|
| `--force` flag accepted | No "unknown flag" error | |
| Registry stops | Status "stopped" after force stop | |

---

### Scenario 4: Rapid Start/Stop Cycling

Start and stop the same registry three times in succession.

```bash
dvm create registry cycle-test --type zot --port 5000

for i in 1 2 3; do
  echo "--- Cycle $i ---"
  dvm start registry cycle-test
  dvm get registries
  dvm stop registry cycle-test
  dvm get registries
done

dvm delete registry cycle-test
```

| Test | Expected | Result |
|------|----------|--------|
| Cycle 1: start/stop | Clean state transitions | |
| Cycle 2: start/stop | Clean state transitions | |
| Cycle 3: start/stop | Clean state transitions | |
| No orphan processes | `ps aux | grep zot` shows nothing after final stop | |

---

### Scenario 5: Delete Running Registry

A running registry must be auto-stopped before deletion. (Regression test for Bug 4.)

```bash
dvm create registry live-delete --type zot --port 5000
dvm start registry live-delete

# Confirm it is running
dvm get registries

# Delete without stopping first
dvm delete registry live-delete

# Should be gone, no orphan process
dvm get registries
ps aux | grep zot
```

| Test | Expected | Result |
|------|----------|--------|
| Delete while running | Command succeeds (auto-stops first) | |
| Registry removed from list | No entry in `get registries` | |
| No orphan process | `ps aux` shows no leftover zot process | |

---

### Scenario 6: Duplicate Registry Name

Attempting to create a registry with an already-used name should produce a clean error.

```bash
dvm create registry dup-test --type zot --port 5000
dvm create registry dup-test --type zot --port 5001
```

| Test | Expected | Result |
|------|----------|--------|
| Second create fails | Clear error: name already exists | |
| Error is not a panic | No stack trace, just error message | |
| First registry unaffected | Still present in `get registries` | |

**Cleanup:**
```bash
dvm delete registry dup-test
```

---

### Scenario 7: Port Conflict

Attempting to create a registry on a port already assigned to another registry should fail cleanly.

```bash
dvm create registry port-a --type zot --port 5000
dvm create registry port-b --type zot --port 5000
```

| Test | Expected | Result |
|------|----------|--------|
| Second create fails | Clear error: port already in use | |
| Error is not a panic | No stack trace, just error message | |
| First registry unaffected | Still present in `get registries` | |

**Cleanup:**
```bash
dvm delete registry port-a
```

---

### Scenario 8: Set Default and Get Defaults with Live Status

Verify `registry set-default` and `registry get-defaults` work correctly, including the `-o` output flag. (Regression test for Bug 5.)

```bash
dvm create registry default-test --type zot --port 5000
dvm start registry default-test

# Set as default
dvm registry set-default default-test

# Get defaults (table)
dvm registry get-defaults

# Get defaults with output flags (Bug 5: -o flag was missing)
dvm registry get-defaults -o yaml
dvm registry get-defaults -o json

dvm stop registry default-test
dvm delete registry default-test
```

| Test | Expected | Result |
|------|----------|--------|
| `set-default` succeeds | No error | |
| `get-defaults` (table) shows registry | Name, type, port, status visible | |
| `get-defaults -o yaml` | Valid YAML, no error | |
| `get-defaults -o json` | Valid JSON, no error | |
| Live status reflected | Running registry shows status "running" | |

---

### Scenario 9: Rollout Undo (Not Yet Implemented)

Verify that `rollout undo` returns a clean "not yet implemented" error rather than panicking or silently doing nothing.

```bash
dvm create registry rollout-test --type zot --port 5000
dvm start registry rollout-test

dvm rollout undo registry rollout-test

dvm stop registry rollout-test
dvm delete registry rollout-test
```

| Test | Expected | Result |
|------|----------|--------|
| Command accepted (no "unknown command" error) | Reaches handler | |
| Returns "not yet implemented" message | Clear informational error | |
| No panic | Clean exit, non-zero exit code | |

---

### Scenario 10: Output Formats Across Registry Commands

Verify `-o json` and `-o yaml` produce clean, parseable output for all registry commands. (Regression test for Bugs 6 and 7.)

```bash
dvm create registry fmt-test --type zot --port 5000
dvm start registry fmt-test

# get registries output formats (Bug 7: status field was missing)
dvm get registries -o json
dvm get registries -o yaml
dvm get registries -o wide

# rollout history output formats (Bug 6: sql.Null types leaked)
dvm rollout history registry fmt-test -o json
dvm rollout history registry fmt-test -o yaml

# get-defaults output formats (Bug 5: -o flag was missing)
dvm registry get-defaults -o json
dvm registry get-defaults -o yaml

dvm stop registry fmt-test
dvm delete registry fmt-test
```

**For JSON validation:**
```bash
dvm get registries -o json | python3 -m json.tool
dvm rollout history registry fmt-test -o json | python3 -m json.tool
dvm registry get-defaults -o json | python3 -m json.tool
```

| Test | Expected | Result |
|------|----------|--------|
| `get registries -o json` parses cleanly | Valid JSON, no sql.Null types (e.g., `{"Valid":true,"String":"..."}`) | |
| `get registries -o json` has status field | `.status` section present (Bug 7) | |
| `get registries -o yaml` parses cleanly | Valid YAML | |
| `get registries -o wide` shows extra columns | Wider table with additional fields | |
| `rollout history -o json` parses cleanly | Valid JSON, no raw sql.Null types (Bug 6) | |
| `rollout history -o yaml` parses cleanly | Valid YAML | |
| `get-defaults -o json` accepted | No "unknown flag" error (Bug 5) | |
| `get-defaults -o yaml` accepted | No "unknown flag" error (Bug 5) | |

---

### Scenario 11: Process Crash Recovery

Simulate a crash (kill -9 the registry process) and verify that a subsequent `start` detects the stale PID and recovers cleanly.

```bash
dvm create registry crash-test --type zot --port 5000
dvm start registry crash-test

# Get the PID
dvm get registries

# Kill the process hard (replace <PID> with actual PID from above)
kill -9 <PID>

# Wait a moment
sleep 2

# Attempt to start again - should detect stale PID and recover
dvm start registry crash-test

# Confirm it is running again
dvm get registries

dvm stop registry crash-test
dvm delete registry crash-test
```

| Test | Expected | Result |
|------|----------|--------|
| `kill -9` kills the process | `ps aux` no longer shows the PID | |
| Second `start` succeeds | No error about PID file or "already running" | |
| Registry running after recovery | Status "running" with a new PID | |

---

### Scenario 12: Verbose Mode

Verify `-v` / `--verbose` flag produces debug output for registry commands.

```bash
dvm create registry verbose-test --type zot --port 5000

dvm start registry verbose-test -v
dvm get registries -v
dvm stop registry verbose-test -v

dvm delete registry verbose-test
```

| Test | Expected | Result |
|------|----------|--------|
| `-v` flag accepted on `start` | No "unknown flag" error | |
| `-v` produces debug output | Additional log lines visible (e.g., config path, PID file path) | |
| `-v` flag accepted on `get registries` | No "unknown flag" error | |
| `-v` flag accepted on `stop` | No "unknown flag" error | |

---

### Scenario 13: Non-Zot Registry Types

Verify that athens, devpi, and verdaccio registry types are recognized and created without error. (Full start/stop requires those binaries; creation and listing can be tested standalone.)

```bash
# Create registries of each type
dvm create registry athens-test --type athens --port 5100
dvm create registry devpi-test --type devpi --port 5101
dvm create registry verdaccio-test --type verdaccio --port 5102

# All should appear in list with correct type
dvm get registries
dvm get registries -o yaml
```

| Test | Expected | Result |
|------|----------|--------|
| `--type athens` accepted | Registry created, type "athens" in list | |
| `--type devpi` accepted | Registry created, type "devpi" in list | |
| `--type verdaccio` accepted | Registry created, type "verdaccio" in list | |
| All three appear in `get registries` | Three rows visible | |
| `-o yaml` shows correct types | type field matches what was specified | |

**Cleanup:**
```bash
dvm delete registry athens-test
dvm delete registry devpi-test
dvm delete registry verdaccio-test
```

---

## Part 5: Package Rename & Language Auto-Detection (v0.34.0)

These scenarios cover the package rename from `rmkohlman` to `maestro` and the new language-specific nvim packages with auto-detection during workspace builds.

**Prerequisites:**
- `dvm` and `nvp` binaries built from v0.34.0+
- At least one app configured in DevOpsMaestro

---

### Scenario 14: Package Rename Verification

Verify that all packages previously named `rmkohlman` are now named `maestro` and that the total package count reflects the new language-specific additions.

```bash
# List all nvp packages — should show "maestro" not "rmkohlman"
nvp get packages

# Count total packages (expect 12)
nvp get packages | grep -c '^'

# Set terminal package using new name
dvm use terminal-package maestro

# List emulator entries — should show "maestro" not "rmkohlman"
dvt emulator list
```

| Test | Expected | Result |
|------|----------|--------|
| `nvp get packages` shows "maestro" | No "rmkohlman" entries in list | |
| Total package count | 12 packages listed | |
| `dvm use terminal-package maestro` | Command succeeds, no error | |
| `dvt emulator list` shows "maestro" | No "rmkohlman" entries in emulator list | |

---

### Scenario 15: Language-Specific Package Listing

Verify all 12 packages are present and that language-specific packages have the correct metadata and plugin composition.

```bash
# List all packages — verify all 12 are present
nvp get packages

# Inspect the maestro-go package
nvp get package maestro-go

# Inspect maestro-go in yaml for full detail
nvp get package maestro-go -o yaml
```

**Expected package list (12 total):**

| Package | Type |
|---------|------|
| `core` | Base package |
| `full` | Full package |
| `maestro` | Renamed from rmkohlman |
| `go-dev` | Go development |
| `python-dev` | Python development |
| `maestro-go` | Language-specific (new) |
| `maestro-python` | Language-specific (new) |
| `maestro-rust` | Language-specific (new) |
| `maestro-node` | Language-specific (new) |
| `maestro-java` | Language-specific (new) |
| `maestro-gleam` | Language-specific (new) |
| `maestro-dotnet` | Language-specific (new) |

| Test | Expected | Result |
|------|----------|--------|
| All 12 packages listed | Exactly 12 rows in `nvp get packages` | |
| `maestro-go` exists | Package appears in list | |
| `maestro-go` extends maestro | `extends: maestro` field present in yaml output | |
| `maestro-go` has Go-specific plugins | `nvim-dap-go`, `neotest-go`, `gopher-nvim` present in plugin list | |
| All 7 language packages present | maestro-go, maestro-python, maestro-rust, maestro-node, maestro-java, maestro-gleam, maestro-dotnet all listed | |

---

### Scenario 16: Language Auto-Detection in Build

Verify that `dvm build` selects the correct language-specific nvim package automatically based on the detected app language, and that an explicit `pluginPackage` override takes precedence.

```bash
# --- Test 1: Go app auto-selects maestro-go ---

# Create a Go app workspace (assumes a Go app is registered in DevOpsMaestro)
dvm create workspace go-autodetect-ws --app <your-go-app>

# Build — auto-detection should select maestro-go
dvm build --workspace go-autodetect-ws

# Inspect the built workspace config to confirm package selection
dvm get workspace go-autodetect-ws -o yaml
```

```bash
# --- Test 2: Python app auto-selects maestro-python ---

# Create a Python app workspace
dvm create workspace py-autodetect-ws --app <your-python-app>

# Build — auto-detection should select maestro-python
dvm build --workspace py-autodetect-ws

# Inspect the built workspace config
dvm get workspace py-autodetect-ws -o yaml
```

```bash
# --- Test 3: Explicit pluginPackage overrides auto-detection ---

# Create a Go app workspace with an explicit pluginPackage set
dvm create workspace go-explicit-ws --app <your-go-app> --plugin-package maestro

# Build — should use "maestro", NOT auto-detect "maestro-go"
dvm build --workspace go-explicit-ws

# Confirm the explicit package was used
dvm get workspace go-explicit-ws -o yaml
```

| Test | Expected | Result |
|------|----------|--------|
| Go workspace build auto-detects language | Build output mentions language detection | |
| Go workspace uses `maestro-go` package | `pluginPackage: maestro-go` in workspace yaml | |
| Python workspace build auto-detects language | Build output mentions language detection | |
| Python workspace uses `maestro-python` package | `pluginPackage: maestro-python` in workspace yaml | |
| Explicit `pluginPackage` skips auto-detection | `pluginPackage: maestro` used, not `maestro-go` | |
| No error when fallback package used | Build completes without error for explicit override | |

**Cleanup:**
```bash
dvm delete workspace go-autodetect-ws
dvm delete workspace py-autodetect-ws
dvm delete workspace go-explicit-ws
```

---

## Part 6: Credential Management (v0.35.0+)

These scenarios cover the full lifecycle of credential CLI commands. Credentials are scoped to a single hierarchy level (ecosystem, domain, app, or workspace) and reference a secret source (keychain or environment variable).

**Prerequisites:**
- `dvm` binary built from v0.35.0+
- At least one ecosystem, domain, app, and workspace configured
- macOS Keychain entry `dvm-github` with a value (for keychain tests)
- Environment variable `MY_NPM_TOKEN` set (for env-source tests)

---

### Scenario 17: Create Credential — Keychain Source

Create a credential backed by macOS Keychain, scoped to an ecosystem.

```bash
# Create keychain-sourced credential
dvm create credential GITHUB_TOKEN \
  --source keychain \
  --service dvm-github \
  --ecosystem <eco-name>

# Verify it appears in the list
dvm get credentials --ecosystem <eco-name>

# Get it by name
dvm get credential GITHUB_TOKEN --ecosystem <eco-name>
```

| Test | Expected | Result |
|------|----------|--------|
| Create succeeds | No error, credential appears in list | |
| `get credentials` lists it | Credential visible with correct name and source | |
| `get credential <NAME>` returns detail | Single credential detail shown | |

**Cleanup:**
```bash
dvm delete credential GITHUB_TOKEN --ecosystem <eco-name> --force
```

---

### Scenario 18: Create Credential — Env Source

Create a credential backed by an environment variable, scoped to an app.

```bash
# Create env-sourced credential
dvm create credential NPM_TOKEN \
  --source env \
  --env-var MY_NPM_TOKEN \
  --app <app-name>

# Verify
dvm get credentials --app <app-name>
dvm get credential NPM_TOKEN --app <app-name>
```

| Test | Expected | Result |
|------|----------|--------|
| Create succeeds | No error, credential appears in list | |
| Source is "env" | `source: env` shown in detail output | |
| `env-var` field preserved | `MY_NPM_TOKEN` visible in credential detail | |

**Cleanup:**
```bash
dvm delete credential NPM_TOKEN --app <app-name> --force
```

---

### Scenario 19: Create Credential — Optional Description

Verify that `--description` is accepted and stored.

```bash
dvm create credential CI_TOKEN \
  --source env \
  --env-var CI_SECRET \
  --description "Used for CI pipeline authentication" \
  --domain <domain-name>

dvm get credential CI_TOKEN --domain <domain-name> -o yaml
```

| Test | Expected | Result |
|------|----------|--------|
| `--description` flag accepted | No "unknown flag" error | |
| Description stored | `description: "Used for CI pipeline authentication"` in yaml output | |

**Cleanup:**
```bash
dvm delete credential CI_TOKEN --domain <domain-name> --force
```

---

### Scenario 20: Create Credential — Alias

Verify `cred` works as an alias for `credential` in all three verbs.

```bash
# create alias
dvm create cred ALIAS_TOKEN --source env --env-var ALIAS_VAR --ecosystem <eco-name>

# get alias
dvm get cred ALIAS_TOKEN --ecosystem <eco-name>
dvm get creds --ecosystem <eco-name>

# delete alias
dvm delete cred ALIAS_TOKEN --ecosystem <eco-name> --force
```

| Test | Expected | Result |
|------|----------|--------|
| `create cred` accepted | Command succeeds, no "unknown command" error | |
| `get cred <NAME>` accepted | Returns credential detail | |
| `get creds` accepted | Returns list | |
| `delete cred` accepted | Deletion succeeds | |

---

### Scenario 21: Create Credential — Input Validation Errors

Each of these should fail with a clear, non-panic error message.

```bash
# Missing --source
dvm create credential BAD_TOKEN --ecosystem <eco-name>

# Missing --service when source is keychain
dvm create credential BAD_TOKEN --source keychain --ecosystem <eco-name>

# Missing --env-var when source is env
dvm create credential BAD_TOKEN --source env --ecosystem <eco-name>

# No scope flag (should error: exactly one scope required)
dvm create credential BAD_TOKEN --source env --env-var MY_VAR

# Two scope flags at once
dvm create credential BAD_TOKEN --source env --env-var MY_VAR \
  --ecosystem <eco-name> --app <app-name>
```

| Test | Expected | Result |
|------|----------|--------|
| Missing `--source` | Error: `--source` is required | |
| Keychain missing `--service` | Error: `--service` required for keychain source | |
| Env missing `--env-var` | Error: `--env-var` required for env source | |
| No scope flag | Error: exactly one scope flag required | |
| Two scope flags | Error: only one scope flag allowed | |
| All errors are non-panic | No stack trace in any of the above | |

---

### Scenario 22: Create Credential — Duplicate Name

Creating a credential with an already-used name in the same scope should fail cleanly.

```bash
dvm create credential DUP_TOKEN --source env --env-var DUP_VAR --ecosystem <eco-name>
dvm create credential DUP_TOKEN --source env --env-var DUP_VAR2 --ecosystem <eco-name>
```

| Test | Expected | Result |
|------|----------|--------|
| Second create fails | Clear error: credential already exists | |
| Error is not a panic | No stack trace | |
| First credential unaffected | Still present in `get credentials` | |

**Cleanup:**
```bash
dvm delete credential DUP_TOKEN --ecosystem <eco-name> --force
```

---

### Scenario 23: List Credentials

Verify list commands, scope filtering, and the `-A/--all` flag.

```bash
# Setup: create credentials in different scopes
dvm create credential ECO_CRED  --source env --env-var ECO_VAR  --ecosystem <eco-name>
dvm create credential APP_CRED  --source env --env-var APP_VAR  --app <app-name>
dvm create credential WORK_CRED --source env --env-var WORK_VAR --workspace <ws-name>

# List all credentials across all scopes
dvm get credentials -A

# List filtered by scope
dvm get credentials --ecosystem <eco-name>
dvm get credentials --app <app-name>
dvm get credentials --workspace <ws-name>

# Output formats
dvm get credentials -A -o yaml
dvm get credentials -A -o json
```

| Test | Expected | Result |
|------|----------|--------|
| `-A` lists credentials from all scopes | All three credentials visible | |
| `--ecosystem` filter returns only eco-scoped credentials | Only ECO_CRED listed | |
| `--app` filter returns only app-scoped credentials | Only APP_CRED listed | |
| `--workspace` filter returns only workspace-scoped credentials | Only WORK_CRED listed | |
| `-o yaml` produces valid YAML | Parseable output | |
| `-o json` produces valid JSON | Parseable output (`dvm get credentials -A -o json \| python3 -m json.tool`) | |

**Cleanup:**
```bash
dvm delete credential ECO_CRED  --ecosystem <eco-name> --force
dvm delete credential APP_CRED  --app <app-name> --force
dvm delete credential WORK_CRED --workspace <ws-name> --force
```

---

### Scenario 24: List Credentials — Empty State

Verify the empty-list case produces an informational message rather than an error or blank output.

```bash
# Ensure no credentials exist in scope
dvm get credentials --ecosystem <eco-name>
```

| Test | Expected | Result |
|------|----------|--------|
| Empty list | Info message shown (e.g., "No credentials found") | |
| Exit code 0 | Command exits cleanly | |
| No panic | No stack trace | |

---

### Scenario 25: Get Credential — Nonexistent Name

```bash
dvm get credential DOES_NOT_EXIST --ecosystem <eco-name>
```

| Test | Expected | Result |
|------|----------|--------|
| Command fails | Clear error: credential not found | |
| Error is not a panic | No stack trace | |
| Non-zero exit code | Exit code != 0 | |

---

### Scenario 26: Delete Credential — Force Flag

```bash
dvm create credential DEL_TOKEN --source env --env-var DEL_VAR --ecosystem <eco-name>

# Delete with --force (no confirmation prompt)
dvm delete credential DEL_TOKEN --ecosystem <eco-name> --force

# Confirm it's gone
dvm get credentials --ecosystem <eco-name>
```

| Test | Expected | Result |
|------|----------|--------|
| `--force` skips confirmation prompt | No interactive prompt appears | |
| Credential removed | No longer appears in `get credentials` | |
| `-f` short flag also works | `dvm delete credential ... -f` succeeds | |

---

### Scenario 27: Delete Credential — Interactive Confirmation

```bash
dvm create credential CONFIRM_TOKEN --source env --env-var CONFIRM_VAR --ecosystem <eco-name>

# Delete without --force — type "y" when prompted
dvm delete credential CONFIRM_TOKEN --ecosystem <eco-name>
# > Confirm deletion? [y/N]: y

# Verify gone
dvm get credentials --ecosystem <eco-name>
```

```bash
# Repeat, but type "N" to abort
dvm create credential CONFIRM_TOKEN --source env --env-var CONFIRM_VAR --ecosystem <eco-name>
dvm delete credential CONFIRM_TOKEN --ecosystem <eco-name>
# > Confirm deletion? [y/N]: N
dvm get credentials --ecosystem <eco-name>
```

| Test | Expected | Result |
|------|----------|--------|
| Confirmation prompt appears | `[y/N]` prompt shown | |
| "y" confirms deletion | Credential removed | |
| "N" aborts deletion | Credential still present | |

**Cleanup:**
```bash
dvm delete credential CONFIRM_TOKEN --ecosystem <eco-name> --force
```

---

### Scenario 28: Delete Credential — Nonexistent Name

```bash
dvm delete credential NO_SUCH_CRED --ecosystem <eco-name> --force
```

| Test | Expected | Result |
|------|----------|--------|
| Command fails | Clear error: credential not found | |
| Error is not a panic | No stack trace | |
| Non-zero exit code | Exit code != 0 | |

---

### Scenario 29: Apply Credential from YAML

Verify `dvm apply -f credential.yaml` creates a credential from a YAML manifest.

```bash
# Create credential.yaml
cat <<EOF > /tmp/credential.yaml
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: YAML_TOKEN
  ecosystem: <eco-name>
spec:
  source: env
  envVar: YAML_SECRET_VAR
  description: Created via dvm apply
EOF

# Apply it
dvm apply -f /tmp/credential.yaml

# Verify it exists
dvm get credential YAML_TOKEN --ecosystem <eco-name>
dvm get credential YAML_TOKEN --ecosystem <eco-name> -o yaml
```

| Test | Expected | Result |
|------|----------|--------|
| `dvm apply` succeeds | No error | |
| Credential appears in list | `get credentials` shows YAML_TOKEN | |
| Description persisted | `description: Created via dvm apply` in yaml output | |

---

### Scenario 30: Apply Credential — Update (Re-Apply)

Verify that re-applying a YAML with changed fields updates the credential.

```bash
# Modify the description in credential.yaml
cat <<EOF > /tmp/credential.yaml
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: YAML_TOKEN
  ecosystem: <eco-name>
spec:
  source: env
  envVar: YAML_SECRET_VAR
  description: Updated description via re-apply
EOF

dvm apply -f /tmp/credential.yaml

# Verify update
dvm get credential YAML_TOKEN --ecosystem <eco-name> -o yaml
```

| Test | Expected | Result |
|------|----------|--------|
| Re-apply succeeds | No error | |
| Description updated | `description: Updated description via re-apply` in output | |

**Cleanup:**
```bash
dvm delete credential YAML_TOKEN --ecosystem <eco-name> --force
rm /tmp/credential.yaml
```

---

### Scenario 31: Apply Credential — Invalid YAML (Missing source)

```bash
cat <<EOF > /tmp/bad-credential.yaml
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: BAD_YAML_TOKEN
  ecosystem: <eco-name>
spec:
  envVar: SOME_VAR
  description: Missing source field
EOF

dvm apply -f /tmp/bad-credential.yaml
```

| Test | Expected | Result |
|------|----------|--------|
| Command fails | Clear validation error: `source` is required | |
| Error is not a panic | No stack trace | |
| No credential created | `get credentials` shows no BAD_YAML_TOKEN | |

**Cleanup:**
```bash
rm /tmp/bad-credential.yaml
```

---

---

## Part 7: Registry Version Management (v0.35.1+)

These scenarios cover declarative version management for registry resources. Run them from a clean registry state unless a scenario explicitly depends on prior state.

**Prerequisites:**
- `dvm` binary built from v0.35.1+
- `zot` binary available in PATH (or will be downloaded)
- Port 5003 available for test registry

---

### Scenario 32: Create Registry with `--version`

Create a registry specifying an explicit version via the `--version` flag.

```bash
dvm create registry test --type zot --version 2.1.14 --port 5003

# Verify it appears in the list with the correct version
dvm get registries
```

| Test | Expected | Result |
|------|----------|--------|
| Create succeeds | No error | |
| Registry appears in `get registries` | Row visible in table | |
| Version: 2.1.14 in table | VERSION column shows `2.1.14` | |

**Cleanup (defer to end of Part 7).**

---

### Scenario 33: VERSION Column in `get registries` Table

Verify the VERSION column appears after the TYPE column in the list output.

```bash
dvm get registries
```

| Test | Expected | Result |
|------|----------|--------|
| VERSION column present | Column header `VERSION` visible | |
| VERSION column position | Appears after TYPE column | |
| Version value correct | `2.1.14` shown for the test registry | |

---

### Scenario 34: Version in `get registry <name>` Detail View

Verify the version field is shown in the single-registry detail view.

```bash
dvm get registry test
```

| Test | Expected | Result |
|------|----------|--------|
| `Version:` field shown | `Version: 2.1.14` present in output | |
| Field appears alongside Name, Type, Port, Status | Layout matches existing detail fields | |

---

### Scenario 35: Version in YAML Output

Verify the version is serialized under `spec.version` in YAML output.

```bash
dvm get registry test -o yaml
```

| Test | Expected | Result |
|------|----------|--------|
| `spec.version` field present | `spec.version: 2.1.14` in YAML | |
| Valid YAML | Parses without error | |
| Other spec fields intact | `spec.type`, `spec.port` still present | |

---

### Scenario 36: Start Downloads Correct Version

Verify that `dvm start registry` downloads the version declared in the registry spec before launching.

```bash
dvm start registry test
```

| Test | Expected | Result |
|------|----------|--------|
| Start succeeds | No error | |
| Binary at v2.1.14 downloaded | Version reconciliation log or binary at correct version | |
| Registry reaches "running" status | `dvm get registries` shows status: running | |

```bash
dvm stop registry test
```

---

### Scenario 37: Upgrade via Apply

Apply a YAML manifest that changes the version from `2.1.14` to `2.1.15`, then restart to trigger binary reconciliation.

```bash
cat <<EOF > /tmp/registry-upgrade.yaml
apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: test
spec:
  type: zot
  port: 5003
  version: 2.1.15
EOF

dvm apply -f /tmp/registry-upgrade.yaml

# Verify version updated in the DB
dvm get registry test

# Restart — EnsureBinary() should detect mismatch and download v2.1.15
dvm start registry test
dvm get registries
dvm stop registry test
```

| Test | Expected | Result |
|------|----------|--------|
| Apply succeeds | No error | |
| `get registry test` shows 2.1.15 | `Version: 2.1.15` in detail output | |
| Start reconciles binary | v2.1.15 downloaded on start | |
| Registry reaches "running" status | Status: running after restart | |

**Cleanup:**
```bash
rm /tmp/registry-upgrade.yaml
```

---

### Scenario 38: Invalid Version Rejected

Verify that a non-semver string is rejected at creation time with a clear error.

```bash
dvm create registry x --type zot --version "abc" --port 5004
```

| Test | Expected | Result |
|------|----------|--------|
| Command fails | Non-zero exit code | |
| Error mentions semver | Error text: "invalid version" or "must be semver" | |
| Error is not a panic | No stack trace | |
| No registry created | `dvm get registries` does not list `x` | |

---

### Scenario 39: Rollback on Failed Download

Apply a manifest referencing a nonexistent version, then attempt to start. Verify the original binary is preserved and a clear error is returned.

```bash
cat <<EOF > /tmp/registry-bad-version.yaml
apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: test
spec:
  type: zot
  port: 5003
  version: 99.99.99
EOF

dvm apply -f /tmp/registry-bad-version.yaml

# Start should fail — binary download for 99.99.99 will fail
dvm start registry test
```

| Test | Expected | Result |
|------|----------|--------|
| Start fails | Non-zero exit code, clear download error | |
| Error is not a panic | No stack trace | |
| Original binary preserved | Previous zot binary still present and intact | |
| Registry not left in broken state | `dvm get registries` shows status: stopped (not running) | |

**Cleanup:**
```bash
rm /tmp/registry-bad-version.yaml
dvm delete registry test
```

---

## Part 8: Credential Injection & Environment Variables (v0.36.0)

These scenarios cover the new workspace `env` field, runtime environment injection, GitRepo credential wiring, SQLite foreign key enforcement, build-arg redaction, and the new env var validation package.

**Prerequisites:**
- `dvm` binary built from v0.36.0+
- At least one ecosystem, domain, app, and workspace configured
- At least one credential created (for scenarios 42 and 45)

---

### Scenario 40: Foreign Key Cascade Delete

Verify that deleting a credential cascades to any GitRepo that references it.

```bash
# Create a credential
dvm create credential GH_TOKEN --source env --env-var GITHUB_TOKEN --ecosystem <eco-name>

# Create a gitrepo referencing the credential
dvm create gitrepo test-cascade-repo \
  --url https://github.com/test/repo \
  --auth-type token \
  --credential GH_TOKEN

# Delete the credential
dvm delete credential GH_TOKEN --ecosystem <eco-name> --force

# Verify the gitrepo was cascade-deleted
dvm get gitrepos
```

| Test | Expected | Result |
|------|----------|--------|
| Credential delete succeeds | No error | |
| GitRepo is cascade-deleted | `test-cascade-repo` no longer appears in `dvm get gitrepos` | |
| No orphaned rows | No gitrepo entry references a deleted credential | |

---

### Scenario 41: GitRepo Single-Item Detail View

Verify `dvm get gitrepo <name>` renders cleanly in all output formats.

```bash
# Create a gitrepo (no credential required)
dvm create gitrepo detail-test-repo --url https://github.com/test/repo --auth-type none

# Human-readable (Key: Value format)
dvm get gitrepo detail-test-repo

# YAML output
dvm get gitrepo detail-test-repo -o yaml

# JSON output
dvm get gitrepo detail-test-repo -o json
```

**For JSON validation:**
```bash
dvm get gitrepo detail-test-repo -o json | python3 -m json.tool
```

| Test | Expected | Result |
|------|----------|--------|
| Default output renders cleanly | `Key: Value` format, no `map[string]interface{}` raw output | |
| `-o yaml` produces valid YAML | Parseable YAML, no error | |
| `-o json` produces valid JSON | Parseable JSON, no error | |
| All fields present | Name, URL, auth type visible in output | |

**Cleanup:**
```bash
dvm delete gitrepo detail-test-repo --force
```

---

### Scenario 42: GitRepo Credential Wiring

Verify that `--credential` on `dvm create gitrepo` is stored and retrievable.

```bash
# Create the credential first
dvm create credential MY_CRED --source env --env-var MY_GH_TOKEN --ecosystem <eco-name>

# Create a gitrepo referencing the credential
dvm create gitrepo test-wired-repo \
  --url https://github.com/test/repo \
  --auth-type token \
  --credential MY_CRED

# Verify the credential reference is stored
dvm get gitrepo test-wired-repo
dvm get gitrepo test-wired-repo -o yaml
```

| Test | Expected | Result |
|------|----------|--------|
| Create succeeds | No error | |
| Credential reference shown | `Credential: MY_CRED` visible in detail output | |
| Credential in YAML | `credential: MY_CRED` (or equivalent) present under spec in yaml output | |

**Cleanup:**
```bash
dvm delete gitrepo test-wired-repo --force
dvm delete credential MY_CRED --ecosystem <eco-name> --force
```

---

### Scenario 43: Runtime Environment Injection

Verify that `dvm attach` injects workspace env vars and theme env vars into the running container.

```bash
# Create a workspace YAML with env vars
cat <<EOF > /tmp/env-test-workspace.yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: env-inject-ws
  app: <app-name>
spec:
  env:
    MY_CUSTOM_VAR: hello-from-workspace
    BUILD_ENV: testing
EOF

dvm apply -f /tmp/env-test-workspace.yaml

# Build and attach
dvm build --workspace env-inject-ws
dvm attach --workspace env-inject-ws
```

**Inside the container, verify:**
```bash
# Workspace env vars
echo $MY_CUSTOM_VAR    # Expected: hello-from-workspace
echo $BUILD_ENV        # Expected: testing

# Theme env vars (if a theme is set)
echo $NVIM_THEME       # Expected: theme name

exit
```

| Test | Expected | Result |
|------|----------|--------|
| Workspace env vars injected | `MY_CUSTOM_VAR=hello-from-workspace` inside container | |
| `BUILD_ENV` injected | `BUILD_ENV=testing` inside container | |
| Theme env vars present | `NVIM_THEME` (and related vars) set inside container | |
| No env injection errors on attach | `dvm attach` succeeds without error | |

**Cleanup:**
```bash
dvm delete workspace env-inject-ws --force
rm /tmp/env-test-workspace.yaml
```

---

### Scenario 44: Workspace `env` Field Round-Trip

Verify that the `env` section in a workspace YAML is stored and retrievable without data loss.

```bash
# Create workspace YAML with env section
cat <<EOF > /tmp/roundtrip-workspace.yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: roundtrip-env-ws
  app: <app-name>
spec:
  env:
    DATABASE_URL: postgres://localhost/mydb
    LOG_LEVEL: debug
    FEATURE_FLAGS: flag1,flag2,flag3
EOF

# Apply it
dvm apply -f /tmp/roundtrip-workspace.yaml

# Retrieve and verify env section is preserved
dvm get workspace roundtrip-env-ws -o yaml
```

| Test | Expected | Result |
|------|----------|--------|
| Apply succeeds | No error | |
| `DATABASE_URL` preserved | `DATABASE_URL: postgres://localhost/mydb` in yaml output | |
| `LOG_LEVEL` preserved | `LOG_LEVEL: debug` in yaml output | |
| `FEATURE_FLAGS` preserved | `FEATURE_FLAGS: flag1,flag2,flag3` in yaml output | |
| `env` section present | `spec.env` (or equivalent) block visible in output | |

**Cleanup:**
```bash
dvm delete workspace roundtrip-env-ws --force
rm /tmp/roundtrip-workspace.yaml
```

---

### Scenario 45: Build-Arg Credential Redaction

Verify that credential values are redacted in build command output and never printed in plain text.

```bash
# Create a credential with a known value
dvm create credential BUILD_SECRET --source env --env-var MY_SECRET_VALUE --ecosystem <eco-name>

# Run a build with the credential (observe output carefully)
dvm build --workspace <your-workspace> --credential BUILD_SECRET
```

**Inspect the build output:**

| Test | Expected | Result |
|------|----------|--------|
| Build-arg value not shown in plain text | No raw secret values visible in output | |
| Redacted placeholder shown | `--build-arg BUILD_SECRET=***REDACTED***` (or similar) in output | |
| Build still succeeds | Image built successfully despite redaction in logs | |

**Cleanup:**
```bash
dvm delete credential BUILD_SECRET --ecosystem <eco-name> --force
```

---

### Scenario 46: Env Var Validation

Verify that invalid env var names and dangerous env var names are rejected with clear errors.

```bash
# Invalid name: starts with a number
dvm apply -f - <<EOF
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: invalid-env-ws
  app: <app-name>
spec:
  env:
    1INVALID_VAR: some-value
EOF

# Invalid name: contains special characters
dvm apply -f - <<EOF
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: invalid-env-ws
  app: <app-name>
spec:
  env:
    INVALID-VAR-NAME: some-value
EOF

# Dangerous env var: LD_PRELOAD (denylisted)
dvm apply -f - <<EOF
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: invalid-env-ws
  app: <app-name>
spec:
  env:
    LD_PRELOAD: /tmp/evil.so
EOF

# Dangerous env var: DYLD_INSERT_LIBRARIES (denylisted)
dvm apply -f - <<EOF
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: invalid-env-ws
  app: <app-name>
spec:
  env:
    DYLD_INSERT_LIBRARIES: /tmp/evil.dylib
EOF
```

| Test | Expected | Result |
|------|----------|--------|
| `1INVALID_VAR` rejected | Error: invalid env var name (must start with letter or `_`) | |
| `INVALID-VAR-NAME` rejected | Error: invalid env var name (hyphens not allowed) | |
| `LD_PRELOAD` rejected | Error: env var name is denylisted / not permitted | |
| `DYLD_INSERT_LIBRARIES` rejected | Error: env var name is denylisted / not permitted | |
| All errors are non-panic | No stack traces in any of the above | |
| No workspace created for any invalid case | `dvm get workspaces` does not list `invalid-env-ws` | |

---

## Part 9: Runtime Credential & Env Injection (v0.37.0)

These scenarios cover runtime credential injection, runtime registry env injection, the `--env` flag on workspace creation, DVM_* prefix reservation, and bootstrapping all 5 registries on init.

**Prerequisites:**
- `dvm` binary built from v0.37.0+
- At least one ecosystem, domain, app, and workspace configured
- At least one credential created (for scenarios 48 and 49)
- Environment variable `MY_TEST_SECRET` set (for env source tests)

---

### Scenario 47: Bootstrap All 5 Registries on Init

Verify that `dvm admin init` creates all 5 default registries (oci, pypi, npm, go, http).

```bash
# Fresh init (or re-init)
rm -rf ~/.devopsmaestro
dvm admin init

# All 5 registries should exist
dvm get registries
```

| Test | Expected | Result |
|------|----------|--------|
| Init creates OCI registry | `oci-default` (or similar) with type zot, port 5001 | |
| Init creates PyPI registry | Registry with type devpi, port 5002 | |
| Init creates npm registry | Registry with type verdaccio, port 5003 | |
| Init creates Go registry | Registry with type athens, port 5004 | |
| Init creates HTTP registry | Registry with type squid, port 5005 | |
| All registries are on-demand | Lifecycle is `on-demand` for all | |
| Re-running init is idempotent | Running `dvm admin init` again does not duplicate registries | |

---

### Scenario 48: Runtime Credential Injection in Attach

Verify that credentials from the hierarchy are injected into the container at runtime.

```bash
# Create a credential
dvm create credential MY_SECRET --source env --env-var MY_TEST_SECRET --ecosystem <eco-name>

# Build and attach to a workspace
dvm build --workspace <ws-name>
dvm attach --workspace <ws-name>
```

**Inside the container, verify:**
```bash
echo $MY_SECRET    # Expected: value of MY_TEST_SECRET from host
exit
```

| Test | Expected | Result |
|------|----------|--------|
| Credential injected at runtime | `MY_SECRET` env var available inside container | |
| Value matches host env var | Value equals `MY_TEST_SECRET` from host environment | |
| No error on attach | `dvm attach` succeeds without credential-related errors | |

**Cleanup:**
```bash
dvm delete credential MY_SECRET --ecosystem <eco-name> --force
```

---

### Scenario 49: Dangerous Credential Names Filtered at Runtime

Verify that credentials with dangerous names (LD_PRELOAD, etc.) are silently filtered from runtime injection.

```bash
# Create a credential with a dangerous name (if the create command allows it)
# NOTE: This tests the runtime filter in buildRuntimeEnv, not the create command
dvm attach --workspace <ws-name>
```

**Inside the container, verify:**
```bash
echo $LD_PRELOAD    # Expected: empty (not injected)
exit
```

| Test | Expected | Result |
|------|----------|--------|
| Dangerous env vars not injected | `LD_PRELOAD` is empty inside container even if a credential with that name exists | |
| No error or warning on attach | Filtering is silent, attach succeeds normally | |

---

### Scenario 50: Runtime Registry Env Injection in Attach

Verify that registry-specific environment variables are injected into the container at runtime.

```bash
# Ensure at least one registry is enabled (should be after init)
dvm get registries

# Attach to a workspace
dvm attach --workspace <ws-name>
```

**Inside the container, verify:**
```bash
# Check for registry env vars (depends on which registries are enabled)
env | grep -i proxy
env | grep -i pip_index
env | grep -i goproxy
env | grep -i npm_config
exit
```

| Test | Expected | Result |
|------|----------|--------|
| HTTP_PROXY set (if squid enabled) | `HTTP_PROXY=http://host.docker.internal:5005` or similar | |
| NO_PROXY includes RFC1918 | `10.0.0.0/8,172.16.0.0/12,192.168.0.0/16` in NO_PROXY | |
| PIP_INDEX_URL set (if devpi enabled) | Points to devpi registry URL | |
| GOPROXY set (if athens enabled) | Points to athens registry URL | |
| NPM_CONFIG_REGISTRY set (if verdaccio enabled) | Points to verdaccio registry URL | |

---

### Scenario 51: `--env` Flag on Workspace Creation

Verify that the `--env` / `-e` flag on `dvm create workspace` stores environment variables.

```bash
# Create a workspace with env vars
dvm create workspace env-test-ws --app <app-name> \
  --env MY_VAR=hello \
  --env LOG_LEVEL=debug \
  -e FEATURE_FLAG=enabled

# Verify env vars are stored
dvm get workspace env-test-ws -o yaml
```

| Test | Expected | Result |
|------|----------|--------|
| `--env` flag accepted | No "unknown flag" error | |
| `-e` short flag works | No error on `-e FEATURE_FLAG=enabled` | |
| `MY_VAR` stored | `MY_VAR: hello` in yaml output | |
| `LOG_LEVEL` stored | `LOG_LEVEL: debug` in yaml output | |
| `FEATURE_FLAG` stored | `FEATURE_FLAG: enabled` in yaml output | |
| Multiple `--env` flags supported | All three vars present in yaml | |

**Cleanup:**
```bash
dvm delete workspace env-test-ws --force
```

---

### Scenario 52: `--env` Flag Validation Errors

Verify that invalid env var names and dangerous/reserved names are rejected.

```bash
# Missing value (no = sign)
dvm create workspace bad-env-ws --app <app-name> --env INVALID_FORMAT

# Invalid key (starts with number)
dvm create workspace bad-env-ws --app <app-name> --env 1BAD_KEY=value

# Dangerous key (denylisted)
dvm create workspace bad-env-ws --app <app-name> --env LD_PRELOAD=/tmp/evil.so

# Reserved DVM_* prefix
dvm create workspace bad-env-ws --app <app-name> --env DVM_WORKSPACE=override
```

| Test | Expected | Result |
|------|----------|--------|
| Missing `=` rejected | Error: invalid format, expected KEY=VALUE | |
| Invalid key format rejected | Error: invalid env var name | |
| Dangerous key rejected (hard error) | Error: env var name is not permitted | |
| DVM_* prefix rejected | Error: reserved key prefix | |
| No workspace created for any invalid case | `dvm get workspaces` does not list `bad-env-ws` | |
| All errors are non-panic | No stack traces | |

---

### Scenario 53: DVM_* Metadata Cannot Be Overridden

Verify that workspace env vars cannot override DVM_WORKSPACE, DVM_APP, etc.

```bash
# Create a workspace with env attempting to override DVM_WORKSPACE
# This should be blocked by the --env flag validation
dvm create workspace override-ws --app <app-name> --env DVM_WORKSPACE=hacked

# Even if set via YAML apply, metadata should win at runtime
cat <<EOF > /tmp/override-test.yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: override-ws
  app: <app-name>
spec:
  env:
    DVM_WORKSPACE: hacked
EOF

dvm apply -f /tmp/override-test.yaml
dvm attach --workspace override-ws
```

**Inside container:**
```bash
echo $DVM_WORKSPACE   # Expected: "override-ws" (the REAL workspace name, not "hacked")
exit
```

| Test | Expected | Result |
|------|----------|--------|
| `--env DVM_WORKSPACE=...` rejected | Hard error on create | |
| YAML apply may accept it | No crash (validation may or may not catch it in YAML path) | |
| Runtime metadata always wins | `DVM_WORKSPACE` equals real workspace name inside container | |

**Cleanup:**
```bash
dvm delete workspace override-ws --force
rm /tmp/override-test.yaml
```

---

## Part 10: Keychain Dual-Field Credentials (v0.37.1)

These scenarios cover the new dual-field keychain credential feature: `--username-var` and `--password-var` flags on `dvm create credential`, the corresponding YAML fields (`usernameVar`, `passwordVar`), updated detail/list views, and runtime injection of both vars during `build` and `attach`.

**Prerequisites:**
- `dvm` binary built from v0.37.1+
- At least one ecosystem configured
- macOS Keychain entry with a username+password for `github.com` (for scenarios 62 and 63)
- macOS Keychain entry for `registry.npmjs.org` (for scenario 55)

---

### Scenario 54: Create Dual-Field Credential — Both Vars

Create a keychain-sourced credential with both `--username-var` and `--password-var`.

```bash
dvm create credential github-creds \
  --source keychain --service github.com \
  --username-var GITHUB_USERNAME \
  --password-var GITHUB_PAT \
  --ecosystem <your-ecosystem>

# Verify it appears in the list
dvm get credentials --ecosystem <your-ecosystem>

# Verify detail view
dvm get credential github-creds --ecosystem <your-ecosystem>
```

| Test | Expected | Result |
|------|----------|--------|
| Create succeeds | No error, success message shown | |
| Credential appears in list | `github-creds` visible in `get credentials` | |
| Detail view shows both vars | `Username: GITHUB_USERNAME` and `Password: GITHUB_PAT` fields present | |

**Cleanup (defer to end of Part 10).**

---

### Scenario 55: Create Dual-Field Credential — Password-Var Only

Create a keychain-sourced credential with only `--password-var` (no `--username-var`).

```bash
dvm create credential npm-token \
  --source keychain --service registry.npmjs.org \
  --password-var NPM_TOKEN \
  --ecosystem <your-ecosystem>

# Verify detail view
dvm get credential npm-token --ecosystem <your-ecosystem>
```

| Test | Expected | Result |
|------|----------|--------|
| Create succeeds | No error, success message shown | |
| Detail view shows password var | `Password: NPM_TOKEN` field present | |
| Username field absent or empty | No `Username:` line, or value is empty | |

**Cleanup (defer to end of Part 10).**

---

### Scenario 56: Reject Dual-Field Flags with Env Source

Verify that `--username-var` and `--password-var` are rejected when `--source env` is specified.

```bash
# username-var with env source
dvm create credential bad-cred \
  --source env --env-var TOKEN \
  --username-var MY_USER \
  --ecosystem <your-ecosystem>

# password-var with env source
dvm create credential bad-cred \
  --source env --env-var TOKEN \
  --password-var MY_PASS \
  --ecosystem <your-ecosystem>
```

| Test | Expected | Result |
|------|----------|--------|
| `--username-var` with env source fails | Error: `--username-var and --password-var are only valid with --source=keychain` | |
| `--password-var` with env source fails | Same error message | |
| No credential created | `get credentials` does not list `bad-cred` | |
| Errors are non-panic | No stack trace in either case | |

---

### Scenario 57: Get Dual-Field Credential Detail View

Verify the detail view renders all dual-field fields in order.

```bash
dvm get credential github-creds --ecosystem <your-ecosystem>
```

| Test | Expected | Result |
|------|----------|--------|
| `Name:` field present | `Name: github-creds` | |
| `Scope:` field present | Scope information shown | |
| `Source:` field present | `Source: keychain` | |
| `Service:` field present | `Service: github.com` | |
| `Username:` field present | `Username: GITHUB_USERNAME` | |
| `Password:` field present | `Password: GITHUB_PAT` | |
| No raw struct output | No `map[string]interface{}` or similar | |

---

### Scenario 58: List Credentials — Mixed Legacy and Dual-Field

Verify that the credential list correctly differentiates legacy single-var credentials from dual-field credentials.

```bash
# Ensure a legacy credential exists alongside dual-field credentials
dvm create credential legacy-cred \
  --source keychain --service legacy.example.com \
  --ecosystem <your-ecosystem>

# List all credentials in the ecosystem
dvm get credentials --ecosystem <your-ecosystem>
```

| Test | Expected | Result |
|------|----------|--------|
| Legacy cred shows single-var format | `legacy-cred` displays as `(source: keychain)` | |
| Dual-field cred shows multi-var format | `github-creds` displays as `(source: keychain, vars: GITHUB_USERNAME, GITHUB_PAT)` | |
| Password-only cred shows correctly | `npm-token` displays `(source: keychain, vars: NPM_TOKEN)` or similar | |
| All credentials visible in one list | All three rows present | |

**Cleanup (defer to end of Part 10).**

---

### Scenario 59: Apply Dual-Field Credential via YAML

Verify `dvm apply -f` creates a credential from a YAML manifest with `usernameVar` and `passwordVar`.

```bash
cat <<EOF > /tmp/dual-field-cred.yaml
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: yaml-github-creds
  ecosystem: <your-ecosystem>
spec:
  source: keychain
  service: github.com
  usernameVar: GITHUB_USERNAME
  passwordVar: GITHUB_PAT
EOF

dvm apply -f /tmp/dual-field-cred.yaml

# Verify
dvm get credential yaml-github-creds --ecosystem <your-ecosystem>
dvm get credential yaml-github-creds --ecosystem <your-ecosystem> -o yaml
```

| Test | Expected | Result |
|------|----------|--------|
| `dvm apply` succeeds | No error | |
| Credential appears in list | `yaml-github-creds` visible | |
| `usernameVar` stored | `Username: GITHUB_USERNAME` in detail view | |
| `passwordVar` stored | `Password: GITHUB_PAT` in detail view | |
| `-o yaml` round-trips both fields | `usernameVar: GITHUB_USERNAME` and `passwordVar: GITHUB_PAT` under spec | |

**Cleanup:**
```bash
dvm delete credential yaml-github-creds --ecosystem <your-ecosystem> --force
rm /tmp/dual-field-cred.yaml
```

---

### Scenario 60: Reject Dual-Field Vars in YAML with Env Source

Verify that a YAML manifest with `source: env` and a `usernameVar` set is rejected with a validation error.

```bash
cat <<EOF > /tmp/bad-dual-field-cred.yaml
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: bad-yaml-cred
  ecosystem: <your-ecosystem>
spec:
  source: env
  envVar: MY_TOKEN
  usernameVar: MY_USER
EOF

dvm apply -f /tmp/bad-dual-field-cred.yaml
```

| Test | Expected | Result |
|------|----------|--------|
| Command fails | Non-zero exit code | |
| Error mentions keychain | Error text references `keychain` (e.g., `usernameVar is only valid with source: keychain`) | |
| Error is not a panic | No stack trace | |
| No credential created | `get credentials` does not list `bad-yaml-cred` | |

**Cleanup:**
```bash
rm /tmp/bad-dual-field-cred.yaml
```

---

### Scenario 61: Env Var Name Validation on Dual-Field Flags

Verify that invalid env var names are rejected for both `--username-var` and `--password-var`.

```bash
# Invalid username-var (contains space)
dvm create credential test-cred \
  --source keychain --service svc \
  --username-var "invalid name" \
  --ecosystem <your-ecosystem>

# Invalid username-var (starts with number)
dvm create credential test-cred \
  --source keychain --service svc \
  --username-var 1INVALID \
  --ecosystem <your-ecosystem>

# Invalid password-var (contains hyphen)
dvm create credential test-cred \
  --source keychain --service svc \
  --password-var INVALID-NAME \
  --ecosystem <your-ecosystem>
```

| Test | Expected | Result |
|------|----------|--------|
| Space in `--username-var` rejected | Error: invalid env var name | |
| Leading digit in `--username-var` rejected | Error: invalid env var name | |
| Hyphen in `--password-var` rejected | Error: invalid env var name | |
| No credential created in any case | `get credentials` does not list `test-cred` | |
| All errors are non-panic | No stack traces | |

---

### Scenario 62: Build with Dual-Field Credential

Verify that both env vars are injected as build args when a dual-field credential is associated with a workspace build.

> **Note:** Requires a macOS Keychain entry for `github.com` with a username and password set.

```bash
# Create the dual-field credential (if not already created in Scenario 54)
dvm create credential github-creds \
  --source keychain --service github.com \
  --username-var GITHUB_USERNAME \
  --password-var GITHUB_PAT \
  --ecosystem <your-ecosystem>

# Build a workspace — observe build output for both build args
dvm build <workspace>
```

**Inspect build output for:**
- `--build-arg GITHUB_USERNAME=...` (value redacted)
- `--build-arg GITHUB_PAT=...` (value redacted)

| Test | Expected | Result |
|------|----------|--------|
| Build succeeds | No error, image built | |
| Both build args present | `GITHUB_USERNAME` and `GITHUB_PAT` both passed as `--build-arg` | |
| Values are redacted in log output | No plaintext credential values visible | |

---

### Scenario 63: Attach with Dual-Field Credential

Verify that both env vars are available inside the container when attaching to a workspace with a dual-field credential.

> **Note:** Requires a macOS Keychain entry for `github.com` with a username and password set.

```bash
# Build and attach (credential must be in hierarchy scope)
dvm build <workspace>
dvm attach <workspace>
```

**Inside the container, verify:**
```bash
echo $GITHUB_USERNAME   # Expected: username from Keychain entry
echo $GITHUB_PAT        # Expected: password/token from Keychain entry
exit
```

| Test | Expected | Result |
|------|----------|--------|
| `GITHUB_USERNAME` available in container | Non-empty value from Keychain | |
| `GITHUB_PAT` available in container | Non-empty value from Keychain | |
| No attach errors | `dvm attach` completes without credential-related errors | |

**Cleanup (all Part 10 credentials):**
```bash
dvm delete credential github-creds --ecosystem <your-ecosystem> --force
dvm delete credential npm-token    --ecosystem <your-ecosystem> --force
dvm delete credential legacy-cred  --ecosystem <your-ecosystem> --force
```

---

### Scenario 64: Create Dual-Field Credential — Username-Var Only

Create a keychain-sourced credential with only `--username-var` (no `--password-var`). Verifies that a single username variable is accepted, stored, and injected correctly without requiring a paired password field.

```bash
dvm create credential sso-user \
  --source keychain --service sso.example.com \
  --username-var SSO_USERNAME \
  --ecosystem <your-ecosystem>

# Verify it appears in the list
dvm get credentials --ecosystem <your-ecosystem>

# Verify detail view
dvm get credential sso-user --ecosystem <your-ecosystem>

# Verify YAML round-trip
dvm get credential sso-user --ecosystem <your-ecosystem> -o yaml
```

| Test | Expected | Result |
|------|----------|--------|
| Create succeeds | No error, success message shown | |
| Credential appears in list | `sso-user` visible in `get credentials` | |
| Detail view shows username var | `Username: SSO_USERNAME` field present | |
| Password field absent or empty | No `Password:` line, or value is empty | |
| `-o yaml` shows `usernameVar` | `usernameVar: SSO_USERNAME` under spec in YAML output | |
| `-o yaml` omits or leaves empty `passwordVar` | No `passwordVar` field, or `passwordVar: ""` in YAML output | |

**Cleanup:**
```bash
dvm delete credential sso-user --ecosystem <your-ecosystem> --force
```

---

### Scenario 65: Dual-Field Credential Fan-Out Produces Exactly 2 Env Vars

Verify that a dual-field keychain credential with both `--username-var` and `--password-var` produces **exactly 2 distinct environment variables** in the build and runtime context — one for the username field, one for the password field — using the names specified at creation time.

> **Note:** Requires a macOS Keychain entry for `fanout.example.com` with a username and password set, or substitute any valid Keychain service available on your machine.

```bash
# Create the dual-field credential
dvm create credential fanout-creds \
  --source keychain --service fanout.example.com \
  --username-var FANOUT_USER \
  --password-var FANOUT_PASS \
  --ecosystem <your-ecosystem>

# --- Build check: inspect build args ---
# Run a build and capture the output; count --build-arg occurrences for this credential
dvm build <workspace> 2>&1 | grep -E 'build-arg (FANOUT_USER|FANOUT_PASS)'
```

**Expected build output (values redacted):**
```
--build-arg FANOUT_USER=***REDACTED***
--build-arg FANOUT_PASS=***REDACTED***
```

```bash
# --- Runtime check: attach and inspect env vars ---
dvm attach <workspace>
```

**Inside the container, verify:**
```bash
# Both vars must be set
echo "FANOUT_USER=${FANOUT_USER}"   # Expected: non-empty username from Keychain
echo "FANOUT_PASS=${FANOUT_PASS}"   # Expected: non-empty password from Keychain

# Confirm only 2 vars were injected for this credential (no extra/duplicate vars)
env | grep '^FANOUT_' | sort
exit
```

| Test | Expected | Result |
|------|----------|--------|
| Build output includes `FANOUT_USER` build arg | `--build-arg FANOUT_USER=...` present in build log | |
| Build output includes `FANOUT_PASS` build arg | `--build-arg FANOUT_PASS=...` present in build log | |
| Exactly 2 `FANOUT_*` build args present | `grep 'FANOUT_'` in build output returns exactly 2 lines | |
| `FANOUT_USER` available in container | Non-empty value matching Keychain username | |
| `FANOUT_PASS` available in container | Non-empty value matching Keychain password | |
| Env var names match `--username-var` / `--password-var` | Variable names are `FANOUT_USER` and `FANOUT_PASS` exactly | |
| No extra `FANOUT_*` variables injected | `env \| grep '^FANOUT_'` returns exactly 2 lines | |
| Values are redacted in build log | No plaintext credential values visible in build output | |

**Cleanup:**
```bash
dvm delete credential fanout-creds --ecosystem <your-ecosystem> --force
```

---

## Part 11: Registry Bug Fix Verification (v0.37.2)

These scenarios verify the four registry bugs fixed in v0.37.2: binary download timeout, log file persistence, version defaulting on init and create, and the latent idle-timer deadlock. Run them from a clean registry state unless a scenario explicitly depends on prior state.

**Prerequisites:**
- `dvm` binary built from v0.37.2+
- Zot binary **not** cached (delete `~/.devopsmaestro/bin/zot` if it exists) for Scenario 66
- Port 5010 available for test registries

---

### Scenario 66: Binary Download Timeout (First Use)

Verify that `dvm start registry` completes the Zot binary download (or fails with a timeout error) and does **not** hang indefinitely when the binary is not yet cached.

**Setup:** Remove the cached Zot binary if present.

```bash
rm -f ~/.devopsmaestro/bin/zot*
```

```bash
# Create and start a zot registry — binary must be downloaded on first use
dvm create registry timeout-test --type zot --port 5010
dvm start registry timeout-test
```

| Test | Expected | Result |
|------|----------|--------|
| Command returns (does not hang) | `dvm start` exits within a reasonable time (≤ 5 min on slow connections) | |
| Download completes successfully | Registry reaches status: running | |
| OR download fails with a clear error | Timeout/network error message shown, non-zero exit code (no hang) | |
| No indefinite block | Terminal prompt returns after success or failure | |

**Cleanup:**
```bash
dvm stop registry timeout-test 2>/dev/null || true
dvm delete registry timeout-test
```

---

### Scenario 67: Log File Persistence

Verify that `zot.log` is created in the registry's storage directory and written to while the registry is running, and that the file is cleanly closed after the registry stops (no corruption or truncation).

```bash
# Create and start a registry
dvm create registry log-test --type zot --port 5010
dvm start registry log-test

# Locate and inspect the log file
# (storage directory is typically ~/.devopsmaestro/registries/log-test/)
LOG_FILE=~/.devopsmaestro/registries/log-test/zot.log

ls -lh "$LOG_FILE"
cat "$LOG_FILE"
```

| Test | Expected | Result |
|------|----------|--------|
| `zot.log` exists | File is present in registry storage directory | |
| Log file has content | File is non-empty (Zot writes startup lines on launch) | |
| Log file is being written to | `tail -f "$LOG_FILE"` shows activity while registry runs | |

```bash
# Stop the registry
dvm stop registry log-test

# Verify the log file is intact after stop
ls -lh "$LOG_FILE"
tail -20 "$LOG_FILE"
```

| Test | Expected | Result |
|------|----------|--------|
| Log file still exists after stop | File not deleted on stop | |
| Log file is not corrupted | `tail` outputs readable text, no binary garbage | |
| Log file is not truncated | Content from startup still present | |

**Cleanup:**
```bash
dvm delete registry log-test
```

---

### Scenario 68: Version Defaulting — `dvm init`

Verify that `dvm admin init` populates a non-empty default version for OCI registries in the generated YAML/config, and that the created registries show `version: 2.1.15` (not empty or blank).

```bash
# Fresh init
rm -rf ~/.devopsmaestro
dvm admin init

# Check that the OCI registry has a version set
dvm get registries
dvm get registries -o yaml
```

| Test | Expected | Result |
|------|----------|--------|
| Init completes without error | No error message | |
| OCI registry has a version | VERSION column shows `2.1.15` (not blank) | |
| `-o yaml` shows `spec.version` | `spec.version: 2.1.15` present in YAML output | |
| No other registry types have empty version | All bootstrapped registries show a non-empty version | |

---

### Scenario 69: Version Defaulting — `dvm create registry`

Verify that `dvm create registry` without an explicit `--version` flag populates the default version (`2.1.15` for zot) in the stored resource.

```bash
# Create a registry without specifying a version
dvm create registry default-ver-test --type zot --port 5010

# Confirm the default version was applied
dvm get registry default-ver-test
dvm get registry default-ver-test -o yaml
```

| Test | Expected | Result |
|------|----------|--------|
| Create succeeds | No error | |
| VERSION column shows `2.1.15` | `dvm get registries` table row shows `2.1.15`, not blank | |
| `spec.version` in YAML | `spec.version: 2.1.15` in YAML output (not empty string) | |

**Cleanup:**
```bash
dvm delete registry default-ver-test
```

---

### Scenario 70: Version Defaulting — `dvm registry enable devpi`

Verify that enabling a non-zot registry type (devpi) via `dvm registry enable` also has its default version populated.

```bash
# Enable the devpi registry
dvm registry enable devpi

# Confirm version is set
dvm get registries
dvm get registries -o yaml
```

| Test | Expected | Result |
|------|----------|--------|
| Enable command succeeds | No error | |
| devpi registry appears in list | Row visible in `dvm get registries` table | |
| devpi registry has a version | VERSION column is non-empty for the devpi registry | |
| `-o yaml` confirms `spec.version` set | `spec.version` field present and non-empty for devpi | |

---

### Scenario 71: Idle Timer Deadlock (Informational)

> **Note:** This scenario is primarily covered by automated tests. The deadlock is a latent race condition that manifests under specific timing — it is difficult to trigger manually. This entry documents the fix for tracking purposes and provides a smoke-test to confirm on-demand registries start cleanly.

**Background:** On-demand registries with an idle timeout previously risked a deadlock in the start sequence if the idle timer fired during startup. The automated test covers this directly; the steps below are a smoke-test only.

```bash
# Create an on-demand registry (lifecycle: on-demand with idle timeout)
dvm create registry idle-test --type zot --port 5010

# Start it and confirm it becomes running without hanging
dvm start registry idle-test
dvm get registries
```

| Test | Expected | Result |
|------|----------|--------|
| On-demand registry starts cleanly | Status: running, no hang or deadlock | |
| `dvm start` returns promptly | Terminal prompt returns after start | |
| No error about mutex or channel | No deadlock-related error messages in output | |
| Automated test covers deadlock path | See `pkg/registry/..._test.go` for the definitive test | |

**Cleanup:**
```bash
dvm stop registry idle-test
dvm delete registry idle-test
```

---

## Part 12: BuildKit Builder Stage Robustness (v0.37.4)

These scenarios verify that Dockerfile builder stages fail loudly on download or install errors instead of silently producing a broken image layer. They focus on the hardened curl flags, `set -e` enforcement, download-to-file patterns, and binary verification added in v0.37.4.

**Prerequisites:**
- `dvm` binary built from v0.37.4+
- A working Docker / Colima environment with BuildKit enabled
- Network access (scenarios 72–73 require pulling remote binaries)

---

### Scenario 72: Successful Build with All Builder Stages

Verify a normal `dvm build` completes without errors and that all expected builder-stage binaries (neovim, lazygit, starship, tree-sitter) end up in the final image.

```bash
# Build a workspace image (uses the full builder pipeline)
dvm build --workspace <your-workspace>
```

**Expected:** Build completes with exit code 0. All builder stages appear in the BuildKit output without errors.

**Inside the container (after `dvm attach`), verify each binary is present:**

```bash
nvim --version | head -1          # Expected: NVIM v0.x.x
lazygit --version                 # Expected: version info
starship --version                # Expected: starship x.y.z
tree-sitter --version             # Expected: tree-sitter x.y.z
exit
```

| Test | Expected | Result |
|------|----------|--------|
| Build exits with code 0 | No error, image created | |
| Neovim binary present | `nvim --version` succeeds | |
| Lazygit binary present | `lazygit --version` succeeds | |
| Starship binary present | `starship --version` succeeds | |
| Tree-sitter binary present | `tree-sitter --version` succeeds | |

---

### Scenario 73: Generated Dockerfile Contains Hardened Curl Flags

Verify the generated Dockerfile uses hardened `curl` flags (`--retry 3 --connect-timeout 30 -f`) in every builder stage. Use verbose mode to inspect the generated Dockerfile before building.

```bash
# Generate and display the Dockerfile without building
dvm build --workspace <your-workspace> --dry-run
# OR inspect the generated Dockerfile in the build context:
dvm build --workspace <your-workspace> --verbose 2>&1 | head -200
```

Search the Dockerfile output for each builder stage and confirm:

```bash
# Pipe the verbose output to grep to check flags
dvm build --workspace <your-workspace> --dry-run 2>&1 | grep -E 'curl.*retry'
```

| Test | Expected | Result |
|------|----------|--------|
| Neovim stage curl contains `--retry 3` | Flag present in neovim builder `RUN` block | |
| Lazygit stage curl contains `--retry 3` | Flag present in lazygit builder `RUN` block | |
| Starship stage curl contains `--retry 3` | Flag present in starship builder `RUN` block | |
| Tree-sitter stage curl contains `--retry 3` | Flag present in tree-sitter builder `RUN` block | |
| All curl calls include `-f` | HTTP errors cause failures (no silent error-page downloads) | |

---

### Scenario 74: Broken Download URL Fails Explicitly

Verify that if a builder stage's download URL is unreachable or returns an HTTP error, the build fails with a **clear, actionable error** — not a cryptic checksum or `": not found"` error at a later `COPY --from=<stage>` step.

> **Note:** This requires temporarily breaking a download URL, which cannot be done directly. This scenario is a conceptual regression test — use it as a mental checklist when investigating any future build failures.

**What to observe if a download URL breaks:**
- The broken stage should output a non-zero exit code with a `curl` error message (e.g., `curl: (22) The requested URL returned error: 404`)
- The build should fail **at that stage**, not silently continue and fail at `COPY --from=<stage>`
- The error message should identify the specific stage and URL that failed

**Automated coverage:** The 6 tests added in v0.37.4 (`builders/dockerfile_generator_test.go`) verify the structural patterns that enable this behavior:

```bash
# Run the automated tests that cover the hardened patterns
go test ./builders/... -v -run 'TestDockerfileGenerator'
```

| Test | Expected | Result |
|------|----------|--------|
| Automated tests pass | All 6 new `TestDockerfileGenerator_*` tests pass | |
| `set -e` test passes | `TestDockerfileGenerator_SetEOnAllBuilderStages` ✅ | |
| Hardened curl flags test passes | `TestDockerfileGenerator_HardenedCurlFlagsInAllBuilderStages` ✅ | |
| No-pipe test passes | `TestDockerfileGenerator_NoPipeToShellInBuilderStages` ✅ | |
| Version guard test passes | `TestDockerfileGenerator_LazygitVersionGuard` ✅ | |
| Binary verification test passes | `TestDockerfileGenerator_BinaryVerificationInAllBuilderStages` ✅ | |
| Golangci-lint test passes | `TestDockerfileGenerator_GolangciLintHardenedInstall` ✅ | |

---

## Part 13: BuildKit Structural Improvements (v0.37.5)

These scenarios verify the structural improvements made to the Dockerfile generator in v0.37.5: dynamic tree-sitter versioning, fail-fast architecture detection, and custom user support for nvim config.

**Prerequisites:**
- `dvm` binary built from v0.37.5+
- A working Docker / Colima environment with BuildKit enabled
- Network access (Scenario 46 requires pulling from GitHub API)

---

### Scenario 46: Tree-sitter Dynamic Versioning

Verify the tree-sitter builder stage queries the GitHub API for the latest version at build time instead of using a hardcoded version string.

**Prerequisites:** Running Colima/Docker, a workspace with dvm build capability.

```bash
# Generate the Dockerfile and inspect the tree-sitter builder stage
dvm build --workspace <your-workspace> --dry-run 2>&1 > Dockerfile.dvm
```

**Verify:**

```bash
# Should return 0 — no hardcoded version present
grep -c "v0.24.6" Dockerfile.dvm

# Should return >0 — dynamic version variable present
grep -c "TREESITTER_VERSION" Dockerfile.dvm
```

| Test | Expected | Result |
|------|----------|--------|
| No hardcoded `v0.24.6` | `grep -c "v0.24.6" Dockerfile.dvm` returns `0` | |
| Dynamic version variable present | `grep -c "TREESITTER_VERSION" Dockerfile.dvm` returns `>0` | |
| GitHub API query present | `grep "api.github.com" Dockerfile.dvm` matches tree-sitter stage | |
| Version validation guard present | `grep "\[ -n.*TREESITTER_VERSION" Dockerfile.dvm` returns a match | |

---

### Scenario 47: Builder Stage Architecture Fail-Fast

Verify that every builder stage fails explicitly with an error message when an unknown architecture is detected, instead of silently falling back to x86_64.

**Prerequisites:** None (inspect generated Dockerfile).

```bash
# Generate and inspect the Dockerfile
dvm build --workspace <your-workspace> --dry-run 2>&1 > Dockerfile.dvm
```

**Verify:**

```bash
# Should return matches for each builder stage — not 0
grep "Unsupported architecture" Dockerfile.dvm
```

| Test | Expected | Result |
|------|----------|--------|
| Unsupported architecture message present | `grep "Unsupported architecture" Dockerfile.dvm` returns matches | |
| `exit 1` paired with error message | Each `"Unsupported architecture"` line is followed by or paired with `exit 1` | |
| No silent x86_64 fallback | No `else` branch in builder arch detection silently sets `ARCH=amd64` without error | |

---

### Scenario 48: Custom User Nvim Config

Verify that when a workspace is configured with a custom `container.user`, the generated Dockerfile copies nvim config to the correct user home directory instead of hardcoding `/home/dev/`.

**Prerequisites:** A workspace YAML with `container.user: myuser`.

```bash
# Build (or dry-run) with a workspace that has container.user: myuser
dvm build --workspace <your-custom-user-workspace> --dry-run 2>&1 > Dockerfile.dvm
```

**Verify:**

```bash
# Should return matches — custom user path is used for nvim config
grep "/home/myuser/" Dockerfile.dvm

# Should only appear in user creation (useradd/adduser), NOT in nvim COPY/chown directives
grep "/home/dev/" Dockerfile.dvm
```

| Test | Expected | Result |
|------|----------|--------|
| Nvim COPY uses custom user path | `grep "/home/myuser/" Dockerfile.dvm` returns matches | |
| `/home/dev/` not used in nvim section | Any `/home/dev/` references are limited to user creation lines, not nvim COPY/chown | |
| chown uses custom user | `grep "chown.*myuser" Dockerfile.dvm` returns a match | |
| USER directive references custom user | `grep "^USER myuser" Dockerfile.dvm` returns a match | |

---

## Part 14: Dockerfile Generator Purity (v0.38.0)

These scenarios verify the correctness fixes and structural improvements made to the Dockerfile generator in v0.38.0: computed Alpine detection, PathConfig injection for nvim staging, and error propagation from the nvim section.

**Prerequisites:**
- `dvm` binary built from v0.38.0+
- A working Docker / Colima environment with BuildKit enabled

---

### Scenario 49: Computed Alpine Detection Matches Generated Image

Verify that the generated Dockerfile's package manager (`apk` vs `apt-get`) and user-creation command (`adduser` vs `useradd`) are consistent with the `FROM` image selected for the given language, and that switching a golang workspace to a Debian base image correctly uses Debian tooling throughout.

**Setup:** Two workspaces — one golang workspace using the default Alpine image, one golang workspace explicitly configured with a Debian base image.

```bash
# Inspect the generated Dockerfile for an Alpine golang workspace
dvm build --workspace <golang-alpine-workspace> --dry-run 2>&1 > /tmp/Dockerfile.alpine.dvm
```

**Verify Alpine workspace:**

```bash
# Base image should be Alpine (e.g., golang:1.x-alpine)
grep "^FROM golang" /tmp/Dockerfile.alpine.dvm | head -1

# Package manager should be apk
grep "apk add" /tmp/Dockerfile.alpine.dvm

# User creation should use adduser (Alpine style)
grep "adduser" /tmp/Dockerfile.alpine.dvm
```

```bash
# Inspect the generated Dockerfile for a Debian golang workspace
dvm build --workspace <golang-debian-workspace> --dry-run 2>&1 > /tmp/Dockerfile.debian.dvm
```

**Verify Debian workspace:**

```bash
# Base image should be Debian (e.g., golang:1.x or golang:1.x-bookworm)
grep "^FROM golang" /tmp/Dockerfile.debian.dvm | head -1

# Package manager should be apt-get (not apk)
grep "apt-get install" /tmp/Dockerfile.debian.dvm
grep -c "apk add" /tmp/Dockerfile.debian.dvm    # Expected: 0

# User creation should use useradd (Debian style)
grep "useradd" /tmp/Dockerfile.debian.dvm
grep -c "adduser" /tmp/Dockerfile.debian.dvm    # Expected: 0 (or only in Alpine stages)
```

| Test | Expected | Result |
|------|----------|--------|
| Alpine workspace uses `apk add` | Package install uses `apk`, not `apt-get` | |
| Alpine workspace uses `adduser` | User creation uses `adduser` | |
| Debian workspace uses `apt-get install` | Package install uses `apt-get`, not `apk` | |
| Debian workspace uses `useradd` | User creation uses `useradd`, not `adduser` | |
| Automated tests pass | `TestIsAlpine_ComputedPerLanguage` and `TestIsAlpine_FieldMatchesGeneratedImage` ✅ | |

```bash
# Automated coverage
go test ./builders/... -v -run 'TestIsAlpine'
```

---

### Scenario 50: PathConfig Injection Works for Nvim Staging

Verify that the nvim config staging section in the generated Dockerfile references the correct path from the host, and that the path used is derived from the PathConfig rather than a hardcoded `os.UserHomeDir()` call.

**Prerequisites:** A workspace configured to include neovim (i.e., a plugin manifest with nvim enabled).

```bash
# Build (or dry-run) with a workspace that has nvim enabled
dvm build --workspace <nvim-enabled-workspace> --dry-run 2>&1 > /tmp/Dockerfile.nvim.dvm
```

**Verify the nvim config COPY path is present and plausible:**

```bash
# The COPY directive for nvim config should reference a path
grep -i "nvim" /tmp/Dockerfile.nvim.dvm

# The path in the COPY instruction should match ~/.config/nvim or equivalent
grep "COPY.*nvim" /tmp/Dockerfile.nvim.dvm
```

**Verify the automated path injection test:**

```bash
go test ./builders/... -v -run 'TestGenerateNvimSection_UsesPathConfig'
```

| Test | Expected | Result |
|------|----------|--------|
| Nvim COPY present when nvim config exists | `COPY` directive for nvim config appears in Dockerfile | |
| Path is consistent with host home dir | No literal `/Users/dev` or other hardcoded developer path | |
| `TestGenerateNvimSection_UsesPathConfig` passes | Injected PathConfig controls path used in section ✅ | |

---

### Scenario 51: Error Propagation from Nvim Section

Verify that when the nvim config directory is absent, `dvm build` either succeeds with a graceful skip comment in the Dockerfile **or** returns a clear error — not a silent empty section or a panic.

**Setup:** A workspace where no `~/.config/nvim` directory exists on the host.

```bash
# Temporarily rename or remove the nvim config directory
mv ~/.config/nvim ~/.config/nvim.bak 2>/dev/null || echo "No nvim config to move"

# Run a build (or dry-run) that would include the nvim section
dvm build --workspace <nvim-enabled-workspace> --dry-run 2>&1 > /tmp/Dockerfile.nonvim.dvm
echo "Exit code: $?"
```

**Verify behavior:**

```bash
# Option A: Build succeeds with a skip comment
grep -i "nvim" /tmp/Dockerfile.nonvim.dvm
# Expected: a comment like "# nvim config not found, skipping"

# Option B: Build returns a clear error (non-zero exit, no panic)
# Expected: error message references nvim config path, no stack trace
```

```bash
# Restore nvim config
mv ~/.config/nvim.bak ~/.config/nvim 2>/dev/null || true
```

| Test | Expected | Result |
|------|----------|--------|
| No panic when nvim config absent | No stack trace or nil pointer dereference | |
| Graceful skip: comment in Dockerfile | `# nvim config not found` (or similar) in Dockerfile output | |
| OR clear error returned | Non-zero exit with descriptive error message referencing nvim path | |
| Silent empty section NOT acceptable | Dockerfile does not contain an empty/incomplete nvim section | |
| Automated test passes | `TestGenerate_NvimConfig_GracefulSkip` ✅ | |

```bash
# Automated coverage
go test ./builders/... -v -run 'TestGenerate_NvimConfig_GracefulSkip'
```

---

## Part 15: Python HTTPS Token Substitution (v0.38.1)

These scenarios verify the two bug fixes in v0.38.1: the SSH detection regex false positive and the Dockerfile generator dispatch for Python private dependencies. Each scenario focuses on the generated Dockerfile content, which can be inspected without a live Docker build using `--dry-run`.

**Prerequisites:**
- `dvm` binary built from v0.38.1+
- A Python workspace with a `requirements.txt` accessible to `dvm build`
- No live Docker build environment required for Scenarios 52–54 (dry-run only)

---

### Scenario 52: HTTPS-Only Private Dependencies — sed Substitution Path

Verify that a `requirements.txt` containing only HTTPS private dependencies (with `${VAR}` token placeholders) generates a Dockerfile with sed substitution — and does **not** include an SSH mount.

**Setup:** Create a `requirements.txt` with HTTPS private deps:

```
flask==2.3.0
beansmodels @ git+https://${GITHUB_USERNAME}:${GITHUB_PAT}@github.com/RogersCommunications/beans-models.git@2025.03.27.1430
```

```bash
# Dry-run the build and save the generated Dockerfile
dvm build --workspace <your-python-workspace> --dry-run 2>&1 > /tmp/Dockerfile.https.dvm
```

**Verify sed substitution is present:**

```bash
# ARG declarations for both token variables
grep "ARG GITHUB_USERNAME" /tmp/Dockerfile.https.dvm
grep "ARG GITHUB_PAT" /tmp/Dockerfile.https.dvm

# requirements-template.txt is copied (sed operates on the template)
grep "requirements-template.txt" /tmp/Dockerfile.https.dvm

# sed substitution rewrites placeholders before pip install
grep 'sed.*GITHUB_USERNAME' /tmp/Dockerfile.https.dvm
grep 'sed.*GITHUB_PAT' /tmp/Dockerfile.https.dvm
```

**Verify SSH mount is absent:**

```bash
# Must return 0 (no match) — HTTPS-only deps should not produce an SSH mount
grep -c -- '--mount=type=ssh' /tmp/Dockerfile.https.dvm
```

| Test | Expected | Result |
|------|----------|--------|
| `ARG GITHUB_USERNAME` declared | Present in Dockerfile | |
| `ARG GITHUB_PAT` declared | Present in Dockerfile | |
| `requirements-template.txt` referenced | Template copy step present | |
| `sed` substitution for `GITHUB_USERNAME` | `sed "s/\${GITHUB_USERNAME}/$GITHUB_USERNAME/g"` (or equivalent) present | |
| `sed` substitution for `GITHUB_PAT` | `sed "s/\${GITHUB_PAT}/$GITHUB_PAT/g"` (or equivalent) present | |
| No SSH mount in pip install | `grep -c '--mount=type=ssh'` returns `0` | |

---

### Scenario 53: SSH-Only Private Dependencies — SSH Mount Path

Verify that a `requirements.txt` containing only SSH private dependencies generates a Dockerfile with an SSH mount — and does **not** include sed substitution.

**Setup:** Create a `requirements.txt` with SSH-only private deps:

```
flask==2.3.0
mylib @ git+ssh://git@github.com/myorg/mylib.git@v1.2.3
```

```bash
# Dry-run the build and save the generated Dockerfile
dvm build --workspace <your-python-workspace> --dry-run 2>&1 > /tmp/Dockerfile.ssh.dvm
```

**Verify SSH mount is present:**

```bash
# pip install should use --mount=type=ssh
grep -- '--mount=type=ssh' /tmp/Dockerfile.ssh.dvm
```

**Verify sed substitution is absent:**

```bash
# Must return 0 — SSH-only deps should not produce sed substitution
grep -c 'requirements-template.txt' /tmp/Dockerfile.ssh.dvm
grep -c 'sed.*GITHUB' /tmp/Dockerfile.ssh.dvm
```

| Test | Expected | Result |
|------|----------|--------|
| `--mount=type=ssh` present in pip install | SSH mount directive present | |
| No `requirements-template.txt` | `grep -c 'requirements-template.txt'` returns `0` | |
| No sed substitution | `grep -c 'sed.*GITHUB'` returns `0` | |

---

### Scenario 54: Mixed Dependencies — Both sed and SSH Mount

Verify that a `requirements.txt` containing both HTTPS token deps and SSH deps generates a Dockerfile with **both** sed substitution and an SSH mount in the pip install step.

**Setup:** Create a `requirements.txt` with mixed private deps:

```
flask==2.3.0
beansmodels @ git+https://${GITHUB_USERNAME}:${GITHUB_PAT}@github.com/RogersCommunications/beans-models.git@2025.03.27.1430
mylib @ git+ssh://git@github.com/myorg/mylib.git@v1.2.3
```

```bash
# Dry-run the build and save the generated Dockerfile
dvm build --workspace <your-python-workspace> --dry-run 2>&1 > /tmp/Dockerfile.mixed.dvm
```

**Verify both mechanisms are present:**

```bash
# sed substitution for HTTPS credentials
grep 'sed.*GITHUB_USERNAME' /tmp/Dockerfile.mixed.dvm
grep 'sed.*GITHUB_PAT'      /tmp/Dockerfile.mixed.dvm

# SSH mount for SSH dependencies
grep -- '--mount=type=ssh'  /tmp/Dockerfile.mixed.dvm

# ARG declarations for token variables
grep 'ARG GITHUB_USERNAME'  /tmp/Dockerfile.mixed.dvm
grep 'ARG GITHUB_PAT'       /tmp/Dockerfile.mixed.dvm
```

| Test | Expected | Result |
|------|----------|--------|
| `ARG GITHUB_USERNAME` declared | Present | |
| `ARG GITHUB_PAT` declared | Present | |
| sed substitution for `GITHUB_USERNAME` | Present | |
| sed substitution for `GITHUB_PAT` | Present | |
| `--mount=type=ssh` present | Present | |

---

### Scenario 55: Version-Pinned HTTPS URL Not Misclassified as SSH

Verify that an HTTPS URL containing `.git@<tag>` (e.g., `repo.git@2025.03.27.1430`) does **not** trigger SSH detection — this was the root cause of the v0.38.1 regression.

**Setup:** Create a `requirements.txt` with a version-pinned HTTPS URL (no `${VAR}` tokens):

```
flask==2.3.0
beansmodels @ git+https://github.com/RogersCommunications/beans-models.git@2025.03.27.1430
```

```bash
# Dry-run the build and save the generated Dockerfile
dvm build --workspace <your-python-workspace> --dry-run 2>&1 > /tmp/Dockerfile.noprivate.dvm
echo "Exit code: $?"
```

**Verify neither SSH nor sed substitution is injected:**

```bash
# Plain pip install — no credential handling expected
grep -c -- '--mount=type=ssh' /tmp/Dockerfile.noprivate.dvm
grep -c 'requirements-template.txt' /tmp/Dockerfile.noprivate.dvm
grep -c 'ARG GITHUB' /tmp/Dockerfile.noprivate.dvm
```

| Test | Expected | Result |
|------|----------|--------|
| Build completes without error | Exit code 0 | |
| No SSH mount injected | `grep -c '--mount=type=ssh'` returns `0` | |
| No sed substitution injected | `grep -c 'requirements-template.txt'` returns `0` | |
| No ARG declarations for GitHub vars | `grep -c 'ARG GITHUB'` returns `0` | |
| Plain `pip install -r requirements.txt` | Standard pip install command present | |

---

### Scenario 56: Automated Test Coverage (Regression Gate)

Run the automated tests added in v0.38.1 to confirm the regression is covered and all 12 subtests pass.

```bash
# Run both new test functions
go test ./utils/... -v -run 'TestDetectPythonPrivateRepos'
go test ./builders/... -v -run 'TestDockerfileGenerator_PythonPrivateRepos'
```

| Test | Expected | Result |
|------|----------|--------|
| `TestDetectPythonPrivateRepos` passes | All 8 subtests ✅ | |
| `TestDockerfileGenerator_PythonPrivateRepos` passes | All 4 subtests ✅ | |
| No regressions in existing tests | `go test ./...` all green ✅ | |

**Verify individual subtests:**

```bash
go test ./utils/... -v -run 'TestDetectPythonPrivateRepos/HTTPS-only'
go test ./utils/... -v -run 'TestDetectPythonPrivateRepos/SSH-only'
go test ./utils/... -v -run 'TestDetectPythonPrivateRepos/Mixed'
go test ./builders/... -v -run 'TestDockerfileGenerator_PythonPrivateRepos/HTTPS-only_sed_path'
go test ./builders/... -v -run 'TestDockerfileGenerator_PythonPrivateRepos/SSH-only_mount_path'
go test ./builders/... -v -run 'TestDockerfileGenerator_PythonPrivateRepos/Mixed'
go test ./builders/... -v -run 'TestDockerfileGenerator_PythonPrivateRepos/No_private_repos_plain_install'
```

---

## Part 16: Credential Resolution Robustness (v0.38.2)

These scenarios verify the three credential resolution bug fixes in v0.38.2: user-visible warnings when credential resolution fails, the env var fallback rescue path for failed keychain lookups, and correct dual-field keychain resolution with the `-a $USER` filter.

**Prerequisites:**
- `dvm` binary built from v0.38.2+
- At least one ecosystem configured
- macOS Keychain entry available for dual-field tests (Scenario 59)
- Environment variable `GITHUB_USERNAME` set (for Scenario 58)

---

### Scenario 57: Build with Intentionally Missing Keychain Credential

Verify that when a credential references a Keychain service that does not exist, `dvm build` displays a visible warning to the user instead of silently proceeding with an empty credential value.

**Setup:** Create a credential referencing a Keychain service that does not exist on your machine.

```bash
# Create a credential referencing a nonexistent keychain service
dvm create credential MISSING_CRED \
  --source keychain \
  --service nonexistent.service.that.does.not.exist \
  --ecosystem <your-ecosystem>

# Associate it with a workspace (or ensure it is in the credential hierarchy)
# Then run a build
dvm build --workspace <your-workspace>
```

| Test | Expected | Result |
|------|----------|--------|
| Build does not hang or panic | `dvm build` exits cleanly (success or failure) | |
| Warning is displayed to user | A `render.Warning()`-style message appears in terminal output | |
| Warning identifies the credential | Warning text includes the credential name (`MISSING_CRED`) or service name | |
| No silent empty value | Build does not proceed with `https://:****@...` style URLs without warning | |
| Warning text visible before build output | Warning appears before Docker build log lines | |

**Cleanup:**
```bash
dvm delete credential MISSING_CRED --ecosystem <your-ecosystem> --force
```

---

### Scenario 58: Build with Env Var Fallback for Failed Keychain Credential

Verify that when a keychain lookup fails but a matching environment variable is set on the host, the env var value is used as a fallback — and no warning is displayed (the "env vars always win" contract).

**Setup:** Set an env var on the host that matches a credential's variable name, and create a credential whose keychain service does not exist.

```bash
# Set the env var on the host
export GITHUB_USERNAME=my-github-user

# Create a credential that would normally pull from keychain
dvm create credential GITHUB_USERNAME \
  --source keychain \
  --service nonexistent.service \
  --ecosystem <your-ecosystem>

# Run a build — GITHUB_USERNAME should be rescued from the env var
dvm build --workspace <your-workspace>
```

| Test | Expected | Result |
|------|----------|--------|
| Build succeeds | Exit code 0, no credential-related errors | |
| No warning displayed | No `render.Warning()` output for `GITHUB_USERNAME` | |
| Credential value is non-empty | Build args include `GITHUB_USERNAME=my-github-user` (value redacted in log) | |
| Env var rescue is transparent | Normal build output, no indication that keychain failed | |
| Unrescued credentials still warn | If a second credential has no env var fallback, its warning still appears | |

**Cleanup:**
```bash
dvm delete credential GITHUB_USERNAME --ecosystem <your-ecosystem> --force
unset GITHUB_USERNAME
```

---

### Scenario 59: Dual-Field Keychain Credential Build — Both Fields Resolve

Verify that a dual-field keychain credential correctly resolves both `GITHUB_USERNAME` (account field) and `GITHUB_PAT` (password field) with the `-a $USER` filter active, ensuring the correct system-user Keychain entry is used.

> **Note:** Requires a macOS Keychain entry for `github.com` with both a username (account) and password set for the current system user.

```bash
# Create the dual-field credential (if not already created)
dvm create credential github-creds \
  --source keychain --service github.com \
  --username-var GITHUB_USERNAME \
  --password-var GITHUB_PAT \
  --ecosystem <your-ecosystem>

# Run a build and observe both build args
dvm build --workspace <your-workspace>
```

**Inspect build output for both args:**

```bash
# Both build args should be present with non-empty (redacted) values
dvm build --workspace <your-workspace> 2>&1 | grep -E 'build-arg (GITHUB_USERNAME|GITHUB_PAT)'
```

**Expected build output (values redacted):**
```
--build-arg GITHUB_USERNAME=***REDACTED***
--build-arg GITHUB_PAT=***REDACTED***
```

```bash
# After build: attach and verify both vars inside the container
dvm attach --workspace <your-workspace>
```

**Inside the container, verify:**
```bash
echo "GITHUB_USERNAME=${GITHUB_USERNAME}"   # Expected: non-empty username from Keychain
echo "GITHUB_PAT=${GITHUB_PAT}"             # Expected: non-empty password/token from Keychain
exit
```

| Test | Expected | Result |
|------|----------|--------|
| Build succeeds | Exit code 0, image built | |
| No warnings for `GITHUB_USERNAME` | No warning output for username credential | |
| No warnings for `GITHUB_PAT` | No warning output for password credential | |
| `GITHUB_USERNAME` build arg present | `--build-arg GITHUB_USERNAME=...` in build log | |
| `GITHUB_PAT` build arg present | `--build-arg GITHUB_PAT=...` in build log | |
| Both values non-empty | No `https://:****@...` URL pattern in build output | |
| `GITHUB_USERNAME` available in container | Non-empty value from Keychain account field | |
| `GITHUB_PAT` available in container | Non-empty value from Keychain password field | |

**Cleanup:**
```bash
dvm delete credential github-creds --ecosystem <your-ecosystem> --force
```

---

### Scenario 60: Automated Test Coverage (Regression Gate)

Run the automated tests added in v0.38.2 to confirm all 14 new test functions pass.

```bash
# Warning return tests (build_helpers)
go test ./cmd/... -v -run 'TestLoadBuildCredentials'

# Env var rescue tests (credentials)
go test ./config/... -v -run 'TestResolveCredentials'

# keychainExitError helper tests
go test ./config/... -v -run 'TestKeychainExitError'
```

| Test | Expected | Result |
|------|----------|--------|
| `cmd/build_helpers_test.go` (5 functions) | All pass ✅ | |
| `config/credentials_test.go` (5 new functions) | All pass ✅ | |
| `config/keychain_darwin_test.go` (4 functions) | All pass ✅ | |
| Full test suite | `go test ./...` all green ✅ | |

```bash
# Full regression check
go test ./...
```

---

## Part 17: MaestroVault Integration (v0.40.0)

> **Prerequisite**: MaestroVault installed (`brew install rmkohlman/tap/maestrovault`), `mav serve` running, `MAV_TOKEN` set

### Scenario 61: Create Credential with Vault Source

**Goal**: Create a vault-sourced credential and verify it stores correctly

```bash
# Store a test secret in MaestroVault first
mav set github-pat default "ghp_test123" --metadata "type=pat"

# Create credential referencing the vault secret
dvm create credential github-creds \
  --source vault \
  --vault-secret github-pat \
  --vault-environment default \
  --env-var GITHUB_PAT \
  --ecosystem myorg

# Verify
dvm get credential github-creds --ecosystem myorg
```

**Expected**: Credential created with `source: vault`, `vaultSecret: github-pat`, `vaultEnvironment: default`

### Scenario 62: Vault Dual-Field Credential

**Goal**: Create a dual-field credential with username and password from vault

```bash
# Store secrets
mav set github-username default "rmkohlman"
mav set github-token default "ghp_abc123"

# Create dual-field credential
dvm create credential github-auth \
  --source vault \
  --vault-secret github-token \
  --vault-username-secret github-username \
  --username-var GITHUB_USERNAME \
  --password-var GITHUB_PAT \
  --ecosystem myorg

# Verify
dvm get credential github-auth --ecosystem myorg
```

**Expected**: Shows both `vaultSecret` and `vaultUsernameSecret` fields

### Scenario 63: Build Resolves Vault Credentials

**Goal**: `dvm build` resolves credentials from MaestroVault

```bash
# Ensure MAV_TOKEN is set
export MAV_TOKEN=mvt_your_token

# Run a build with verbose logging
dvm build -v

# Check output for vault credential resolution
```

**Expected**: Build log shows "resolved build credentials" with vault-sourced values (no warnings)

### Scenario 64: Auto-Start Vault Daemon

**Goal**: If vault daemon is not running, `dvm build` auto-starts it

```bash
# Stop the vault daemon
mav stop  # or kill the process

# Run a build — daemon should auto-start
MAV_TOKEN=mvt_your_token dvm build -v
```

**Expected**: Build succeeds; `mav serve --no-touchid` is started automatically

### Scenario 65: Missing MAV_TOKEN Degrades Gracefully

**Goal**: Without `MAV_TOKEN`, vault credentials warn but env var rescue works

```bash
# Unset MAV_TOKEN
unset MAV_TOKEN

# Set env var fallback
export GITHUB_PAT=ghp_fallback_value

# Run build
dvm build -v
```

**Expected**: Warning about failed vault resolution, but credential rescued by env var

### Scenario 66: YAML Apply with Vault Fields

**Goal**: Apply a credential YAML with vault fields

```bash
cat <<EOF | dvm apply -f -
apiVersion: dvm/v1
kind: Credential
metadata:
  name: vault-cred-test
  ecosystem: myorg
spec:
  source: vault
  vaultSecret: my-api-key
  vaultEnvironment: production
  envVar: API_KEY
EOF

dvm get credential vault-cred-test --ecosystem myorg
```

**Expected**: Credential created with vault fields populated

### Scenario 67: Old Keychain Source Rejected

**Goal**: Verify `source: keychain` is rejected

```bash
cat <<EOF | dvm apply -f -
apiVersion: dvm/v1
kind: Credential
metadata:
  name: old-keychain-cred
  ecosystem: myorg
spec:
  source: keychain
  service: github.com
  envVar: GITHUB_PAT
EOF
```

**Expected**: Error — `source: keychain` is no longer valid

### Scenario 68: DB Migration Verification

**Goal**: Verify migrations 013+014 applied correctly on upgrade

```bash
# After upgrading to v0.40.0
dvm admin migrate

# Check that old keychain credentials were migrated
dvm get credentials -A
```

**Expected**: All previously `source: keychain` credentials show as `source: vault` with `vaultSecret` populated

### Scenario 69: Automated Test Coverage (Regression Gate)

```bash
cd ~/Developer/tools/devopsmaestro
go test $(go list ./... | grep -v integration_test) -short -count=1
```

**Expected**: All 60 packages pass, 0 failures

---

## Part 18: MaestroVault Fields Integration (v0.41.0)

> **Prerequisite**: MaestroVault v0.7.0+ installed (`brew install rmkohlman/tap/maestrovault`), `mav serve` running, `MAV_TOKEN` set, vault secret with multiple fields (e.g., `mav set github-multi default --field username=rmkohlman --field pat=ghp_abc123`)

### Scenario 70: Create Credential with Vault Fields (CLI)

**Goal**: Create a credential using `--vault-field` in both mapped and shorthand syntax, verify storage

```bash
# Store a multi-field secret in MaestroVault
mav set github-multi default --field username=rmkohlman --field pat=ghp_abc123

# Create credential using mapped syntax
dvm create credential github-fields \
  --source vault \
  --vault-secret github-multi \
  --vault-environment default \
  --vault-field GITHUB_USERNAME=username \
  --vault-field GITHUB_PAT=pat \
  --ecosystem myorg

# Verify display
dvm get credential github-fields --ecosystem myorg
```

**Expected**: Credential created; `dvm get credential` shows vault field mappings:
- `GITHUB_PAT` → `pat`
- `GITHUB_USERNAME` → `username`
(displayed in sorted alphabetical order)

```bash
# Also test shorthand syntax (field name == env var name)
dvm create credential github-shorthand \
  --source vault \
  --vault-secret github-multi \
  --vault-environment default \
  --vault-field GITHUB_USERNAME \
  --vault-field GITHUB_PAT \
  --ecosystem myorg

dvm get credential github-shorthand --ecosystem myorg
```

**Expected**: Field mapping stored as `GITHUB_USERNAME` → `GITHUB_USERNAME`, `GITHUB_PAT` → `GITHUB_PAT`

### Scenario 71: Vault Fields Mutual Exclusivity

**Goal**: Verify `--vault-field` cannot be combined with legacy dual-field flags

```bash
# Attempt 1: --vault-field + --username-var
dvm create credential bad-cred-1 \
  --source vault \
  --vault-secret github-multi \
  --vault-field GITHUB_PAT=pat \
  --username-var GITHUB_USERNAME \
  --ecosystem myorg
```

**Expected**: Error — `--vault-field` is mutually exclusive with `--username-var`/`--password-var`

```bash
# Attempt 2: --vault-field + --vault-username-secret
dvm create credential bad-cred-2 \
  --source vault \
  --vault-secret github-multi \
  --vault-field GITHUB_PAT=pat \
  --vault-username-secret github-username \
  --ecosystem myorg
```

**Expected**: Error — `--vault-field` is mutually exclusive with `--vault-username-secret`

### Scenario 72: YAML Apply with Vault Fields

**Goal**: Apply a credential YAML with a `vaultFields` map and verify field mappings are stored

```bash
cat <<EOF | dvm apply -f -
apiVersion: dvm/v1
kind: Credential
metadata:
  name: yaml-vault-fields
  ecosystem: myorg
spec:
  source: vault
  vaultSecret: github-multi
  vaultEnvironment: default
  vaultFields:
    GITHUB_USERNAME: username
    GITHUB_PAT: pat
    GITHUB_EMAIL: email
EOF

dvm get credential yaml-vault-fields --ecosystem myorg
```

**Expected**: Credential created; field mappings show:
- `GITHUB_EMAIL` → `email`
- `GITHUB_PAT` → `pat`
- `GITHUB_USERNAME` → `username`

### Scenario 73: YAML Mutual Exclusivity Validation

**Goal**: Verify YAML apply rejects `vaultFields` combined with legacy dual-field YAML syntax

```bash
cat <<EOF | dvm apply -f -
apiVersion: dvm/v1
kind: Credential
metadata:
  name: bad-yaml-cred
  ecosystem: myorg
spec:
  source: vault
  vaultSecret: github-multi
  vaultFields:
    GITHUB_PAT: pat
  usernameVar: GITHUB_USERNAME
EOF
```

**Expected**: Error — `vaultFields` and `usernameVar`/`passwordVar` are mutually exclusive

```bash
# Also test vaultFields + vaultUsernameSecret
cat <<EOF | dvm apply -f -
apiVersion: dvm/v1
kind: Credential
metadata:
  name: bad-yaml-cred-2
  ecosystem: myorg
spec:
  source: vault
  vaultSecret: github-multi
  vaultFields:
    GITHUB_PAT: pat
  vaultUsernameSecret: github-username
EOF
```

**Expected**: Error — `vaultFields` and `vaultUsernameSecret` are mutually exclusive

### Scenario 74: Max 50 Fields Cap

**Goal**: Creating a credential with 51 vault fields is rejected

```bash
# Build a create command with 51 --vault-field flags
ARGS=""
for i in $(seq 1 51); do
  ARGS="$ARGS --vault-field FIELD_${i}=field_${i}"
done

dvm create credential too-many-fields \
  --source vault \
  --vault-secret big-secret \
  $ARGS \
  --ecosystem myorg
```

**Expected**: Error — credential exceeds the maximum of 50 vault fields

### Scenario 75: Build Resolves Vault Fields

**Goal**: `dvm build -v` resolves field-level vault secrets and injects them as env vars

```bash
# Ensure credential with vault fields is configured for the current workspace
# (use yaml-vault-fields from Scenario 72 or github-fields from Scenario 70)

export MAV_TOKEN=mvt_your_token

# Run build with verbose output
dvm build -v
```

**Expected**: Build log shows field-level credential resolution — each `vaultFields` entry injected as a separate env var. No warnings about unresolved credentials.

### Scenario 76: Vault Fields Display

**Goal**: `dvm get credential` shows vault field mappings sorted alphabetically

```bash
# Create a credential with intentionally unsorted fields
dvm create credential display-test \
  --source vault \
  --vault-secret github-multi \
  --vault-field ZEBRA_VAR=z_field \
  --vault-field APPLE_VAR=a_field \
  --vault-field MIDDLE_VAR=m_field \
  --ecosystem myorg

dvm get credential display-test --ecosystem myorg
```

**Expected**: Output lists field mappings in alphabetical order regardless of insertion order:
- `APPLE_VAR` → `a_field`
- `MIDDLE_VAR` → `m_field`
- `ZEBRA_VAR` → `z_field`

### Scenario 77: Automated Test Coverage (Regression Gate)

```bash
cd ~/Developer/tools/devopsmaestro
go test $(go list ./... | grep -v integration_test) -short -count=1
```

**Expected**: All 57 packages pass, 0 failures

---

## Part 19: Auto-Token Creation (v0.43.0)

> **Prerequisite**: MaestroVault (`mav`) available in PATH; a workspace with at least one vault credential configured. For scenarios testing graceful degradation, temporarily rename/remove `mav` from PATH.

### Scenario 78: Auto-Token Creation Without MAV_TOKEN

**Goal**: When no token source is available but `mav` is in PATH, `dvm build` auto-creates a read-only token and persists it

**Prerequisites**: No `MAV_TOKEN` env var set, no `vault.token` in viper config, no `~/.devopsmaestro/.vault_token` file, `mav` is in PATH

```bash
# Ensure clean state — no existing token sources
unset MAV_TOKEN
rm -f ~/.devopsmaestro/.vault_token

# Run build with verbose output against a workspace that has vault credentials
dvm build -v
```

**Expected**:
- Build log contains `"auto-created MaestroVault token"`
- Token file created at `~/.devopsmaestro/.vault_token`
- Build proceeds with full vault credential resolution (no unresolved credential warnings)

---

### Scenario 79: Token File Reuse

**Goal**: On subsequent builds, the persisted token file is reused — no redundant auto-creation

**Prerequisites**: `~/.devopsmaestro/.vault_token` file exists from Scenario 78 (or any prior auto-creation)

```bash
# Run build again (no MAV_TOKEN env var)
dvm build -v
```

**Expected**:
- Build log does NOT contain `"auto-created MaestroVault token"`
- Vault credential resolution still works (token is read from file)

---

### Scenario 80: MAV_TOKEN Env Takes Priority

**Goal**: `MAV_TOKEN` environment variable takes precedence over the persisted token file

**Prerequisites**: `~/.devopsmaestro/.vault_token` file exists; set `MAV_TOKEN` to a known token value

```bash
export MAV_TOKEN=mvt_env_token_test

# Confirm token file also exists
ls ~/.devopsmaestro/.vault_token

dvm build -v
```

**Expected**:
- Vault resolution uses the `MAV_TOKEN` value (`mvt_env_token_test`), not the file token
- No `"auto-created MaestroVault token"` log message (env var satisfied the resolution chain at step 1)

---

### Scenario 81: Graceful Degradation Without mav

**Goal**: When `mav` is not in PATH and no token exists from any source, build continues without vault rather than failing

**Prerequisites**: `mav` not in PATH (e.g., temporarily rename `mav` binary), no `MAV_TOKEN`, no `~/.devopsmaestro/.vault_token`

```bash
# Remove token sources
unset MAV_TOKEN
rm -f ~/.devopsmaestro/.vault_token

# Simulate mav not in PATH (restore when done)
# Option A: temporarily move it
mv $(which mav) /tmp/mav_backup

dvm build -v

# Restore mav
mv /tmp/mav_backup $(dirname $(which dvm))/mav
```

**Expected**:
- Build log contains `"failed to auto-create vault token"` (or similar degradation message)
- Build does NOT abort — it continues to completion
- Credentials sourced from environment variables still resolve correctly
- Vault-sourced credentials are skipped/warned, not fatal

---

### Scenario 82: Token File Permissions

**Goal**: Auto-created token file has restrictive 0600 permissions (owner read/write only)

**Prerequisites**: Auto-token creation has occurred (run Scenario 78 first if needed)

```bash
ls -la ~/.devopsmaestro/.vault_token
```

**Expected**: File permissions show `-rw-------` (0600), owned by the current user:
```
-rw------- 1 <user> <group> <size> <date> /Users/<user>/.devopsmaestro/.vault_token
```

No group or other read/write/execute bits set.

---

### Scenario 83: Automated Test Coverage (Regression Gate)

**Goal**: All unit tests for vault token resolution pass

```bash
cd ~/Developer/tools/devopsmaestro
go test ./config/ -run TestResolveVaultToken -v -count=1
```

**Expected**: All 21 tests pass, 0 failures

---

## Part 20: Build Output Secret Redaction (v0.43.2)

> **Prerequisite**: `dvm` binary built from v0.43.2+. A workspace with at least one vault or build-arg credential configured. For Scenarios 84 and 86, a private GitHub repository accessible via a PAT (`ghp_...`).

### Scenario 84: Build with Private Repo Credentials — PAT Not Visible in Output

**Goal**: Verify that a GitHub PAT passed as a `--build-arg` value is never printed in plain text during a Docker/BuildKit build, even when `pip install` or `go get` downloads from a private GitHub URL containing the token.

**Prerequisites**: A Python workspace whose `requirements.txt` includes a private GitHub dependency using `${GITHUB_PAT}` token substitution.

```bash
# Create a credential with a known PAT value
dvm create credential github-creds \
  --source vault \
  --vault-secret github-pat \
  --vault-field GITHUB_PAT=pat \
  --ecosystem <your-ecosystem>

# Run the build and capture all output
dvm build --workspace <your-workspace> 2>&1 | tee /tmp/build-output.txt
```

**Inspect the captured output:**

```bash
# The raw PAT value must NOT appear anywhere in the output
grep "ghp_" /tmp/build-output.txt        # Expected: no matches

# The redaction marker must appear instead of the PAT value
grep '\*\*\*' /tmp/build-output.txt      # Expected: at least one match

# The download URL structure should be visible but with credentials replaced
grep "github.com" /tmp/build-output.txt  # Expected: URL present, but no token embedded
```

| Test | Expected | Result |
|------|----------|--------|
| Raw PAT not in build output | `grep "ghp_" /tmp/build-output.txt` returns 0 matches | |
| Redaction marker present | `grep '\*\*\*' /tmp/build-output.txt` returns at least 1 match | |
| Build succeeds | Exit code 0, image built successfully | |
| No credential warnings | No `render.Warning()` output for resolved credentials | |

**Cleanup:**
```bash
rm /tmp/build-output.txt
dvm delete credential github-creds --ecosystem <your-ecosystem> --force
```

---

### Scenario 85: Build Without Credentials — Output Unchanged (Zero Overhead)

**Goal**: Verify that when a workspace has no credentials configured, the `RedactingWriter` fast path is taken — the inner writer is used directly with no redaction wrapping — and build output is identical to pre-v0.43.2 behavior.

**Prerequisites**: A workspace with no credentials in its hierarchy (no ecosystem/domain/app/workspace-scoped credentials).

```bash
# Ensure no credentials exist in the workspace hierarchy
dvm get credentials -A

# Run a standard build and capture output
dvm build --workspace <credential-free-workspace> 2>&1 | tee /tmp/build-no-creds.txt
```

**Verify build output:**

```bash
# No redaction markers should appear in a credential-free build
grep -c '\*\*\*' /tmp/build-no-creds.txt   # Expected: 0

# Standard Docker/BuildKit output should be complete and unmodified
grep "DONE\|Step\|RUN\|FROM" /tmp/build-no-creds.txt
```

| Test | Expected | Result |
|------|----------|--------|
| Build succeeds | Exit code 0, image built | |
| No redaction markers in output | `grep -c '\*\*\*'` returns `0` | |
| Full build output present | Docker/BuildKit progress lines visible and complete | |
| No truncation or missing lines | Output is not shorter or different from a known-good build | |

**Cleanup:**
```bash
rm /tmp/build-no-creds.txt
```

---

### Scenario 86: Build with Multiple Credentials — All Redacted

**Goal**: Verify that when a workspace has multiple credentials in its hierarchy (e.g., `GITHUB_PAT`, `GITHUB_USERNAME`, and an `NPM_TOKEN`), all secret values are independently redacted from build output — including partial appearances and appearances split across log lines.

**Prerequisites**: A workspace with at least three credentials configured, each with a distinct secret value of 8+ characters.

```bash
# Create multiple credentials
dvm create credential github-username \
  --source vault \
  --vault-secret github-multi \
  --vault-field GITHUB_USERNAME=username \
  --ecosystem <your-ecosystem>

dvm create credential github-pat \
  --source vault \
  --vault-secret github-multi \
  --vault-field GITHUB_PAT=pat \
  --ecosystem <your-ecosystem>

dvm create credential npm-token \
  --source vault \
  --vault-secret npm-secret \
  --vault-field NPM_TOKEN=token \
  --ecosystem <your-ecosystem>

# Run build and capture all output
dvm build --workspace <your-workspace> 2>&1 | tee /tmp/build-multi-creds.txt
```

**Retrieve the actual secret values from vault (for verification):**

```bash
# Get the values so you can check they are absent from output
GITHUB_PAT_VALUE=$(mav get github-multi default --field pat 2>/dev/null)
GITHUB_USERNAME_VALUE=$(mav get github-multi default --field username 2>/dev/null)
NPM_TOKEN_VALUE=$(mav get npm-secret default --field token 2>/dev/null)
```

**Verify each secret is absent from build output:**

```bash
# None of the raw secret values should appear in build output
grep -F "$GITHUB_PAT_VALUE"      /tmp/build-multi-creds.txt   # Expected: no match
grep -F "$GITHUB_USERNAME_VALUE" /tmp/build-multi-creds.txt   # Expected: no match
grep -F "$NPM_TOKEN_VALUE"       /tmp/build-multi-creds.txt   # Expected: no match

# The redaction marker should appear for each secret that appeared in output
grep -c '\*\*\*' /tmp/build-multi-creds.txt   # Expected: at least 3 matches
```

| Test | Expected | Result |
|------|----------|--------|
| `GITHUB_PAT` value not in output | `grep -F "$GITHUB_PAT_VALUE"` returns 0 matches | |
| `GITHUB_USERNAME` value not in output | `grep -F "$GITHUB_USERNAME_VALUE"` returns 0 matches | |
| `NPM_TOKEN` value not in output | `grep -F "$NPM_TOKEN_VALUE"` returns 0 matches | |
| Redaction marker present for each secret | `grep -c '\*\*\*'` returns 3 or more matches | |
| Build succeeds | Exit code 0, image built successfully | |
| All credentials resolve without warnings | No `render.Warning()` output for any credential | |

**Cleanup:**
```bash
rm /tmp/build-multi-creds.txt
dvm delete credential github-username --ecosystem <your-ecosystem> --force
dvm delete credential github-pat      --ecosystem <your-ecosystem> --force
dvm delete credential npm-token       --ecosystem <your-ecosystem> --force
```

---

---

## Part 21: Container Neovim Environment (v0.44.0)

> **Prerequisite**: `dvm` binary built from v0.44.0+. A workspace configured with Neovim (nvim plugin set assigned). Docker or Colima containerd runtime available.

### Scenario 87: Python Workspace — Node 22+ and Mason Installs Without Errors

**Goal**: Verify that a Python workspace container has Node 22 or later available, and that Mason completes tool installation without errors during the first Neovim launch.

**Prerequisites**: A Python workspace with a Neovim plugin set. `dvm build` must be run fresh (no cached image).

```bash
# Build the Python workspace image (no cache to force fresh install)
dvm build --workspace <your-python-workspace> --no-cache

# Attach to the container
dvm attach --workspace <your-python-workspace>
```

Inside the container:

```bash
# Verify Node.js version is 22 or later
node --version
# Expected: v22.x.x or higher

# Launch Neovim and observe Mason output
nvim .
# Wait 30–60 seconds for Mason to complete installation on first launch
# Press :MasonLog to inspect the Mason log
```

| Test | Expected | Result |
|------|----------|--------|
| `node --version` is 22+ | Output is `v22.x.x` or higher | |
| Mason opens without error popups | No red error notifications on first launch | |
| `pylint` available in container | `which pylint` returns a path | |
| `:MasonLog` shows no ENOENT errors | No "ENOENT" or "not found" lines in Mason log | |
| Neovim exits cleanly | `:qa` exits with code 0 | |

---

### Scenario 88: Go Workspace — gopls and goimports Installed via Mason

**Goal**: Verify that a Go workspace container has both `gopls` (LSP) and `goimports` (formatter) installed and available to Neovim via Mason.

**Prerequisites**: A Go workspace with a Neovim plugin set. `dvm build` run fresh.

```bash
# Build the Go workspace image
dvm build --workspace <your-go-workspace> --no-cache

# Attach to the container
dvm attach --workspace <your-go-workspace>
```

Inside the container:

```bash
# Verify gopls is available (Mason-installed)
which gopls || mason-tool-show gopls
# Expected: a path or Mason status showing "installed"

# Verify goimports is available
which goimports
# Expected: a path inside the Mason bin directory

# Open a .go file and confirm LSP attaches without error
nvim main.go
# Expected: no "LSP not found" or ENOENT errors; gopls attaches
```

| Test | Expected | Result |
|------|----------|--------|
| `which gopls` returns a path | Mason-installed `gopls` binary found | |
| `which goimports` returns a path | Mason-installed `goimports` binary found | |
| LSP attaches in Neovim for `.go` file | No "LSP not found" notification | |
| No Mason ENOENT errors in `:MasonLog` | Mason log is clean | |

---

### Scenario 89: Missing Linter Binary — No ENOENT Crash on BufEnter

**Goal**: Verify that when a linter binary (pylint or shellcheck) is absent from the container, opening a buffer in Neovim does not produce an ENOENT error or fill the message history with errors.

**Prerequisites**: A workspace where the linter binary is intentionally absent (e.g., a Node.js workspace where pylint is not installed, or a custom image where shellcheck is removed after build).

```bash
dvm attach --workspace <your-workspace>
```

Inside the container:

```bash
# Confirm shellcheck is absent (use a workspace type where it is not installed)
which shellcheck 2>/dev/null || echo "shellcheck not found"

# Open a file and switch buffers repeatedly to trigger BufEnter
nvim .
# Inside Neovim: open several files with :e, :bn, :bp
```

| Test | Expected | Result |
|------|----------|--------|
| `which shellcheck` returns "not found" | shellcheck absent from PATH | |
| No ENOENT error notification on buffer switch | Neovim message area stays clean | |
| `:messages` shows no "ENOENT" lines | No executable-not-found errors logged | |
| Neovim remains functional | Can open, edit, and close files normally | |

---

### Scenario 90: No Mason ensure_installed Errors on Neovim Launch

**Goal**: Verify that Neovim launches inside a container without attempting to download any LSP servers or Mason tools at runtime — all tools should already be present from the build stage.

**Prerequisites**: Any workspace with Neovim configured. The container must have been built with `dvm build` from v0.44.0+.

```bash
dvm build --workspace <your-workspace> --no-cache
dvm attach --workspace <your-workspace>
```

Inside the container:

```bash
# Launch Neovim and capture startup messages
nvim --headless -c "sleep 5" -c "qa" 2>&1 | tee /tmp/nvim-startup.txt

# Inspect for Mason download or ensure_installed activity
grep -i "installing\|downloading\|ensure_installed" /tmp/nvim-startup.txt || echo "No runtime installs detected"
```

Also verify interactively:

```bash
nvim .
# Observe: no Mason "Installing..." progress bar on first launch
# :checkhealth mason  — all tools should report "installed"
```

| Test | Expected | Result |
|------|----------|--------|
| No "Installing..." progress bar on Neovim launch | Mason does not attempt runtime installs | |
| `:checkhealth mason` shows all tools installed | No tools in "not installed" state | |
| No "ensure_installed" references in startup output | `grep ensure_installed` returns 0 matches | |
| Neovim startup completes in under 5 seconds | No network-wait delay on open | |
| auto-session restores last session cleanly | Previous session files reopen without error | |

---

## Part 22: Registry Startup Resilience (v0.45.0)

> **Prerequisite**: `dvm` binary built from v0.45.0+. At least one registry (Athens, Zot, or Devpi) already running at its configured port. `dvm admin init` completed.

### Scenario 91: Athens Already Running — Adopted Without Error

**Goal**: Verify that `dvm start registry` adopts a healthy, already-running Athens instance instead of returning a port-in-use error.

**Prerequisites**: Athens registry already started (e.g., from a previous `dvm start registry` invocation that was not stopped).

```bash
# Confirm Athens is listening on its configured port
curl -s http://localhost:3000/healthz
# Expected: HTTP 200

# Attempt to start the registry again
dvm start registry --name athens-local
```

| Test | Expected | Result |
|------|----------|--------|
| `curl /healthz` returns 200 | Athens is healthy and already running | |
| `dvm start registry` exits without error | No "port in use" error message | |
| `dvm get registries` shows status Running | Athens registry listed as Running | |
| No second Athens process spawned | `pgrep athens` returns one PID | |

---

### Scenario 92: Zot Checksum Download — Correct Manifest URL

**Goal**: Verify that installing or upgrading Zot downloads the correct checksum by fetching the `checksums.sha256.txt` manifest rather than appending `.sha256` to the binary URL.

**Prerequisites**: Zot binary not yet downloaded (or `--version` set to a version not yet cached). Internet access available.

```bash
# Remove any cached Zot binary to force a fresh download
rm -f ~/.devopsmaestro/bin/zot*

# Start a Zot registry (triggers download)
dvm start registry --name zot-local
```

| Test | Expected | Result |
|------|----------|--------|
| Download completes without checksum 404 error | No "404 Not Found" in output | |
| Zot binary passes checksum verification | No "checksum mismatch" error | |
| `dvm get registries` shows Zot as Running | Zot registry listed as Running | |
| `~/.devopsmaestro/bin/zot` binary is present | File exists and is executable | |

---

### Scenario 93: Devpi Pip Fallback — Installs via pip When pipx Absent

**Goal**: Verify that Devpi installation succeeds on a system where `pipx` is not available, falling back to `python3 -m pip install --user`.

**Prerequisites**: A test environment where `pipx` is absent from PATH but `python3` and `pip` are available.

```bash
# Simulate missing pipx (rename or unlink temporarily)
which pipx && sudo mv $(which pipx) /tmp/pipx-backup || echo "pipx not found — good"

# Attempt to start a Devpi registry (triggers installation)
dvm start registry --name devpi-local
```

| Test | Expected | Result |
|------|----------|--------|
| No "pipx not found" fatal error | Fallback pip install attempted | |
| `devpi-server` installed in Python user base | `python3 -m devpi-server --version` returns a version | |
| Devpi registry starts successfully | `dvm get registries` shows Devpi as Running | |
| Restore pipx afterward | `sudo mv /tmp/pipx-backup $(dirname $(which python3))/pipx` | |

---

### Scenario 94: Full Registry Start Verification — All 5 Registries

**Goal**: Verify that all 5 supported registry types (Squid, Verdaccio, Athens, Zot, Devpi) can be started cleanly from a stopped state and listed by `dvm get registries`.

**Prerequisites**: All registries stopped. All required binaries available or downloadable.

```bash
# Stop any running registries
dvm stop registry --all

# Start each registry
dvm start registry --name squid-local
dvm start registry --name verdaccio-local
dvm start registry --name athens-local
dvm start registry --name zot-local
dvm start registry --name devpi-local

# List all registries
dvm get registries
```

| Test | Expected | Result |
|------|----------|--------|
| `dvm start registry --name squid-local` exits 0 | Squid starts without error | |
| `dvm start registry --name verdaccio-local` exits 0 | Verdaccio starts without error | |
| `dvm start registry --name athens-local` exits 0 | Athens starts without error | |
| `dvm start registry --name zot-local` exits 0 | Zot starts without error | |
| `dvm start registry --name devpi-local` exits 0 | Devpi starts without error | |
| `dvm get registries` shows all 5 as Running | All 5 registries listed as Running | |

---

## Part 23: Auto-Detect Git Default Branch (v0.49.0)

> **Prerequisite**: `dvm` binary built from v0.49.0+. Internet access available (scenarios 95–97 query GitHub). `dvm admin init` completed.

### Scenario 95: GitRepo auto-detects master branch

```bash
# Using a well-known repo that defaults to 'master'
dvm create gitrepo test-master --url https://github.com/git/git.git --no-sync
dvm get gitrepo test-master -o yaml
# Expected: defaultRef should be "master" (detected, not hardcoded "main")
dvm delete gitrepo test-master
```

| Test | Expected | Result |
|------|----------|--------|
| `dvm create gitrepo` exits 0 | GitRepo created without error | |
| `dvm get gitrepo test-master -o yaml` shows `defaultRef: master` | Detection resolved `master`, not `main` | |
| `dvm delete gitrepo test-master` exits 0 | GitRepo deleted cleanly | |

---

### Scenario 96: GitRepo auto-detects main branch

```bash
dvm create gitrepo test-main --url https://github.com/rmkohlman/devopsmaestro.git --no-sync
dvm get gitrepo test-main -o yaml
# Expected: defaultRef should be "main" (detected correctly)
dvm delete gitrepo test-main
```

| Test | Expected | Result |
|------|----------|--------|
| `dvm create gitrepo` exits 0 | GitRepo created without error | |
| `dvm get gitrepo test-main -o yaml` shows `defaultRef: main` | Detection resolved `main` correctly | |
| `dvm delete gitrepo test-main` exits 0 | GitRepo deleted cleanly | |

---

### Scenario 97: GitRepo --default-ref flag overrides detection

```bash
dvm create gitrepo test-override --url https://github.com/rmkohlman/devopsmaestro.git --default-ref develop --no-sync
dvm get gitrepo test-override -o yaml
# Expected: defaultRef should be "develop" (flag override, not auto-detected "main")
dvm delete gitrepo test-override
```

| Test | Expected | Result |
|------|----------|--------|
| `dvm create gitrepo` exits 0 | GitRepo created without error | |
| `dvm get gitrepo test-override -o yaml` shows `defaultRef: develop` | Flag override respected over auto-detection | |
| `dvm delete gitrepo test-override` exits 0 | GitRepo deleted cleanly | |

---

### Scenario 98: App with --repo auto-detects default branch

```bash
# Create app with --repo URL (auto-creates GitRepo)
dvm create ecosystem test-eco
dvm use ecosystem test-eco
dvm create domain test-dom
dvm use domain test-dom
dvm create app test-app --repo https://github.com/git/git.git
dvm get gitrepo git-git -o yaml
# Expected: defaultRef should be "master" (auto-detected for auto-created GitRepo)
# Cleanup
dvm delete app test-app
dvm delete gitrepo git-git
dvm delete domain test-dom
dvm delete ecosystem test-eco
```

| Test | Expected | Result |
|------|----------|--------|
| `dvm create app test-app --repo` exits 0 | App and GitRepo created without error | |
| `dvm get gitrepo git-git -o yaml` shows `defaultRef: master` | Auto-created GitRepo detected `master`, not `main` | |
| Cleanup commands all exit 0 | All resources deleted cleanly | |

---

## Part 24: Scoped Hierarchical Views (v0.52.0)

> **Prerequisite**: `dvm` binary built from v0.52.0+. At least one ecosystem, domain, app, and workspace configured. `dvm admin init` completed.

New flags on `dvm get all`: `-e/--ecosystem`, `-d/--domain`, `-a/--app`, `-A/--all`

---

### Scenario 99: Default Scoping to Active Context

```bash
# Setup: ensure active context is set
dvm use ecosystem <eco-name>
dvm use domain <domain-name>
dvm use app <app-name>

# Get all — should only show resources in the active scope
dvm get all
```

| Test | Expected | Result |
|------|----------|--------|
| Only active ecosystem shown | Ecosystems section shows only the active ecosystem | |
| Domains filtered | Only domains in the active ecosystem shown | |
| Apps filtered | Only apps in the active domain shown | |
| Workspaces filtered | Only workspaces for the active app shown | |
| Global resources shown | Registries, git repos, nvim plugins, nvim themes always shown | |

---

### Scenario 100: Explicit Ecosystem Flag

```bash
dvm get all -e <eco-name>
```

| Test | Expected | Result |
|------|----------|--------|
| Only specified ecosystem shown | Ecosystems section shows only <eco-name> | |
| Domains filtered to ecosystem | Only domains belonging to <eco-name> | |
| Apps filtered | Only apps in those domains | |

---

### Scenario 101: Explicit Ecosystem + Domain

```bash
dvm get all -e <eco-name> -d <domain-name>
```

| Test | Expected | Result |
|------|----------|--------|
| Only specified domain shown | Domains section shows only <domain-name> | |
| Apps filtered to domain | Only apps in <domain-name> | |
| Workspaces filtered | Only workspaces for apps in <domain-name> | |

---

### Scenario 102: Explicit Ecosystem + Domain + App

```bash
dvm get all -e <eco-name> -d <domain-name> -a <app-name>
```

| Test | Expected | Result |
|------|----------|--------|
| Only specified app shown | Apps section shows only <app-name> | |
| Workspaces filtered to app | Only workspaces for <app-name> | |
| Credentials filtered | Only credentials scoped to the app or its parents | |

---

### Scenario 103: Show All Flag (-A)

```bash
dvm get all -A
```

| Test | Expected | Result |
|------|----------|--------|
| All ecosystems shown | Every ecosystem regardless of context | |
| All domains shown | Every domain across all ecosystems | |
| All resources shown | All apps, workspaces, credentials | |

---

### Scenario 104: -A Combined with Scope Flags (Error)

```bash
dvm get all -A -e <eco-name>
```

| Test | Expected | Result |
|------|----------|--------|
| Command fails | Error message shown | |
| Error message | "--all (-A) cannot be combined with scoping flags (-e, -d, -a)" | |
| No stack trace | Clean error output | |

---

### Scenario 105: Domain Without Ecosystem (Error)

```bash
# Clear context first
dvm use --clear
dvm get all -d <domain-name>
```

| Test | Expected | Result |
|------|----------|--------|
| Command fails | Error message shown | |
| Error message | "--domain requires an ecosystem. Use -e <name> or 'dvm use ecosystem <name>'" | |
| No stack trace | Clean error output | |

---

### Scenario 106: No Context, No Flags (Discovery Mode)

```bash
# Clear all context
dvm use --clear
dvm get all
```

| Test | Expected | Result |
|------|----------|--------|
| All resources shown | Same as -A behavior | |
| No error | Command succeeds | |

---

### Scenario 107: Scoped Output Formats

```bash
dvm get all -e <eco-name> -o json
dvm get all -e <eco-name> -o yaml
dvm get all -e <eco-name> -o wide
```

| Test | Expected | Result |
|------|----------|--------|
| JSON output filtered | Only resources in scope in JSON output | |
| YAML output filtered | Only resources in scope in YAML output | |
| Wide output filtered | Only resources in scope with extra columns | |

---

---

## Part 25: List Format Export (v0.53.0)

> **Prerequisite**: `dvm` binary built from v0.53.0+. At least one ecosystem, domain, app, and workspace configured with data. `dvm admin init` completed.

`dvm get all -o yaml` and `-o json` now produce a `kind: List` document. `dvm apply -f` now accepts `kind: List` documents.

---

### Scenario 108: Export All Resources as YAML List

Verify that `dvm get all -A -o yaml` produces a valid `kind: List` document.

```bash
dvm get all -A -o yaml
```

| Test | Expected | Result |
|------|----------|--------|
| Output starts with `kind: List` | First non-empty line is `kind: List` | |
| `apiVersion` present | `apiVersion: devopsmaestro.io/v1` in output | |
| `items:` section present | `items:` key present in YAML | |
| Each item has `kind:` field | Individual `kind: Ecosystem`, `kind: Workspace`, etc. entries under items | |
| Each item has `metadata:` and `spec:` | Full resource YAML per item, not a summary | |
| No `AllResources` flat struct | No top-level `ecosystems:`, `domains:`, `apps:` keys | |

---

### Scenario 109: Export All Resources as JSON List

Verify that `dvm get all -A -o json` produces a valid `kind: List` JSON document.

```bash
dvm get all -A -o json
dvm get all -A -o json | python3 -m json.tool
```

| Test | Expected | Result |
|------|----------|--------|
| Output parses as valid JSON | `python3 -m json.tool` succeeds with no errors | |
| Top-level `kind` is `"List"` | `{"kind": "List", ...}` at the root | |
| `items` is an array | `"items": [...]` present | |
| Each item has full resource fields | `kind`, `apiVersion`, `metadata`, `spec` present on each item | |

---

### Scenario 110: Scoped Export Excludes Global Resources

Verify that scoped YAML/JSON export omits global resources (registries, git repos, nvim resources, terminal resources).

```bash
# Export scoped to a specific ecosystem
dvm get all -e <eco-name> -o yaml
```

| Test | Expected | Result |
|------|----------|--------|
| Only hierarchical resources in output | Ecosystems, Domains, Apps, Workspaces, Credentials matching scope | |
| No `kind: Registry` items | Registry items absent from scoped YAML | |
| No `kind: GitRepo` items | GitRepo items absent from scoped YAML | |
| No `kind: NvimPlugin` or `kind: NvimTheme` | Nvim resources absent from scoped YAML | |
| No `kind: TerminalPrompt` or `kind: TerminalPackage` | Terminal resources absent from scoped YAML | |
| Table output unaffected | `dvm get all -e <eco-name>` (no `-o`) still shows global resources in table | |

---

### Scenario 111: Round-Trip — Export Then Re-Import

Verify that output from `dvm get all -A -o yaml` can be fed back to `dvm apply -f -` without errors.

```bash
# Export all resources
dvm get all -A -o yaml > /tmp/dvm-backup.yaml

# Inspect the file
head -20 /tmp/dvm-backup.yaml

# Apply back (re-apply should be idempotent — update or no-op for existing resources)
dvm apply -f /tmp/dvm-backup.yaml
```

| Test | Expected | Result |
|------|----------|--------|
| Export succeeds | `/tmp/dvm-backup.yaml` created, non-empty | |
| File starts with `kind: List` | First line of file is `kind: List` | |
| Apply succeeds | `dvm apply` completes without fatal errors | |
| Items applied in order | No dependency errors (e.g., Workspace applied before Ecosystem) | |
| Existing resources updated or skipped | No "not found" or "already exists" hard errors | |

**Cleanup:**
```bash
rm /tmp/dvm-backup.yaml
```

---

### Scenario 112: Apply a `kind: List` File

Verify that `dvm apply -f <list-file>` processes all items in a List document individually, with continue-on-error behavior.

```bash
# Create a List YAML with two valid items and one invalid item
cat <<EOF > /tmp/test-list.yaml
kind: List
apiVersion: devopsmaestro.io/v1
items:
  - kind: Ecosystem
    apiVersion: devopsmaestro.io/v1
    metadata:
      name: list-test-eco
    spec: {}
  - kind: InvalidKind
    apiVersion: devopsmaestro.io/v1
    metadata:
      name: bad-item
    spec: {}
  - kind: Domain
    apiVersion: devopsmaestro.io/v1
    metadata:
      name: list-test-domain
      ecosystem: list-test-eco
    spec: {}
EOF

dvm apply -f /tmp/test-list.yaml
```

| Test | Expected | Result |
|------|----------|--------|
| First item (Ecosystem) applied | `list-test-eco` created or updated | |
| Invalid item reports an error | Error message for `InvalidKind` shown | |
| Third item (Domain) still applied | `list-test-domain` created despite previous error | |
| Command exits with non-zero code | At least one item failed, so exit code != 0 | |
| No panic | Clean error output, no stack trace | |

**Cleanup:**
```bash
rm /tmp/test-list.yaml
dvm delete domain list-test-eco/list-test-domain --force
dvm delete ecosystem list-test-eco --force
```

---

### Scenario 113: Pipe Export Directly to Apply

Verify the canonical round-trip pipeline works end-to-end.

```bash
# Pipe export directly to apply
dvm get all -A -o yaml | dvm apply -f -
```

| Test | Expected | Result |
|------|----------|--------|
| Pipeline completes without error | No fatal errors | |
| stdin input accepted by apply | `-f -` reads from stdin correctly | |
| Resources unchanged (idempotent) | Re-applying existing resources does not corrupt state | |

---

### Scenario 114: Dependency Order in List Output

Verify that items in the exported List appear in apply-safe dependency order.

```bash
dvm get all -A -o yaml | grep '^  - kind:' | head -20
```

| Test | Expected | Result |
|------|----------|--------|
| Ecosystems appear before Domains | `kind: Ecosystem` entries precede `kind: Domain` entries | |
| Domains appear before Apps | `kind: Domain` entries precede `kind: App` entries | |
| Apps appear before Workspaces | `kind: App` entries precede `kind: Workspace` entries | |
| GitRepos appear before Credentials and Workspaces | `kind: GitRepo` entries before `kind: Credential` and `kind: Workspace` | |

---

## Part 26: Corporate Build Configuration (v0.54.0)

These scenarios verify the three changes shipped in v0.54.0: CA certificate injection, build args as `ARG` declarations, and the `USER` directive fix.

**Prerequisites:**
- MaestroVault configured with at least one test secret containing a PEM certificate
- An existing app and workspace
- `dvm` binary built from v0.54.0 source

---

### Scenario 115: Build Args Emitted as ARG (Not ENV)

Verify that `spec.build.args` keys appear as `ARG` declarations in the generated Dockerfile and are not persisted as `ENV` in the image.

```bash
# Configure a workspace with build args
dvm apply -f - <<EOF
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: test-build-args
  app: my-api
spec:
  build:
    args:
      GITHUB_PAT: test-token-value
      PIP_INDEX_URL: https://user:pass@pypi.corp.example/simple
EOF

# Build and capture the generated Dockerfile (use --no-cache to see fresh output)
dvm build -a my-api -w test-build-args --no-cache 2>&1 | head -80
```

| Test | Expected | Result |
|------|----------|--------|
| Build completes | No fatal errors | |
| `ARG GITHUB_PAT` present | `ARG GITHUB_PAT` line in generated Dockerfile | |
| `ARG PIP_INDEX_URL` present | `ARG PIP_INDEX_URL` line in generated Dockerfile | |
| No `ENV GITHUB_PAT` | `ENV GITHUB_PAT` does NOT appear in generated Dockerfile | |
| No `ENV PIP_INDEX_URL` | `ENV PIP_INDEX_URL` does NOT appear in generated Dockerfile | |

**Cleanup:**
```bash
dvm delete workspace my-api/test-build-args --force
```

---

### Scenario 116: USER Directive Follows container.user

Verify that the generated Dockerfile's `USER` directive reflects `container.user` when set.

```bash
# Configure a workspace with a custom user
dvm apply -f - <<EOF
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: test-custom-user
  app: my-api
spec:
  container:
    user: ray
    uid: 1001
    gid: 1001
EOF

dvm build -a my-api -w test-custom-user --no-cache 2>&1 | grep -i "USER"
```

| Test | Expected | Result |
|------|----------|--------|
| Build completes | No fatal errors | |
| `USER ray` in Dockerfile | `USER ray` directive present (not `USER dev`) | |

**Cleanup:**
```bash
dvm delete workspace my-api/test-custom-user --force
```

---

### Scenario 117: USER Directive Defaults to "dev" When container.user Unset

Verify that the `USER` directive defaults to `dev` when `container.user` is not configured.

```bash
# Configure a workspace with no container.user
dvm apply -f - <<EOF
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: test-default-user
  app: my-api
spec:
  build:
    devStage:
      packages:
        - curl
EOF

dvm build -a my-api -w test-default-user --no-cache 2>&1 | grep -i "USER"
```

| Test | Expected | Result |
|------|----------|--------|
| Build completes | No fatal errors | |
| `USER dev` in Dockerfile | `USER dev` directive present (default) | |

**Cleanup:**
```bash
dvm delete workspace my-api/test-default-user --force
```

---

### Scenario 118: CA Certificate Injected from MaestroVault

Verify that a valid CA certificate is fetched from MaestroVault, written to the build context, and the Dockerfile contains the correct injection statements.

```bash
# Assumes MaestroVault has a secret named "test-ca-cert" with a "cert" field
dvm apply -f - <<EOF
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: test-ca-inject
  app: my-api
spec:
  build:
    caCerts:
      - name: test-corp-ca
        vaultSecret: test-ca-cert
EOF

dvm build -a my-api -w test-ca-inject --no-cache 2>&1
```

| Test | Expected | Result |
|------|----------|--------|
| Build completes | No fatal errors | |
| `COPY certs/` line present | `COPY certs/ /usr/local/share/ca-certificates/custom/` in Dockerfile | |
| `RUN update-ca-certificates` present | Line present in Dockerfile | |
| `SSL_CERT_FILE` set | `ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt` in Dockerfile | |
| `REQUESTS_CA_BUNDLE` set | `ENV REQUESTS_CA_BUNDLE=/etc/ssl/certs/ca-certificates.crt` in Dockerfile | |
| `NODE_EXTRA_CA_CERTS` set | `ENV NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt` in Dockerfile | |

**Cleanup:**
```bash
dvm delete workspace my-api/test-ca-inject --force
```

---

### Scenario 119: Missing CA Certificate Causes Fatal Build Error

Verify that a `caCerts` entry referencing a non-existent MaestroVault secret causes an immediate fatal error rather than a silent degradation.

```bash
dvm apply -f - <<EOF
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: test-ca-missing
  app: my-api
spec:
  build:
    caCerts:
      - name: missing-ca
        vaultSecret: this-secret-does-not-exist
EOF

dvm build -a my-api -w test-ca-missing 2>&1
echo "Exit code: $?"
```

| Test | Expected | Result |
|------|----------|--------|
| Build fails | Non-zero exit code | |
| Error message references vault secret | Error output mentions `this-secret-does-not-exist` or the certificate name | |
| No partial image built | No `dvm-test-ca-missing-my-api` image in `docker images` | |

**Cleanup:**
```bash
dvm delete workspace my-api/test-ca-missing --force
```

---

### Scenario 120: Alpine Image Receives ca-certificates Package Automatically

Verify that when `caCerts` is configured and the base image is Alpine-based, `ca-certificates` is automatically included in the `apk add` package list.

```bash
dvm apply -f - <<EOF
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: test-alpine-ca
  app: my-api
spec:
  image:
    baseImage: python:3.12-alpine
  build:
    caCerts:
      - name: test-corp-ca
        vaultSecret: test-ca-cert
EOF

dvm build -a my-api -w test-alpine-ca --no-cache 2>&1 | grep -i "ca-certificates"
```

| Test | Expected | Result |
|------|----------|--------|
| Build completes | No fatal errors | |
| `ca-certificates` in `apk add` | Line contains `apk add ... ca-certificates ...` | |

**Cleanup:**
```bash
dvm delete workspace my-api/test-alpine-ca --force
```

---

## Test Results Summary

| Part | Tests | Pass | Fail |
|------|-------|------|------|
| Part 1: Setup & Build | 5 segments | | |
| Interactive: Environment | 6 checks | | |
| Interactive: Python | 5 checks | | |
| Interactive: Neovim | 2 checks | | |
| Part 2: Post-Attach | 5 segments | | |
| Part 3: Namespaced Commands | 5 checks | | |
| Part 4: Registry System | 13 scenarios | | |
| Part 5: Package Rename & Auto-Detection | 3 scenarios | | |
| Part 6: Credential Management | 15 scenarios | | |
| Part 7: Registry Version Management | 8 scenarios | | |
| Part 8: Credential Injection & Env Vars | 7 scenarios | | |
| Part 9: Runtime Credential & Env Injection | 7 scenarios | | |
| Part 10: Vault Dual-Field Credentials | 12 scenarios | | |
| Part 11: Registry Bug Fix Verification | 6 scenarios | | |
| Part 12: BuildKit Builder Stage Robustness | 3 scenarios | | |
| Part 13: BuildKit Structural Improvements | 3 scenarios | | |
| Part 14: Dockerfile Generator Purity | 3 scenarios | | |
| Part 15: Python HTTPS Token Substitution | 5 scenarios | | |
| Part 16: Credential Resolution Robustness | 4 scenarios | | |
| Part 17: MaestroVault Integration | 9 scenarios | | |
| Part 18: MaestroVault Fields Integration | 8 scenarios | | |
| Part 19: Auto-Token Creation | 6 scenarios | | |
| Part 20: Build Output Secret Redaction | 3 scenarios | | |
| Part 21: Container Neovim Environment | 4 scenarios | | |
| Part 22: Registry Startup Resilience | 4 scenarios | | |
| Part 23: Auto-Detect Git Default Branch | 4 scenarios | | |
| Part 24: Scoped Hierarchical Views | 9 scenarios | | |
| Part 25: List Format Export | 7 scenarios | | |
| Part 26: Corporate Build Configuration | 6 scenarios | | |

---

## Cleanup

```bash
# Remove test containers
docker ps -aq --filter "name=dvm-" | xargs -r docker rm -f 2>/dev/null || true

# For Colima containerd
colima ssh -- sudo nerdctl --namespace devopsmaestro ps -aq 2>/dev/null | \
  xargs -I {} colima ssh -- sudo nerdctl --namespace devopsmaestro rm -f {} 2>/dev/null || true

# Remove test images
docker images | grep dvm- | awk '{print $3}' | xargs -r docker rmi 2>/dev/null || true

# Full reset
rm -rf ~/.devopsmaestro
rm -rf ~/Developer/sandbox/dvm-test-*
```

---

## Troubleshooting

### `cd` commands fail with "zoxide: no match found"

Use `builtin cd` instead of `cd` to bypass zoxide:
```bash
builtin cd ~/Developer/sandbox/dvm-test-fastapi
```

### Build hangs on "Detecting app language"

Check that the app path is correct:
```bash
sqlite3 ~/.devopsmaestro/devopsmaestro.db "SELECT name, path FROM apps;"
```

If the path is wrong (e.g., `/Users/you` instead of the app dir), re-run Part 1.

### Image not found after build

For Colima containerd, images are in the `devopsmaestro` namespace:
```bash
colima nerdctl -- --namespace devopsmaestro images
```

---

**Tested by:** ________________  
**Date:** ________________  
**Version**: v0.54.0  
**Platform:** ________________
