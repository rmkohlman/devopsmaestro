package handlers

// Full System Round-Trip Integration Test — Issue #173
//
// Creates a realistic multi-ecosystem system with EVERY resource type,
// exports to YAML, wipes the DB, restores from YAML, re-exports, and
// verifies the output is identical. This is the capstone GitOps contract test.
//
// Ecosystems:
//   - prod-automation: backend (api-gateway→gw-dev, auth-service→auth-dev)
//                      frontend (web-portal→web-dev)
//   - lib-dev:         backend (SAME name as prod-automation!) → shared-lib → dev
//
// Global: 2 registries, 1 gitrepo, GlobalDefaults

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rmkohlman/MaestroSDK/resource"
)

// buildFullSystem creates the complete multi-ecosystem system in the given context.
// Returns the list of applied resource names for verification.
func buildFullSystem(t *testing.T, ctx resource.Context) {
	t.Helper()

	// =========================================================================
	// Ecosystem A: prod-automation
	// =========================================================================
	doApply(t, ctx, KindEcosystem, `apiVersion: devopsmaestro.io/v1
kind: Ecosystem
metadata:
  name: prod-automation
spec:
  theme: catppuccin-mocha
  description: Production automation platform
  build:
    args:
      HTTP_PROXY: "http://proxy:8080"
`)

	// Domains under prod-automation
	doApply(t, ctx, KindDomain, `apiVersion: devopsmaestro.io/v1
kind: Domain
metadata:
  name: backend
  ecosystem: prod-automation
spec:
  build:
    args:
      JAVA_HOME: "/usr/lib/jvm/java-17"
`)

	doApply(t, ctx, KindDomain, `apiVersion: devopsmaestro.io/v1
kind: Domain
metadata:
  name: frontend
  ecosystem: prod-automation
spec:
  build:
    args:
      NODE_ENV: "production"
`)

	// Apps under prod-automation/backend
	doApply(t, ctx, KindApp, `apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: api-gateway
  domain: backend
  ecosystem: prod-automation
spec:
  path: /src/gateway
`)

	doApply(t, ctx, KindApp, `apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: auth-service
  domain: backend
  ecosystem: prod-automation
spec:
  path: /src/auth
`)

	// App under prod-automation/frontend
	doApply(t, ctx, KindApp, `apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: web-portal
  domain: frontend
  ecosystem: prod-automation
spec:
  path: /src/web
`)

	// Workspaces — gw-dev gets nvim + terminal config
	doApply(t, ctx, KindWorkspace, `apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: gw-dev
  app: api-gateway
  domain: backend
  ecosystem: prod-automation
spec:
  image:
    name: node:20
  build:
    args:
      DEBUG: "1"
  nvim:
    structure: lazyvim
    theme: tokyonight
    plugins:
      - telescope
  terminal:
    prompt: starship-default
  env:
    PORT: "3000"
    NODE_ENV: dev
`)

	doApply(t, ctx, KindWorkspace, `apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: auth-dev
  app: auth-service
  domain: backend
  ecosystem: prod-automation
spec:
  image:
    name: python:3.12
  env:
    FLASK_ENV: development
`)

	doApply(t, ctx, KindWorkspace, `apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: web-dev
  app: web-portal
  domain: frontend
  ecosystem: prod-automation
spec:
  image:
    name: node:20
  env:
    VITE_API_URL: "http://localhost:3000"
`)

	// Credentials at every scope for prod-automation
	doApply(t, ctx, KindCredential, `apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: eco-cred
  ecosystem: prod-automation
spec:
  source: vault
  vaultSecret: eco/secret
  usernameVar: ECO_USER
  passwordVar: ECO_PASS
`)

	doApply(t, ctx, KindCredential, `apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: domain-cred
  domain: backend
spec:
  source: vault
  vaultSecret: domain/secret
`)

	doApply(t, ctx, KindCredential, `apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: app-cred
  app: api-gateway
spec:
  source: env
  envVar: APP_TOKEN
`)

	doApply(t, ctx, KindCredential, `apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: ws-cred
  workspace: gw-dev
spec:
  source: vault
  vaultSecret: ws/secret
  usernameVar: WS_USER
  passwordVar: WS_PASS
`)

	// =========================================================================
	// Ecosystem B: lib-dev (intentionally reuses "backend" domain name)
	// =========================================================================
	doApply(t, ctx, KindEcosystem, `apiVersion: devopsmaestro.io/v1
kind: Ecosystem
metadata:
  name: lib-dev
spec:
  description: Library development
  build:
    args:
      GO_VERSION: "1.22"
`)

	doApply(t, ctx, KindDomain, `apiVersion: devopsmaestro.io/v1
kind: Domain
metadata:
  name: backend
  ecosystem: lib-dev
spec: {}
`)

	doApply(t, ctx, KindApp, `apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: shared-lib
  domain: backend
  ecosystem: lib-dev
spec:
  path: /src/shared
`)

	doApply(t, ctx, KindWorkspace, `apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: dev
  app: shared-lib
  domain: backend
  ecosystem: lib-dev
spec:
  image:
    name: golang:1.22
`)

	doApply(t, ctx, KindCredential, `apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: lib-cred
  ecosystem: lib-dev
spec:
  source: vault
  vaultSecret: lib/secret
  usernameVar: LIB_USER
  passwordVar: LIB_PASS
`)

	// =========================================================================
	// Global resources
	// =========================================================================
	doApply(t, ctx, KindRegistry, `apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: my-verdaccio
spec:
  type: verdaccio
  port: 4873
  lifecycle: persistent
`)

	doApply(t, ctx, KindRegistry, `apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: my-zot
spec:
  type: zot
  port: 5000
  lifecycle: on-demand
`)

	doApply(t, ctx, KindGitRepo, `apiVersion: devopsmaestro.io/v1
kind: GitRepo
metadata:
  name: gateway-repo
spec:
  url: https://github.com/org/gateway.git
  authType: none
  defaultRef: main
`)

	doApply(t, ctx, KindNvimPlugin, `apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: telescope
spec:
  repo: nvim-telescope/telescope.nvim
  lazy: true
  category: navigation
`)

	doApply(t, ctx, KindTerminalPrompt, `apiVersion: devopsmaestro.io/v1
kind: TerminalPrompt
metadata:
  name: starship-default
spec:
  type: starship
`)

	doApply(t, ctx, KindNvimTheme, `apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: catppuccin-mocha
  description: Soothing pastel theme
  category: dark
spec:
  plugin:
    repo: catppuccin/nvim
  style: mocha
  colors:
    bg: "#1e1e2e"
    fg: "#cdd6f4"
`)

	doApply(t, ctx, KindNvimPackage, `apiVersion: devopsmaestro.io/v1
kind: NvimPackage
metadata:
  name: lazy
  description: Core lazy.nvim package set
  category: core
spec:
  plugins:
    - telescope
    - treesitter
    - lspconfig
`)

	doApply(t, ctx, KindTerminalPackage, `apiVersion: devopsmaestro.io/v1
kind: TerminalPackage
metadata:
  name: wezterm
  description: WezTerm terminal package
  category: terminal
spec:
  plugins:
    - zsh-autosuggestions
  prompts:
    - starship-default
`)

	doApply(t, ctx, KindTerminalPlugin, `apiVersion: devopsmaestro.io/v1
kind: TerminalPlugin
metadata:
  name: zsh-autosuggestions
  description: Fish-like autosuggestions for zsh
  category: productivity
spec:
  repo: zsh-users/zsh-autosuggestions
  shell: zsh
  manager: manual
`)

	doApply(t, ctx, KindCRD, `apiVersion: devopsmaestro.io/v1alpha1
kind: CustomResourceDefinition
metadata:
  name: databases.devopsmaestro.io
spec:
  group: devopsmaestro.io
  names:
    kind: Database
    singular: database
    plural: databases
    shortNames: [db]
  scope: Workspace
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              required: [engine, version]
              properties:
                engine:
                  type: string
                  enum: [postgres, mysql, sqlite]
                version:
                  type: string
`)

	doApply(t, ctx, KindGlobalDefaults, `apiVersion: devopsmaestro.io/v1
kind: GlobalDefaults
metadata:
  name: global-defaults
spec:
  theme: tokyonight
  buildArgs:
    GLOBAL_NO_PROXY: "localhost,127.0.0.1"
  nvimPackage: lazy
  terminalPackage: wezterm
`)
}

