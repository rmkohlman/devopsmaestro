# ADR-008: Shared Library Architecture for Neovim Management

**Status:** Proposed  
**Date:** 2026-01-23  
**Author:** Robert Kohlman  
**Related:** ADR-007 (Neovim Local Management)

## Context

We want to implement Neovim configuration management in two places:
1. **DevOpsMaestro (`dvm nvim`)** - Integrated workspace-aware commands
2. **nvim-maestro** (future) - Standalone CLI for users who only want Neovim management

**Challenge:** How do we avoid code duplication and maintain a single source of truth?

## Decision

Create a **shared library architecture** with three repositories:

```
┌─────────────────────────────────────────────────────────────────┐
│                    nvim-maestro-lib                             │
│                   (Shared Core Library)                          │
│                                                                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   manager/   │  │   config/    │  │   sync/      │          │
│  │              │  │              │  │              │          │
│  │ - Init()     │  │ - Parser     │  │ - Push()     │          │
│  │ - Sync()     │  │ - Validator  │  │ - Pull()     │          │
│  │ - Status()   │  │ - Templates  │  │ - Diff()     │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└─────────────────────────────────────────────────────────────────┘
                   ▲                            ▲
                   │                            │
                   │ imports as Go module       │
                   │                            │
    ┌──────────────┴─────────────┐   ┌──────────┴──────────────┐
    │    devopsmaestro           │   │   nvim-maestro          │
    │  (Main CLI - Phase 1)      │   │ (Standalone - Phase 3)  │
    │                            │   │                         │
    │  cmd/nvim.go               │   │  main.go                │
    │  ├─ nvim init              │   │  ├─ init                │
    │  ├─ nvim sync <workspace>  │   │  ├─ sync                │
    │  └─ nvim push <workspace>  │   │  └─ status              │
    └────────────────────────────┘   └─────────────────────────┘
```

## Architecture Details

### 1. **nvim-maestro-lib** (Core Library)

**Repository:** `github.com/rmkohlman/nvim-maestro-lib`

**Purpose:** Pure Go library with zero CLI dependencies

**Structure:**
```
nvim-maestro-lib/
├── manager/
│   ├── manager.go          # Main Manager interface
│   ├── init.go             # Initialize local Neovim config
│   ├── sync.go             # Sync logic (bidirectional)
│   └── status.go           # Status checking
├── config/
│   ├── parser.go           # Parse Neovim Lua configs
│   ├── validator.go        # Validate config structure
│   └── templates.go        # Config templates
├── sync/
│   ├── engine.go           # Sync engine
│   ├── diff.go             # Config diff logic
│   └── resolver.go         # Conflict resolution
├── storage/
│   ├── local.go            # Local filesystem operations
│   └── remote.go           # Remote workspace operations
└── go.mod
```

**Key Interfaces:**
```go
// manager/manager.go
package manager

type Manager interface {
    // Initialize local Neovim configuration
    Init(opts InitOptions) error
    
    // Sync local config with remote workspace
    Sync(workspace string, direction SyncDirection) error
    
    // Get status of local config
    Status() (*Status, error)
    
    // List available workspaces
    ListWorkspaces() ([]Workspace, error)
}

type InitOptions struct {
    ConfigPath  string          // Default: ~/.config/nvim
    Template    string          // kickstart, lazyvim, astronvim, custom
    Overwrite   bool           // Overwrite existing config
}

type SyncDirection int
const (
    SyncPull SyncDirection = iota  // workspace → local
    SyncPush                        // local → workspace
    SyncBidirectional              // merge both ways
)

type Status struct {
    ConfigPath    string
    LastSync      time.Time
    SyncedWith    string        // workspace ID
    LocalChanges  bool
    RemoteChanges bool
}
```

### 2. **devopsmaestro** (Integrated CLI)

**Phase 1 Implementation** (v0.3.0)

**New Files:**
```
devopsmaestro/
├── cmd/
│   └── nvim.go              # Cobra commands for 'dvm nvim'
├── go.mod                   # Add: github.com/rmkohlman/nvim-maestro-lib
```

