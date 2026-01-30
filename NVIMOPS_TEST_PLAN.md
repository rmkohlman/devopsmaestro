# NvimOps Manual Test Plan

> **Version**: v0.4.0-dev  
> **Last Updated**: January 2026  
> **Binary**: `nvp`

---

## Overview

NvimOps is a standalone CLI for managing Neovim plugin configurations using a DevOps-style YAML approach. It can be used independently or alongside dvm.

**Key Features:**
- YAML-based plugin definitions (kubectl-style)
- Built-in plugin library (16+ curated plugins)
- Lua code generation for lazy.nvim
- File-based storage (~/.nvp/plugins/)
- No database required

---

## Quick Start

```bash
# Change to project directory first
cd ~/Developer/tools/devopsmaestro

# Run full automated test suite
./tests/manual/nvp/test-nvp.sh

# Or run with verbose output
NVP_VERBOSE=1 ./tests/manual/nvp/test-nvp.sh

# Or run with output preservation (for inspection)
NVP_KEEP_OUTPUT=1 ./tests/manual/nvp/test-nvp.sh

# Run nvim-config replica test (clones your config to temp dir)
./tests/manual/nvp/test-nvim-config-replica.sh

# Run specific test sections manually (see below)
```

---

## Test Structure

| Part | Type | Description |
|------|------|-------------|
| 1 | Automated | Build & basic commands |
| 2 | Automated | Library operations |
| 3 | Automated | Plugin CRUD operations |
| 4 | Automated | Lua generation |
| 5 | Manual | Interactive verification |
| 6 | Automated | Error handling & edge cases |
| 7 | Automated | Shell completions |
| 8 | Automated | Unit tests (Go) - Interface compliance |

---

## Part 1: Build and Basic Commands

### 1.1 Build the Binary

```bash
cd ~/Developer/tools/devopsmaestro
go build -o nvp ./cmd/nvp/
```

**Expected:** Binary compiles without errors.

| Test | Command | Expected | Status |
|------|---------|----------|--------|
| Build succeeds | `go build -o nvp ./cmd/nvp/` | Exit code 0 | |
| Binary exists | `ls -la nvp` | File exists, executable | |
| Binary runs | `./nvp --help` | Shows help text | |

### 1.2 Version Command

```bash
./nvp version
./nvp --version
./nvp version --short
```

**Expected:** Shows version info.

| Test | Command | Expected | Status |
|------|---------|----------|--------|
| Version full | `nvp version` | Version, build time, commit | |
| Version flag | `nvp --version` | Same as above | |
| Version short | `nvp version --short` | Just version number | |

### 1.3 Help Commands

```bash
./nvp --help
./nvp help
./nvp library --help
./nvp apply --help
./nvp generate --help
```

**Expected:** Each command shows relevant help.

| Test | Command | Expected | Status |
|------|---------|----------|--------|
| Root help | `nvp --help` | Lists all commands | |
| Library help | `nvp library --help` | Shows library subcommands | |
| Apply help | `nvp apply --help` | Shows -f flag, examples | |
| Generate help | `nvp generate --help` | Shows output dir options | |

---

## Part 2: Library Operations

The library contains pre-built plugin definitions that ship with nvp.

### 2.1 List Library Plugins

```bash
./nvp library list
./nvp library list -o yaml
./nvp library list -o json
./nvp library list --category lsp
./nvp library list --tag essential
```

**Expected:** Lists available plugins from built-in library.

| Test | Command | Expected | Status |
|------|---------|----------|--------|
| List all | `library list` | 16+ plugins listed | |
| List as YAML | `library list -o yaml` | Valid YAML output | |
| List as JSON | `library list -o json` | Valid JSON array | |
| Filter by category | `library list --category lsp` | Only LSP plugins | |
| Filter by tag | `library list --tag essential` | Tagged plugins | |

### 2.2 Show Library Plugin

