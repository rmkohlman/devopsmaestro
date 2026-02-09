# DevOpsMaestro Manual Test Plan

> **Version**: v0.6.0  
> **Last Updated**: February 2025

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

## Test Results Summary

| Section | Tests | Pass | Fail |
|---------|-------|------|------|
| Part 1: Setup & Build | 5 segments | | |
| Interactive: Environment | 6 checks | | |
| Interactive: Python | 5 checks | | |
| Interactive: Neovim | 2 checks | | |
| Part 2: Post-Attach | 5 segments | | |
| Part 3: Namespaced Commands | 5 checks | | |

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
**Version:** v0.6.0  
**Platform:** ________________