**Command Structure:**
```go
// cmd/nvim.go
package cmd

import (
    nvim "github.com/rmkohlman/nvim-maestro-lib/manager"
)

var nvimCmd = &cobra.Command{
    Use:   "nvim",
    Short: "Manage local Neovim configuration",
}

var nvimInitCmd = &cobra.Command{
    Use:   "init [template]",
    Short: "Initialize local Neovim config from template",
    Run: func(cmd *cobra.Command, args []string) {
        mgr := nvim.NewManager()
        opts := nvim.InitOptions{
            ConfigPath: "~/.config/nvim",
            Template:   args[0], // kickstart, lazyvim, etc.
        }
        if err := mgr.Init(opts); err != nil {
            fmt.Printf("Failed to init: %v\n", err)
            os.Exit(1)
        }
    },
}

var nvimSyncCmd = &cobra.Command{
    Use:   "sync <workspace>",
    Short: "Sync local config with workspace container",
    Run: func(cmd *cobra.Command, args []string) {
        mgr := nvim.NewManager()
        // DevOpsMaestro-specific: resolve workspace from DB
        workspace := getWorkspaceByName(args[0])
        
        if err := mgr.Sync(workspace.ID, nvim.SyncPull); err != nil {
            fmt.Printf("Failed to sync: %v\n", err)
            os.Exit(1)
        }
    },
}
```

### 3. **nvim-maestro** (Standalone CLI)

**Phase 3 Implementation** (v0.5.0+)

**Structure:**
```
nvim-maestro/
├── main.go                  # Simple CLI wrapper
├── go.mod                   # imports: nvim-maestro-lib
└── README.md
```

**Example:**
```go
// main.go
package main

import (
    "fmt"
    "os"
    nvim "github.com/rmkohlman/nvim-maestro-lib/manager"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: nvim-maestro <command>")
        os.Exit(1)
    }
    
    mgr := nvim.NewManager()
    
    switch os.Args[1] {
    case "init":
        // Same library, simpler interface
        mgr.Init(nvim.InitOptions{Template: os.Args[2]})
    case "status":
        status, _ := mgr.Status()
        fmt.Printf("Config: %s\n", status.ConfigPath)
        fmt.Printf("Last sync: %s\n", status.LastSync)
    default:
        fmt.Println("Unknown command")
        os.Exit(1)
    }
}
```

## Implementation Phases

### **Phase 1: Develop in DevOpsMaestro** (v0.3.0 - NOW)
- Create `devopsmaestro/nvim/` package locally
- Implement core features directly in DevOpsMaestro
- Test with real workspaces
- Iterate quickly without external dependencies

**Pros:**
- Fast iteration
- Test with real workspace integration
- No cross-repo coordination needed

**Structure:**
```
devopsmaestro/
├── nvim/
│   ├── manager.go
│   ├── init.go
│   └── sync.go
└── cmd/nvim.go
```

### **Phase 2: Extract to Library** (v0.4.0)
- Create `nvim-maestro-lib` repository
- Move `devopsmaestro/nvim/` → `nvim-maestro-lib/manager/`
- Update `devopsmaestro` to import library
- Publish v1.0.0 of library

**Migration:**
```bash
# Create new repo
gh repo create rmkohlman/nvim-maestro-lib --public

# Copy code
cd nvim-maestro-lib
mkdir -p manager config sync storage
cp ~/devopsmaestro/nvim/*.go manager/

# Publish
git tag v1.0.0
git push origin v1.0.0

# Update devopsmaestro
cd ~/devopsmaestro
go get github.com/rmkohlman/nvim-maestro-lib@v1.0.0
# Update imports in cmd/nvim.go
```

### **Phase 3: Standalone CLI** (v0.5.0+)
- Create `nvim-maestro` repository
- Simple CLI wrapper around library
- Optional for users who don't need DevOpsMaestro
- Same features, simpler interface

## Benefits

### **1. Single Source of Truth**
- All logic lives in `nvim-maestro-lib`
- Bug fixes benefit both tools
- Features developed once

### **2. Clean Separation of Concerns**
```
Library:    Pure business logic, no CLI deps
CLIs:       UI/UX layer, CLI frameworks (Cobra, etc.)
```