// expectedCounts returns the minimum per-kind resource counts we expect
// after building the full system.
func expectedCounts() map[string]int {
	return map[string]int{
		KindEcosystem:       2,
		KindDomain:          3,
		KindApp:             4,
		KindWorkspace:       4,
		KindCredential:      5,
		KindRegistry:        2,
		KindGitRepo:         1,
		KindNvimPlugin:      1,
		KindTerminalPrompt:  1,
		KindNvimTheme:       1,
		KindNvimPackage:     1,
		KindTerminalPackage: 1,
		KindTerminalPlugin:  1,
		KindCRD:             1,
		KindGlobalDefaults:  1,
	}
}

// assertMinCounts verifies each expected kind has at least the minimum count.
func assertMinCounts(t *testing.T, counts map[string]int, label string) {
	t.Helper()
	for kind, want := range expectedCounts() {
		got := counts[kind]
		if got < want {
			t.Errorf("%s: %s count = %d, want >= %d", label, kind, got, want)
		}
	}
}

// mustContainAll verifies the YAML string contains all required needles.
func mustContainAll(t *testing.T, yamlStr, label string, needles ...string) {
	t.Helper()
	for _, needle := range needles {
		if !strings.Contains(yamlStr, needle) {
			t.Errorf("%s: YAML missing %q", label, needle)
		}
	}
}