```bash
./nvp library show telescope
./nvp library show telescope -o yaml
./nvp library show telescope -o json
./nvp library show nonexistent
```

**Expected:** Shows full plugin definition.

| Test | Command | Expected | Status |
|------|---------|----------|--------|
| Show telescope | `library show telescope` | Full plugin info | |
| Show as YAML | `library show telescope -o yaml` | Valid YAML with spec | |
| Show as JSON | `library show telescope -o json` | Valid JSON object | |
| Show nonexistent | `library show foo` | Error: plugin not found | |

### 2.3 Library Metadata

```bash
./nvp library categories
./nvp library tags
./nvp library info
```

**Expected:** Shows library metadata.

| Test | Command | Expected | Status |
|------|---------|----------|--------|
| Categories | `library categories` | List of categories | |
| Tags | `library tags` | List of unique tags | |
| Info summary | `library info` | Name, desc, category for all | |

### 2.4 Install from Library

```bash
./nvp library install telescope
./nvp library install telescope treesitter lspconfig
./nvp library install --all
./nvp library install nonexistent
```

**Expected:** Copies plugin from library to local store.

| Test | Command | Expected | Status |
|------|---------|----------|--------|
| Install single | `library install telescope` | Success message | |
| Install multiple | `library install telescope treesitter` | Both installed | |
| Install all | `library install --all` | All 16+ installed | |
| Install nonexistent | `library install foo` | Error message | |

---

## Part 3: Plugin CRUD Operations

### 3.1 Initialize Store

```bash
./nvp init
./nvp init --dir /tmp/nvp-test
ls -la ~/.nvp/
```

**Expected:** Creates store directory structure.

| Test | Command | Expected | Status |
|------|---------|----------|--------|
| Init default | `nvp init` | Creates ~/.nvp/ | |
| Init custom dir | `nvp init --dir /tmp/test` | Creates /tmp/test/ | |
| Config exists | `ls ~/.nvp/` | plugins/ directory exists | |

### 3.2 Apply Plugin from YAML

**Step 1:** Create test plugin file:

```bash
./tests/manual/nvp/create-test-plugin.sh
```

**Step 2:** Run apply commands (one at a time):

```bash
./nvp apply -f /tmp/test-plugin.yaml
```

```bash
./nvp apply -f /tmp/test-plugin.yaml
```

```bash
cat /tmp/test-plugin.yaml | ./nvp apply -f -
```

**Expected:** Plugin created/updated in store.

| Test | Command | Expected | Status |
|------|---------|----------|--------|
| Apply from file | `apply -f test.yaml` | Plugin created message | |
| Apply update | `apply -f test.yaml` (again) | Plugin configured message | |
| Apply from stdin | `cat test.yaml \| apply -f -` | Plugin created | |
| Apply invalid YAML | `echo "bad" \| apply -f -` | Error: parse failed | |
| Apply wrong kind | (kind: Other) | Error: invalid kind | |

### 3.3 List Plugins

```bash
./nvp list
./nvp list -o yaml
./nvp list -o json
./nvp list -o table
./nvp list --category testing
./nvp list --enabled
./nvp list --disabled
```

**Expected:** Lists plugins from local store.

| Test | Command | Expected | Status |
|------|---------|----------|--------|
| List all | `list` | Table with plugins | |
| List YAML | `list -o yaml` | Valid YAML | |
| List JSON | `list -o json` | Valid JSON array | |
| List by category | `list --category testing` | Filtered results | |
| List enabled | `list --enabled` | Only enabled plugins | |
| List empty | (no plugins) | "No plugins found" | |

### 3.4 Get Plugin

```bash
./nvp get my-plugin
./nvp get my-plugin -o yaml
./nvp get my-plugin -o json
./nvp get nonexistent
```

**Expected:** Shows plugin definition from store.

| Test | Command | Expected | Status |
|------|---------|----------|--------|
| Get existing | `get my-plugin` | Full plugin YAML | |
| Get as JSON | `get my-plugin -o json` | Valid JSON | |
| Get nonexistent | `get foo` | Error: not found | |

