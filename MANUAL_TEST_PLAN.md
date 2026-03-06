# DevOpsMaestro Manual Test Plan

> **Version**: v0.34.0  
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

## Test Results Summary

| Section | Tests | Pass | Fail |
|---------|-------|------|------|
| Part 1: Setup & Build | 5 segments | | |
| Interactive: Environment | 6 checks | | |
| Interactive: Python | 5 checks | | |
| Interactive: Neovim | 2 checks | | |
| Part 2: Post-Attach | 5 segments | | |
| Part 3: Namespaced Commands | 5 checks | | |
| Part 4: Registry System | 13 scenarios | | |
| Part 5: Package Rename & Auto-Detection | 3 scenarios | | |

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
**Version:** v0.34.0  
**Platform:** ________________