### **3. Independent Release Cycles**
```
nvim-maestro-lib:  v1.0.0, v1.1.0, v1.2.0 (frequent)
devopsmaestro:     v0.3.0 (uses lib v1.0.0)
                   v0.4.0 (uses lib v1.2.0)
nvim-maestro:      v1.0.0 (uses lib v1.2.0)
```

### **4. Easy Testing**
- Test library once comprehensively
- CLI tests focus on UX, not logic
- Mock library in CLI tests

### **5. Community Reuse**
- Other developers can use `nvim-maestro-lib`
- Build custom integrations
- Contribute back to core library

## Example: Feature Development Flow

**Scenario:** Add support for `.nvim-maestro.yaml` config files

**Step 1:** Library Implementation
```go
// nvim-maestro-lib/config/parser.go
func ParseMaestroConfig(path string) (*Config, error) {
    // Implementation
}
```

**Step 2:** DevOpsMaestro Usage
```go
// devopsmaestro/cmd/nvim.go
import "github.com/rmkohlman/nvim-maestro-lib/config"

func syncCommand() {
    cfg, _ := config.ParseMaestroConfig(".nvim-maestro.yaml")
    // Use config
}
```

**Step 3:** nvim-maestro Usage
```go
// nvim-maestro/main.go
import "github.com/rmkohlman/nvim-maestro-lib/config"

func main() {
    cfg, _ := config.ParseMaestroConfig(".nvim-maestro.yaml")
    // Same feature, zero duplication
}
```

## Alternatives Considered

### **Alternative 1: Monorepo**
Keep everything in `devopsmaestro` repo

**Pros:** Simple, single repo
**Cons:** 
- Large dependency for users who only want Neovim management
- Tight coupling
- Can't version library independently

**Verdict:** ❌ Rejected - Too coupled

### **Alternative 2: Code Duplication**
Maintain separate implementations

**Pros:** Total independence
**Cons:**
- Duplicate work
- Bugs fixed twice
- Features diverge over time
- Maintenance nightmare

**Verdict:** ❌ Rejected - Unsustainable

### **Alternative 3: Git Submodules**
Use git submodules for shared code

**Pros:** Share code without separate repo
**Cons:**
- Submodules are complex
- Poor Go module support
- Versioning issues

**Verdict:** ❌ Rejected - Go modules better

## Decision Rationale

**Why shared library architecture wins:**

1. **Best of Both Worlds:** Fast iteration in Phase 1, clean separation in Phase 2
2. **Standard Go Practice:** Using Go modules for shared code is idiomatic
3. **Scalable:** Easy to add more consumers of the library
4. **Testable:** Library can be tested in isolation
5. **Versioned:** SemVer for library, independent of CLI versions

## Implementation Checklist

**Phase 1 (v0.3.0):**
- [ ] Create `devopsmaestro/nvim/` package
- [ ] Implement `Manager` interface
- [ ] Implement `Init()` functionality
- [ ] Implement `Sync()` functionality
- [ ] Add `dvm nvim init` command
- [ ] Add `dvm nvim sync` command
- [ ] Add `dvm nvim push` command
- [ ] Write comprehensive tests
- [ ] Document commands

**Phase 2 (v0.4.0):**
- [ ] Create `nvim-maestro-lib` repository
- [ ] Extract code from `devopsmaestro/nvim/`
- [ ] Add library tests
- [ ] Add library documentation
- [ ] Publish v1.0.0 of library
- [ ] Update DevOpsMaestro to import library
- [ ] Verify all functionality still works

**Phase 3 (v0.5.0+):**
- [ ] Create `nvim-maestro` repository
- [ ] Build simple CLI wrapper
- [ ] Add installation instructions
- [ ] Create Homebrew formula
- [ ] Announce standalone tool

## Success Criteria

1. ✅ Zero code duplication between tools
2. ✅ Library has >90% test coverage
3. ✅ Library can be used independently
4. ✅ Both CLIs work identically for core features
5. ✅ Clear documentation for library consumers

## References

- ADR-007: Local Neovim Management
- Go Modules Documentation: https://go.dev/doc/modules/
- Clean Architecture patterns
- DRY (Don't Repeat Yourself) principle

---

**Next Steps:**
Start Phase 1 implementation in `devopsmaestro/nvim/` package for v0.3.0.