### 3.5 Enable/Disable Plugin

```bash
./nvp enable my-plugin
./nvp disable my-plugin
./nvp enable telescope treesitter  # Multiple
./nvp disable --all
./nvp enable --all
```

**Expected:** Toggles plugin enabled state.

| Test | Command | Expected | Status |
|------|---------|----------|--------|
| Enable plugin | `enable my-plugin` | Enabled message | |
| Disable plugin | `disable my-plugin` | Disabled message | |
| Enable multiple | `enable a b c` | All enabled | |
| Enable nonexistent | `enable foo` | Error: not found | |

### 3.6 Delete Plugin

```bash
./nvp delete my-plugin
./nvp delete my-plugin --force
./nvp delete nonexistent
```

**Expected:** Removes plugin from store.

| Test | Command | Expected | Status |
|------|---------|----------|--------|
| Delete with confirm | `delete my-plugin` | Prompts, then deletes | |
| Delete forced | `delete my-plugin --force` | Deletes immediately | |
| Delete nonexistent | `delete foo` | Error: not found | |

---

## Part 4: Lua Generation

### 4.1 Generate Lua Files

**Step 1:** Install some plugins from the library:

```bash
./nvp library install telescope treesitter lspconfig nvim-cmp
```

**Step 2:** Generate Lua files to a test directory:

```bash
./nvp generate --output /tmp/nvim-test-output
```

**Step 3:** (Optional) Test other generate options:

```bash
./nvp generate --dry-run
```

**Expected:** Generates lazy.nvim compatible Lua files.

| Test | Command | Expected | Status |
|------|---------|----------|--------|
| Generate custom dir | `generate --output /tmp/out` | Files in /tmp/out/ | |
| Generate dry-run | `generate --dry-run` | Shows what would generate | |
| Only enabled | (some disabled) | Only enabled plugins generate | |

### 4.2 Verify Generated Lua

**Prerequisite:** You must run 4.1 first to create `/tmp/nvim-test-output/`

**Step 1:** Check file structure:

```bash
ls -la /tmp/nvim-test-output/
```

**Step 2:** Verify Lua syntax (requires `luac` - skip if not installed):

```bash
for f in /tmp/nvim-test-output/*.lua; do luac -p "$f" 2>&1 && echo "$f: OK" || echo "$f: FAIL"; done
```

**Step 3:** Check content format:

```bash
head -20 /tmp/nvim-test-output/telescope.lua
```

**Expected:** Valid lazy.nvim plugin specs.

| Test | Check | Expected | Status |
|------|-------|----------|--------|
| Files exist | `ls *.lua` | One file per enabled plugin | |
| Lua syntax | `luac -p file.lua` | No syntax errors | |
| Has return | `head file.lua` | Starts with `-- Plugin: ...` then `return {` | |
| Has repo | `grep "nvim-telescope"` | Repo URL in output | |

### 4.3 Generate Single Plugin

```bash
./nvp generate-lua telescope
./nvp generate-lua telescope > /tmp/telescope.lua
./nvp generate-lua nonexistent
```

**Expected:** Outputs Lua for single plugin to stdout.

| Test | Command | Expected | Status |
|------|---------|----------|--------|
| Generate single | `generate-lua telescope` | Lua to stdout | |
| Redirect to file | `generate-lua telescope > file` | File created | |
| Nonexistent | `generate-lua foo` | Error: not found | |

---

## Part 5: Interactive Verification

### 5.1 Full Workflow Test

```bash
# Clean start
rm -rf ~/.nvp

# Initialize
./nvp init

# Install from library
./nvp library install telescope treesitter lspconfig mason nvim-cmp

# List what we have
./nvp list

# Generate Lua
./nvp generate --output /tmp/nvp-test/lua/plugins/managed

# View generated structure
tree /tmp/nvp-test/ || find /tmp/nvp-test/ -type f

# Verify a generated file
cat /tmp/nvp-test/lua/plugins/managed/telescope.lua
```

