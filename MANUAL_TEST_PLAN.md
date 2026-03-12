# DevOpsMaestro Manual Test Plan

> **Version**: v0.37.0  
> **Last Updated**: March 2026

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
**Version:** v0.37.0  
**Platform:** ________________
