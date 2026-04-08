package handlers

// Progressive Stacking Integration Test — Issue #172
//
// Incrementally builds ecosystem → domain → app → workspace → registry+globals,
// exporting YAML at each step and verifying it is a strict superset of the
// previous. Final step: wipe DB, restore from YAML, re-export, verify parity.

import (
	"strings"
	"testing"

	"devopsmaestro/db"
	"github.com/rmkohlman/MaestroSDK/resource"

	"gopkg.in/yaml.v3"
)

// createStackingDS creates a fully-schemaed in-memory SQLite DataStore.
func createStackingDS(t *testing.T) *db.SQLDataStore {
	t.Helper()
	cfg := db.DriverConfig{Type: db.DriverMemory}
	driver, err := db.NewMemorySQLiteDriver(cfg)
	if err != nil {
		t.Fatalf("NewMemorySQLiteDriver: %v", err)
	}
	if err := driver.Connect(); err != nil {
		t.Fatalf("driver.Connect: %v", err)
	}
	for _, s := range stackingSchema() {
		if _, err := driver.Execute(s); err != nil {
			driver.Close()
			t.Fatalf("schema exec: %v", err)
		}
	}
	return db.NewSQLDataStore(driver, nil)
}

// stackingSchema returns all DDL statements needed for the progressive stacking test.
func stackingSchema() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS ecosystems (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL UNIQUE, description TEXT, theme TEXT, build_args TEXT, ca_certs TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS domains (id INTEGER PRIMARY KEY AUTOINCREMENT, ecosystem_id INTEGER NOT NULL, name TEXT NOT NULL, description TEXT, theme TEXT, build_args TEXT, ca_certs TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY (ecosystem_id) REFERENCES ecosystems(id) ON DELETE CASCADE, UNIQUE(ecosystem_id, name))`,
		`CREATE TABLE IF NOT EXISTS git_repos (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL UNIQUE, url TEXT NOT NULL, slug TEXT NOT NULL UNIQUE, default_ref TEXT NOT NULL DEFAULT 'main', auth_type TEXT NOT NULL CHECK(auth_type IN ('none','ssh','token')), credential_id INTEGER, auto_sync BOOLEAN NOT NULL DEFAULT 0, sync_interval_minutes INTEGER NOT NULL DEFAULT 0, last_synced_at DATETIME, sync_status TEXT NOT NULL DEFAULT 'pending' CHECK(sync_status IN ('pending','syncing','synced','error')), sync_error TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS apps (id INTEGER PRIMARY KEY AUTOINCREMENT, domain_id INTEGER NOT NULL, name TEXT NOT NULL, path TEXT NOT NULL DEFAULT '', description TEXT, theme TEXT, language TEXT, build_config TEXT, git_repo_id INTEGER, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY (domain_id) REFERENCES domains(id), UNIQUE(domain_id, name))`,
		`CREATE TABLE IF NOT EXISTS workspaces (id INTEGER PRIMARY KEY AUTOINCREMENT, app_id INTEGER NOT NULL, name TEXT NOT NULL, description TEXT, image_name TEXT, container_id TEXT, status TEXT DEFAULT 'stopped', nvim_structure TEXT, nvim_plugins TEXT, theme TEXT, terminal_prompt TEXT, terminal_plugins TEXT, terminal_package TEXT, slug TEXT, ssh_agent_forwarding INTEGER DEFAULT 0, git_repo_id INTEGER, env TEXT NOT NULL DEFAULT '{}', build_config TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY (app_id) REFERENCES apps(id), UNIQUE(app_id, name))`,
		`CREATE TABLE IF NOT EXISTS credentials (id INTEGER PRIMARY KEY AUTOINCREMENT, scope_type TEXT NOT NULL CHECK(scope_type IN ('ecosystem','domain','app','workspace')), scope_id INTEGER, name TEXT NOT NULL, source TEXT NOT NULL CHECK(source IN ('vault','env')), vault_secret TEXT, vault_env TEXT, vault_username_secret TEXT, vault_fields TEXT, env_var TEXT, description TEXT, username_var TEXT, password_var TEXT, expires_at DATETIME, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, UNIQUE(scope_type, scope_id, name))`,
		`CREATE TABLE IF NOT EXISTS registries (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL UNIQUE, type TEXT NOT NULL, version TEXT NOT NULL DEFAULT '', enabled BOOLEAN NOT NULL DEFAULT 1, lifecycle TEXT NOT NULL DEFAULT 'manual', port INTEGER NOT NULL UNIQUE, storage TEXT NOT NULL DEFAULT '', idle_timeout INTEGER DEFAULT 1800, config TEXT, description TEXT, status TEXT DEFAULT 'stopped', created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS nvim_plugins (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL UNIQUE, description TEXT, repo TEXT NOT NULL, branch TEXT, version TEXT, priority INTEGER, lazy INTEGER DEFAULT 0, event TEXT, ft TEXT, keys TEXT, cmd TEXT, dependencies TEXT, build TEXT, config TEXT, init TEXT, opts TEXT, keymaps TEXT, category TEXT, tags TEXT, enabled INTEGER DEFAULT 1, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS terminal_prompts (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL UNIQUE, description TEXT, type TEXT NOT NULL, add_newline BOOLEAN DEFAULT TRUE, palette TEXT, format TEXT, modules TEXT, character TEXT, palette_ref TEXT, colors TEXT, raw_config TEXT, category TEXT, tags TEXT, enabled BOOLEAN DEFAULT TRUE, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS defaults (key TEXT PRIMARY KEY, value TEXT NOT NULL, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS context (id INTEGER PRIMARY KEY CHECK (id = 1), active_ecosystem_id INTEGER, active_domain_id INTEGER, active_app_id INTEGER, active_workspace_id INTEGER, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP)`,
		`INSERT OR IGNORE INTO context (id) VALUES (1)`,
		`CREATE TABLE IF NOT EXISTS nvim_themes (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL UNIQUE, description TEXT, author TEXT, category TEXT, plugin_repo TEXT NOT NULL, plugin_branch TEXT, plugin_tag TEXT, style TEXT, transparent BOOLEAN DEFAULT FALSE, colors TEXT, options TEXT, is_active BOOLEAN DEFAULT FALSE, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS nvim_packages (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL UNIQUE, description TEXT, category TEXT, labels TEXT, plugins TEXT NOT NULL, extends TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS terminal_plugins (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL UNIQUE, description TEXT, repo TEXT NOT NULL, category TEXT, shell TEXT NOT NULL DEFAULT 'zsh', manager TEXT NOT NULL DEFAULT 'manual', load_command TEXT, source_file TEXT, dependencies TEXT NOT NULL DEFAULT '[]', env_vars TEXT NOT NULL DEFAULT '{}', labels TEXT NOT NULL DEFAULT '{}', enabled BOOLEAN NOT NULL DEFAULT 1, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS terminal_packages (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL UNIQUE, description TEXT, category TEXT, labels TEXT, plugins TEXT NOT NULL DEFAULT '[]', prompts TEXT NOT NULL DEFAULT '[]', profiles TEXT NOT NULL DEFAULT '[]', wezterm TEXT, extends TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS custom_resource_definitions (id INTEGER PRIMARY KEY AUTOINCREMENT, kind TEXT NOT NULL UNIQUE, "group" TEXT NOT NULL, singular TEXT NOT NULL, plural TEXT NOT NULL, short_names TEXT, scope TEXT NOT NULL CHECK(scope IN ('Global', 'Workspace', 'App', 'Domain', 'Ecosystem')), versions TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS custom_resources (id INTEGER PRIMARY KEY AUTOINCREMENT, kind TEXT NOT NULL, name TEXT NOT NULL, namespace TEXT, spec TEXT, status TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, UNIQUE(kind, name, namespace))`,
	}
}

// exportSnapshot collects all resources via handler List() methods and returns
// YAML string + per-kind count map.
//
// NOTE: This function calls handler.List() directly on each registered handler.
// This is intentionally different from cmd/get_all.go, which calls
// ds.ListEcosystems() and then constructs model objects via model constructors.
// This test validates the handler layer contract — not the exact CLI output path.
func exportSnapshot(t *testing.T, ds *db.SQLDataStore) (string, map[string]int) {
	t.Helper()
	RegisterAll()
	ctx := resource.Context{DataStore: ds}

	var res []resource.Resource
	listFrom := func(kind string) {
		if items, err := resource.List(ctx, kind); err == nil {
			res = append(res, items...)
		}
	}
	appendFrom := func(h interface {
		List(resource.Context) ([]resource.Resource, error)
	}) {
		if items, err := h.List(ctx); err == nil {
			res = append(res, items...)
		}
	}

	// Export in topological order so that ApplyList restore succeeds:
	// parent resources must precede child/dependent resources.
	// Credentials come LAST in the hierarchy so scope targets (app, workspace)
	// exist when credentials are re-applied during restore.
	listFrom(KindGlobalDefaults)
	appendFrom(NewEcosystemHandler())
	appendFrom(NewDomainHandler())
	appendFrom(NewGitRepoHandler())
	appendFrom(NewRegistryHandler())
	appendFrom(NewAppHandler())
	appendFrom(NewWorkspaceHandler())
	appendFrom(NewCredentialHandler())
	listFrom(KindNvimPlugin)
	listFrom(KindTerminalPrompt)
	appendFrom(NewNvimThemeHandler())
	appendFrom(NewNvimPackageHandler())
	appendFrom(NewTerminalPluginHandler())
	appendFrom(NewTerminalPackageHandler())
	appendFrom(NewCRDHandler())

	list, err := resource.BuildList(ctx, res)
	if err != nil {
		t.Fatalf("BuildList: %v", err)
	}
	yamlBytes, err := yaml.Marshal(list)
	if err != nil {
		t.Fatalf("yaml.Marshal: %v", err)
	}

	counts := make(map[string]int)
	for _, item := range list.Items {
		if m, ok := item.(map[string]any); ok {
			if k, ok := m["kind"].(string); ok {
				counts[k]++
			}
		}
	}
	return string(yamlBytes), counts
}

// checkNonDecreasing asserts each kind count in after is >= the count in before.
func checkNonDecreasing(t *testing.T, before, after map[string]int, step string) {
	t.Helper()
	for kind, prev := range before {
		if after[kind] < prev {
			t.Errorf("step %s: %s count fell %d→%d (resources lost)", step, kind, prev, after[kind])
		}
	}
}

// mustContain asserts the YAML string contains needle.
func mustContain(t *testing.T, yamlStr, needle, ctx string) {
	t.Helper()
	if !strings.Contains(yamlStr, needle) {
		t.Errorf("%s: YAML missing %q", ctx, needle)
	}
}

// doApply applies a YAML document through its registered handler.
func doApply(t *testing.T, ctx resource.Context, kind, yamlStr string) {
	t.Helper()
	h, err := resource.MustGetHandler(kind)
	if err != nil {
		t.Fatalf("no handler for %q: %v", kind, err)
	}
	if _, err := h.Apply(ctx, []byte(yamlStr)); err != nil {
		t.Fatalf("Apply %s: %v", kind, err)
	}
}

// TestProgressiveStacking verifies that incrementally building the hierarchy
// produces a YAML export that is a strict superset at each step, and that the
// final export round-trips through wipe → ApplyList → re-export perfectly.
func TestProgressiveStacking(t *testing.T) {
	RegisterAll()

	ds := createStackingDS(t)
	defer ds.Close()
	ctx := resource.Context{DataStore: ds}

	// Step 1: Ecosystem + GlobalDefaults
	doApply(t, ctx, KindEcosystem, "apiVersion: devopsmaestro.io/v1\nkind: Ecosystem\nmetadata:\n  name: prod-eco\nspec:\n  theme: tokyonight\n")
	doApply(t, ctx, KindGlobalDefaults, "apiVersion: devopsmaestro.io/v1\nkind: GlobalDefaults\nmetadata:\n  name: global-defaults\nspec:\n  buildArgs:\n    GOENV: production\n")
	y1, c1 := exportSnapshot(t, ds)
	mustContain(t, y1, "prod-eco", "step1")
	mustContain(t, y1, "GOENV", "step1")
	if c1[KindEcosystem] < 1 {
		t.Errorf("step1: Ecosystem count = %d, want >=1", c1[KindEcosystem])
	}
	t.Logf("step1 counts: %v", c1)

	// Step 2: Domain + Credential
	doApply(t, ctx, KindDomain, "apiVersion: devopsmaestro.io/v1\nkind: Domain\nmetadata:\n  name: backend\n  ecosystem: prod-eco\nspec: {}\n")
	doApply(t, ctx, KindCredential, "apiVersion: devopsmaestro.io/v1\nkind: Credential\nmetadata:\n  name: gh-token\n  ecosystem: prod-eco\nspec:\n  source: vault\n  vaultSecret: github/token\n")
	y2, c2 := exportSnapshot(t, ds)
	checkNonDecreasing(t, c1, c2, "2")
	mustContain(t, y2, "prod-eco", "step2 eco")
	mustContain(t, y2, "GOENV", "step2 build-arg")
	mustContain(t, y2, "backend", "step2 domain")
	mustContain(t, y2, "gh-token", "step2 cred")
	t.Logf("step2 counts: %v", c2)

	// Step 3: GitRepo + App
	doApply(t, ctx, KindGitRepo, "apiVersion: devopsmaestro.io/v1\nkind: GitRepo\nmetadata:\n  name: api-repo\nspec:\n  url: https://github.com/example/api\n  authType: none\n")
	doApply(t, ctx, KindApp, "apiVersion: devopsmaestro.io/v1\nkind: App\nmetadata:\n  name: api-server\n  domain: backend\n  ecosystem: prod-eco\nspec:\n  path: /apps/api\n")
	y3, c3 := exportSnapshot(t, ds)
	checkNonDecreasing(t, c2, c3, "3")
	mustContain(t, y3, "gh-token", "step3 cred survives")
	mustContain(t, y3, "api-repo", "step3 gitrepo")
	mustContain(t, y3, "api-server", "step3 app")
	t.Logf("step3 counts: %v", c3)

	// Step 4: Workspace + NvimPlugin + TerminalPrompt
	doApply(t, ctx, KindWorkspace, "apiVersion: devopsmaestro.io/v1\nkind: Workspace\nmetadata:\n  name: dev\n  app: api-server\n  domain: backend\n  ecosystem: prod-eco\nspec:\n  image:\n    name: ubuntu:22.04\n")
	doApply(t, ctx, KindNvimPlugin, "apiVersion: devopsmaestro.io/v1\nkind: NvimPlugin\nmetadata:\n  name: telescope\nspec:\n  repo: nvim-telescope/telescope.nvim\n  lazy: true\n")
	// Note: TerminalPrompt uses devopsmaestro.io/v1 (same as all other kinds).
	// The terminal/v1alpha1 apiVersion was accepted by prompt.Parse (no validation)
	// but is inconsistent — the exported YAML always emits devopsmaestro.io/v1.
	doApply(t, ctx, KindTerminalPrompt, "apiVersion: devopsmaestro.io/v1\nkind: TerminalPrompt\nmetadata:\n  name: starship-default\nspec:\n  type: starship\n")
	y4, c4 := exportSnapshot(t, ds)
	checkNonDecreasing(t, c3, c4, "4")
	mustContain(t, y4, "api-server", "step4 app survives")
	mustContain(t, y4, "dev", "step4 workspace")
	mustContain(t, y4, "telescope", "step4 nvim plugin")
	mustContain(t, y4, "starship-default", "step4 terminal prompt")
	t.Logf("step4 counts: %v", c4)

	// Step 5: Registry + updated GlobalDefaults
	doApply(t, ctx, KindRegistry, "apiVersion: devopsmaestro.io/v1\nkind: Registry\nmetadata:\n  name: my-zot\nspec:\n  type: zot\n  port: 5100\n  lifecycle: persistent\n")
	doApply(t, ctx, KindGlobalDefaults, "apiVersion: devopsmaestro.io/v1\nkind: GlobalDefaults\nmetadata:\n  name: global-defaults\nspec:\n  buildArgs:\n    GOENV: production\n    DOCKER_REGISTRY: my-zot\n  theme: catppuccin\n")
	y5, c5 := exportSnapshot(t, ds)
	checkNonDecreasing(t, c4, c5, "5")
	mustContain(t, y5, "dev", "step5 workspace survives")
	mustContain(t, y5, "telescope", "step5 nvim plugin survives")
	mustContain(t, y5, "my-zot", "step5 registry")
	mustContain(t, y5, "DOCKER_REGISTRY", "step5 global build-arg")
	mustContain(t, y5, "catppuccin", "step5 global theme")
	t.Logf("step5 counts: %v", c5)

	// Final: Wipe → Restore → Re-export → Verify parity
	ds2 := createStackingDS(t)
	defer ds2.Close()
	ctx2 := resource.Context{DataStore: ds2}

	applied, err := resource.ApplyList(ctx2, []byte(y5))
	if err != nil {
		t.Errorf("ApplyList errors: %v (applied %d)", err, len(applied))
	}

	yr, cr := exportSnapshot(t, ds2)
	for kind, want := range c5 {
		if cr[kind] < want {
			t.Errorf("restore: %s count = %d, want >=%d", kind, cr[kind], want)
		}
	}
	for _, needle := range []string{"prod-eco", "backend", "api-server", "dev", "telescope",
		"starship-default", "gh-token", "api-repo", "my-zot", "catppuccin"} {
		mustContain(t, yr, needle, "restore")
	}
	t.Logf("step5 total=%d, restored total=%d", sumCounts(c5), sumCounts(cr))

	// Step 7: Idempotency — apply the same YAML again on the already-restored DB.
	// Core kubectl principle: applying the same manifest twice must produce no
	// duplicates. All handlers must behave as upserts, not inserts.
	applied2, err2 := resource.ApplyList(ctx2, []byte(yr))
	if err2 != nil {
		t.Errorf("idempotency ApplyList errors: %v (applied %d)", err2, len(applied2))
	}

	yr2, cr2 := exportSnapshot(t, ds2)
	for kind, want := range cr {
		if cr2[kind] != want {
			t.Errorf("idempotency: %s count changed %d→%d (apply twice must not duplicate or lose resources)", kind, want, cr2[kind])
		}
	}
	for _, needle := range []string{"prod-eco", "backend", "api-server", "dev", "telescope",
		"starship-default", "gh-token", "api-repo", "my-zot", "catppuccin"} {
		mustContain(t, yr2, needle, "idempotency")
	}
	t.Logf("idempotency: first restore total=%d, second apply total=%d", sumCounts(cr), sumCounts(cr2))
}

func sumCounts(m map[string]int) int {
	n := 0
	for _, v := range m {
		n += v
	}
	return n
}