### 5.2 Test with Real Neovim (Optional)

If you want to test the generated config actually works with Neovim:

**Step 1:** Create test directory and generate plugins:

```bash
mkdir -p /tmp/nvp-test/lua/plugins/managed
```

```bash
./nvp library install --all
```

```bash
./nvp generate --output /tmp/nvp-test/lua/plugins/managed
```

**Step 2:** Create minimal init.lua:

```bash
./tests/manual/nvp/create-test-init.sh
```

**Step 3:** Test nvim startup:

```bash
NVIM_APPNAME=nvp-test nvim --headless "+Lazy sync" +qa 2>&1 | head -20
```

```bash
NVIM_APPNAME=nvp-test nvim
```

| Test | Check | Expected | Status |
|------|-------|----------|--------|
| Headless startup | No errors | Plugins load | |
| Interactive startup | `:Lazy` | Shows plugin list | |
| Telescope works | `<leader>ff` | Fuzzy finder opens | |

---

## Part 6: Error Handling & Edge Cases

### 6.1 Invalid Input

```bash
./nvp apply -f /nonexistent/file.yaml
./nvp apply -f /dev/null
./nvp apply  # No -f flag
./nvp get  # No name
./nvp library show  # No name
```

**Expected:** Clear error messages.

| Test | Command | Expected Error | Status |
|------|---------|----------------|--------|
| Missing file | `apply -f /bad/path` | File not found | |
| Empty file | `apply -f /dev/null` | Parse error | |
| Missing flag | `apply` | Must specify -f | |
| Missing arg | `get` | Name required | |

### 6.2 Invalid YAML

**Step 1:** Create invalid YAML test files:

```bash
./tests/manual/nvp/create-invalid-yaml-tests.sh
```

**Step 2:** Test each one (all should fail with validation errors):

```bash
./nvp apply -f /tmp/invalid-apiversion.yaml
```

```bash
./nvp apply -f /tmp/invalid-kind.yaml
```

```bash
./nvp apply -f /tmp/invalid-missing-repo.yaml
```

**Expected:** Validation errors.

| Test | Input | Expected Error | Status |
|------|-------|----------------|--------|
| Wrong apiVersion | `apiVersion: wrong/v1` | Warning or error | |
| Wrong kind | `kind: Other` | Invalid kind | |
| Missing repo | `spec: {}` | Repo is required | |
| Missing name | `metadata: {}` | Name is required | |

### 6.3 Store Edge Cases

```bash
# Uninitialized store
rm -rf ~/.nvp
./nvp list  # Should auto-init or error gracefully

# Corrupted plugin file
mkdir -p ~/.nvp/plugins
echo "not yaml" > ~/.nvp/plugins/bad.yaml
./nvp list  # Should skip or warn about bad file

# Duplicate plugin names
# (apply same name twice - should update, not duplicate)
```

| Test | Scenario | Expected | Status |
|------|----------|----------|--------|
| No store | `list` before init | Auto-init or clear error | |
| Bad file | Corrupted YAML in store | Warns, continues | |
| Duplicate | Apply same name twice | Updates, not duplicate | |

---

## Part 7: Shell Completions

### 7.1 Generate Completions

```bash
./nvp completion bash > /tmp/nvp.bash
./nvp completion zsh > /tmp/_nvp
./nvp completion fish > /tmp/nvp.fish
```

**Expected:** Valid completion scripts.

| Test | Command | Expected | Status |
|------|---------|----------|--------|
| Bash completion | `completion bash` | Valid bash script | |
| Zsh completion | `completion zsh` | Valid zsh script | |
| Fish completion | `completion fish` | Valid fish script | |

### 7.2 Test Completions (Manual)

```bash
# Bash
source /tmp/nvp.bash
./nvp <TAB><TAB>

# Zsh
source /tmp/_nvp
./nvp <TAB>
```