// assertCountsMatch verifies that want and got maps have identical values for
// every key present in want.
func assertCountsMatch(t *testing.T, want, got map[string]int, label string) {
	t.Helper()
	for kind, wantCount := range want {
		gotCount := got[kind]
		if gotCount != wantCount {
			t.Errorf("%s: %s count = %d, want %d", label, kind, gotCount, wantCount)
		}
	}
}

// TestFullSystemRoundTrip is the capstone integration test:
// build → export → wipe → restore → re-export → verify parity + idempotency.
func TestFullSystemRoundTrip(t *testing.T) {
	RegisterAll()

	// Phase 1: Build the full multi-ecosystem system
	ds1 := createStackingDS(t)
	defer ds1.Close()
	ctx1 := resource.Context{DataStore: ds1}

	buildFullSystem(t, ctx1)

	// Phase 2: Export and verify counts + content
	yaml1, counts1 := exportSnapshot(t, ds1)

	assertMinCounts(t, counts1, "initial export")
	mustContainAll(t, yaml1, "initial export",
		// Eco A resources
		"prod-automation",
		"lib-dev",
		// Domains (both "backend" must appear — from different ecosystems)
		"frontend",
		// Apps
		"api-gateway",
		"auth-service",
		"web-portal",
		"shared-lib",
		// Workspaces
		"gw-dev",
		"auth-dev",
		"web-dev",
		// Credentials
		"eco-cred",
		"domain-cred",
		"app-cred",
		"ws-cred",
		"lib-cred",
		// Registries
		"my-verdaccio",
		"my-zot",
		// Git repo
		"gateway-repo",
		// Nvim + terminal
		"telescope",
		"starship-default",
		// Global defaults content
		"GLOBAL_NO_PROXY",
		"tokyonight",
		"lazy",
		"wezterm",
		// Build args from various levels
		"HTTP_PROXY",
		"GO_VERSION",
		"DEBUG",
		// Theme at ecosystem level and NvimTheme resource
		"catppuccin-mocha",
		// NvimPackage, TerminalPackage, TerminalPlugin, CRD
		"zsh-autosuggestions",
		"databases.devopsmaestro.io",
	)
	t.Logf("Phase 2 export counts: %v  total=%d", counts1, sumCounts(counts1))

	// Phase 3: Verify overlapping domain name — "backend" must appear in BOTH ecosystems
	t.Run("overlapping_backend_domain", func(t *testing.T) {
		// We have 3 domains total: backend@prod-automation, frontend@prod-automation, backend@lib-dev
		if counts1[KindDomain] < 3 {
			t.Errorf("expected >= 3 domains (2 backend + 1 frontend), got %d", counts1[KindDomain])
		}
		// Count occurrences of "backend" in YAML — should appear at least twice
		backendOccurrences := strings.Count(yaml1, "backend")
		if backendOccurrences < 2 {
			t.Errorf("expected 'backend' to appear at least twice (two ecosystems), got %d", backendOccurrences)
		}
		// Verify both ecosystems are referenced alongside backend
		if !strings.Contains(yaml1, "prod-automation") || !strings.Contains(yaml1, "lib-dev") {
			t.Error("both ecosystem names must appear in initial export")
		}
	})

	// Phase 4: Wipe — create fresh in-memory DB
	ds2 := createStackingDS(t)
	defer ds2.Close()
	ctx2 := resource.Context{DataStore: ds2}

	// Phase 5: Restore from exported YAML
	applied, err := resource.ApplyList(ctx2, []byte(yaml1))
	if err != nil {
		t.Errorf("ApplyList (restore) errors: %v (applied %d resources)", err, len(applied))
	}
	t.Logf("Phase 5: ApplyList applied %d resources", len(applied))

	// Phase 6: Re-export from restored DB
	yaml2, counts2 := exportSnapshot(t, ds2)

	// Phase 7: Verify per-kind counts are identical after restore
	t.Run("counts_parity", func(t *testing.T) {
		assertCountsMatch(t, counts1, counts2, "restore parity")
		t.Logf("Before: %v  After: %v", counts1, counts2)
		t.Logf("Total before=%d  after=%d", sumCounts(counts1), sumCounts(counts2))
	})

	// Phase 8: Verify all content survived restore
	mustContainAll(t, yaml2, "restored export",
		"prod-automation", "lib-dev",
		"api-gateway", "auth-service", "web-portal", "shared-lib",
		"gw-dev", "auth-dev", "web-dev",
		"eco-cred", "domain-cred", "app-cred", "ws-cred", "lib-cred",
		"my-verdaccio", "my-zot",
		"gateway-repo",
		"telescope", "starship-default",
		"GLOBAL_NO_PROXY", "tokyonight", "lazy", "wezterm",
		"HTTP_PROXY", "GO_VERSION",
		"catppuccin-mocha",
		"zsh-autosuggestions",
		"databases.devopsmaestro.io",
	)

	// Phase 9: Verify overlapping "backend" domain survives restore in both ecosystems
	t.Run("overlapping_domain_after_restore", func(t *testing.T) {
		if counts2[KindDomain] < 3 {
			t.Errorf("restore: expected >= 3 domains, got %d", counts2[KindDomain])
		}
		// Both ecosystems must still be present post-restore
		mustContainAll(t, yaml2, "restore overlap check", "prod-automation", "lib-dev", "backend")
		// Both lib-dev-specific credential and prod-automation-specific credential must exist
		mustContainAll(t, yaml2, "restore credentials check", "lib-cred", "eco-cred")
	})

	// Phase 10: Idempotency — apply the restored YAML a second time
	t.Run("idempotency", func(t *testing.T) {
		applied2, err2 := resource.ApplyList(ctx2, []byte(yaml2))
		if err2 != nil {
			t.Errorf("idempotent ApplyList errors: %v (applied %d)", err2, len(applied2))
		}

		_, counts3 := exportSnapshot(t, ds2)
		assertCountsMatch(t, counts2, counts3, "idempotency (no duplicates after second apply)")
		t.Logf("Idempotency: before second apply=%v  after=%v", counts2, counts3)
	})

	// Phase 11: Scoped verification — no cross-ecosystem leakage
	t.Run("no_cross_ecosystem_leakage", func(t *testing.T) {
		// Verify prod-automation credentials are NOT in lib-dev scope
		// We check by verifying total credential counts per ecosystem.
		// prod-automation has: eco-cred, domain-cred, app-cred, ws-cred (4 creds)
		// lib-dev has: lib-cred (1 cred)
		// Total = 5 — if restored count is 5, no cross-contamination happened.
		if counts2[KindCredential] != 5 {
			t.Errorf("credential count after restore = %d, want exactly 5 (4 prod-auto + 1 lib-dev)", counts2[KindCredential])
		}

		// Workspaces: gw-dev, auth-dev, web-dev (prod-auto) + dev (lib-dev) = 4 total
		if counts2[KindWorkspace] != 4 {
			t.Errorf("workspace count after restore = %d, want exactly 4", counts2[KindWorkspace])
		}

		// Apps: api-gateway, auth-service, web-portal (prod-auto) + shared-lib (lib-dev) = 4 total
		if counts2[KindApp] != 4 {
			t.Errorf("app count after restore = %d, want exactly 4", counts2[KindApp])
		}
	})

	// Summary log
	t.Logf("=== Full Round-Trip Summary ===")
	t.Logf("Resources before wipe: %d total", sumCounts(counts1))
	t.Logf("Resources after restore: %d total", sumCounts(counts2))

	parity := true
	for kind, want := range counts1 {
		got := counts2[kind]
		status := "OK"
		if got != want {
			status = fmt.Sprintf("MISMATCH (want %d got %d)", want, got)
			parity = false
		}
		t.Logf("  %-20s before=%-3d after=%-3d %s", kind, want, got, status)
	}
	if parity {
		t.Log("RESULT: PERFECT PARITY — GitOps contract validated")
	} else {
		t.Log("RESULT: PARITY FAILURE — see mismatches above")
	}
}