---

## Test Summary Checklist

### Part 1: Build & Basic Commands
- [ ] Binary builds successfully
- [ ] `--help` works for all commands
- [ ] `version` command works

### Part 2: Library Operations
- [ ] `library list` shows 16+ plugins
- [ ] `library show <name>` works
- [ ] `library categories` and `tags` work
- [ ] `library install` copies to store
- [ ] Output formats (table, yaml, json) work

### Part 3: Plugin CRUD
- [ ] `init` creates store directory
- [ ] `apply -f` creates/updates plugins
- [ ] `apply -f -` reads from stdin
- [ ] `list` shows stored plugins
- [ ] `get` retrieves single plugin
- [ ] `enable/disable` toggle plugin state
- [ ] `delete` removes plugins

### Part 4: Lua Generation
- [ ] `generate` creates Lua files
- [ ] Generated Lua has valid syntax
- [ ] Only enabled plugins are generated
- [ ] `generate-lua <name>` outputs single plugin

### Part 5: Integration
- [ ] Full workflow works end-to-end
- [ ] Generated config works with Neovim (optional)

### Part 6: Error Handling
- [ ] Invalid files show clear errors
- [ ] Invalid YAML shows validation errors
- [ ] Missing arguments show usage

### Part 7: Completions
- [ ] All shell completions generate

### Part 8: Unit Tests (Go)
- [ ] `go test ./pkg/nvimops/...` passes
- [ ] PluginStore interface compliance (MemoryStore, FileStore, ReadOnlyStore)
- [ ] LuaGenerator interface compliance (Generator, MockGenerator)
- [ ] Store swappability tests pass
- [ ] Generator swappability tests pass
- [ ] Library accessible through PluginStore interface
- [ ] Error type helpers work correctly

---

## Cleanup

```bash
# Remove test artifacts
rm -rf ~/.nvp
rm -rf /tmp/nvp-test
rm -rf /tmp/nvp-test-*
rm -f /tmp/test-plugin.yaml
rm -f /tmp/nvp.bash /tmp/_nvp /tmp/nvp.fish

# Remove binary (if testing locally)
rm -f ./nvp
```

---

## Part 8: Unit Tests (Go) - Interface Compliance

The `pkg/nvimops` library is designed with decoupled components that can be swapped or extended.

### 8.1 Run All Unit Tests

```bash
cd ~/Developer/tools/devopsmaestro
go test ./pkg/nvimops/... -v
```

**Expected:** All tests pass.

### 8.2 Interface Compliance Tests

The tests verify that all implementations satisfy their interfaces:

| Interface | Implementations | Test File |
|-----------|-----------------|-----------|
| `PluginStore` | MemoryStore, FileStore, ReadOnlyStore | `store/interface_test.go` |
| `LuaGenerator` | Generator, MockGenerator | `plugin/interface_test.go` |
| `ReadOnlySource` | Library | `store/interface_test.go` |

```bash
# Run interface compliance tests specifically
go test ./pkg/nvimops/store -run "TestInterfaceCompliance" -v
go test ./pkg/nvimops/plugin -run "TestGeneratorInterface" -v
```

### 8.3 Store Swappability Tests

Verifies that code written against `PluginStore` interface works with any implementation:

```bash
go test ./pkg/nvimops/store -run "TestStoreSwappability" -v
```

**Expected:** Same test logic passes against:
- `MemoryStore` (in-memory)
- `FileStore` (disk-based)
- `ReadOnlyStore` wrapping `Library` (embedded)

### 8.4 Generator Swappability Tests

Verifies that code written against `LuaGenerator` interface works with any implementation:

```bash
go test ./pkg/nvimops/plugin -run "TestGeneratorSwappability" -v
```

**Expected:** Same test logic passes against:
- `Generator` (default implementation)
- `MockGenerator` (for testing)
- Custom implementations

### 8.5 Library Through Store Interface

Tests that the embedded Library can be accessed via the `PluginStore` interface:

```bash
go test ./pkg/nvimops/store -run "TestReadOnlyStore" -v
```

**Expected:**
- Read operations work (Get, List, Exists)
- Write operations return `ErrReadOnly`
- Library plugins accessible through standard interface

### 8.6 Error Type Handling

```bash
go test ./pkg/nvimops/store -run "TestErrorTypes" -v
```

**Expected:** Error type checking helpers work:
- `IsNotFound(err)` for `ErrNotFound`
- `IsAlreadyExists(err)` for `ErrAlreadyExists`
- `IsReadOnly(err)` for `ErrReadOnly`

### 8.7 Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              Interfaces                                  │
├────────────────────────────────┬────────────────────────────────────────┤
│         PluginStore            │           LuaGenerator                  │
│  - Create/Update/Delete        │   - GenerateLua(plugin)                │
│  - Get/List/Exists             │   - GenerateLuaFile(plugin)            │
│  - ListByCategory/Tag          │                                         │
└───────────────┬────────────────┴───────────────────┬────────────────────┘
                │                                     │
┌───────────────┴─────────────────┐ ┌────────────────┴────────────────────┐
│        Implementations          │ │          Implementations            │
├─────────────────────────────────┤ ├─────────────────────────────────────┤
│  MemoryStore   (testing)        │ │  Generator    (default lazy.nvim)   │
│  FileStore     (production)     │ │  MockGenerator (testing)            │
│  ReadOnlyStore (library wrap)   │ │  (custom...)  (extensible)          │
│  (DBStore...)  (future/dvm)     │ │                                      │
└─────────────────────────────────┘ └─────────────────────────────────────┘
                │
                │ wraps
                ▼
┌─────────────────────────────────┐
│        ReadOnlySource           │
├─────────────────────────────────┤
│  Library  (embedded plugins)    │
│  (custom sources...)            │
└─────────────────────────────────┘
```

### 8.8 Extensibility Examples

The decoupled architecture allows:

**Custom Store Implementation:**
```go
// Example: Database-backed store for dvm integration
type DBPluginStore struct {
    db *sql.DB
}

// Implements PluginStore interface
func (s *DBPluginStore) Get(name string) (*plugin.Plugin, error) { ... }
func (s *DBPluginStore) List() ([]*plugin.Plugin, error) { ... }
// etc.

// Use with Manager
mgr, _ := nvimops.NewWithOptions(nvimops.Options{
    Store: &DBPluginStore{db: db},
})
```

**Custom Generator Implementation:**
```go
// Example: Generator for different plugin manager (packer, vim-plug)
type PackerGenerator struct {}

// Implements LuaGenerator interface
func (g *PackerGenerator) GenerateLua(p *plugin.Plugin) (string, error) {
    return fmt.Sprintf("use '%s'", p.Repo), nil
}

// Use with Manager
mgr, _ := nvimops.NewWithOptions(nvimops.Options{
    Generator: &PackerGenerator{},
})
```

---

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `NVP_CONFIG_DIR` | Config/store directory | `~/.nvp` |
| `NVP_OUTPUT_DIR` | Default generate output | `~/.config/nvim/lua/plugins/nvp` |
| `NO_COLOR` | Disable colored output | (unset) |

---

## Troubleshooting

### "permission denied" on generate

The output directory may not be writable:
```bash
mkdir -p ~/.config/nvim/lua/plugins/nvp
chmod 755 ~/.config/nvim/lua/plugins/nvp
```

### "plugin not found" after install

Check the store directory:
```bash
ls -la ~/.nvp/plugins/
```

### Lua syntax errors in generated files

Report a bug! Include:
1. Plugin YAML definition
2. Generated Lua file
3. Error message from `luac -p`

---

**Tested by:** ________________  
**Date:** ________________  
**Version:** v0.4.0-dev  
**Platform:** ________________
