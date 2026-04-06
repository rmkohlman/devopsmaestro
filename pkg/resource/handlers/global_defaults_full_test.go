package handlers

// =============================================================================
// TDD Phase 2 (RED): Bug #177 — GlobalDefaults missing nvim-package,
// terminal-package, plugins, registry defaults, and registry-idle-timeout.
//
// Root cause (per issue #177):
//   - globalDefaultsSpec only has Theme, BuildArgs, CACerts fields
//   - loadGlobalDefaults() only reads "build-args", "ca-certs", "theme"
//   - Apply() only writes those 3 keys back to the defaults store
//   - Delete() only clears "build-args" and "ca-certs" (misses theme too)
//   - GlobalDefaultsResource has no fields for the missing keys
//
// Fix required:
//   1. Add NvimPackage, TerminalPackage, Plugins, RegistryOCI, RegistryPyPI,
//      RegistryNPM, RegistryGo, RegistryHTTP, RegistryIdleTimeout fields to
//      globalDefaultsSpec and GlobalDefaultsResource
//   2. Update loadGlobalDefaults() to read all 10 keys
//   3. Update Apply() to write all 10 keys to the defaults store
//   4. Update Delete() to clear ALL keys (not just build-args and ca-certs)
//
// ALL tests in this file MUST FAIL until the fix is implemented.
// =============================================================================

import (
	"encoding/json"
	"strings"
	"testing"

	"devopsmaestro/db"
	"github.com/rmkohlman/MaestroSDK/resource"
)

// =============================================================================
// Test 1: ToYAML must include nvim-package when set in the defaults store
//
// RED: GlobalDefaultsResource has no NvimPackage field. loadGlobalDefaults()
// never reads "nvim-package" from the store. ToYAML() never emits
// "nvimPackage:" in the spec output.
// =============================================================================

func TestGlobalDefaults_ExportsNvimPackage(t *testing.T) {
	h := NewGlobalDefaultsHandler()
	store := db.NewMockDataStore()

	if err := store.SetDefault("nvim-package", "astronvim"); err != nil {
		t.Fatalf("SetDefault(nvim-package): %v", err)
	}

	ctx := resource.Context{DataStore: store}

	res, err := h.Get(ctx, "global-defaults")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	yamlStr := string(yamlBytes)

	// RED: loadGlobalDefaults() never reads "nvim-package" → GlobalDefaultsResource
	// carries no nvimPackage → ToYAML() never emits "nvimPackage:" → this fails.
	if !strings.Contains(yamlStr, "nvimPackage:") {
		t.Errorf(
			"ToYAML() output must contain 'nvimPackage:' field when 'nvim-package' default is set\n"+
				"Bug: globalDefaultsSpec has no NvimPackage field and loadGlobalDefaults() "+
				"does not read 'nvim-package' from the defaults table\ngot YAML:\n%s",
			yamlStr,
		)
	}

	if !strings.Contains(yamlStr, "astronvim") {
		t.Errorf(
			"ToYAML() output must contain the nvim-package value 'astronvim'\ngot YAML:\n%s",
			yamlStr,
		)
	}
}

// =============================================================================
// Test 2: ToYAML must include terminal-package when set in the defaults store
//
// RED: globalDefaultsSpec has no TerminalPackage field. loadGlobalDefaults()
// never reads "terminal-package". ToYAML() never emits "terminalPackage:".
// =============================================================================

func TestGlobalDefaults_ExportsTerminalPackage(t *testing.T) {
	h := NewGlobalDefaultsHandler()
	store := db.NewMockDataStore()

	if err := store.SetDefault("terminal-package", "wezterm"); err != nil {
		t.Fatalf("SetDefault(terminal-package): %v", err)
	}

	ctx := resource.Context{DataStore: store}

	res, err := h.Get(ctx, "global-defaults")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	yamlStr := string(yamlBytes)

	// RED: loadGlobalDefaults() never reads "terminal-package" → this fails.
	if !strings.Contains(yamlStr, "terminalPackage:") {
		t.Errorf(
			"ToYAML() output must contain 'terminalPackage:' field when 'terminal-package' default is set\n"+
				"Bug: globalDefaultsSpec has no TerminalPackage field and loadGlobalDefaults() "+
				"does not read 'terminal-package' from the defaults table\ngot YAML:\n%s",
			yamlStr,
		)
	}

	if !strings.Contains(yamlStr, "wezterm") {
		t.Errorf(
			"ToYAML() output must contain the terminal-package value 'wezterm'\ngot YAML:\n%s",
			yamlStr,
		)
	}
}

// =============================================================================
// Test 3: ToYAML must include plugins when set in the defaults store
//
// RED: globalDefaultsSpec has no Plugins field. loadGlobalDefaults() never
// reads "plugins". ToYAML() never emits "plugins:" in the spec output.
// The "plugins" key stores a JSON array of plugin names as a string.
// =============================================================================

func TestGlobalDefaults_ExportsPlugins(t *testing.T) {
	h := NewGlobalDefaultsHandler()
	store := db.NewMockDataStore()

	// "plugins" is stored as a JSON array string in the defaults table
	pluginsJSON, err := json.Marshal([]string{"nvim-telescope", "nvim-treesitter", "nvim-lspconfig"})
	if err != nil {
		t.Fatalf("json.Marshal plugins: %v", err)
	}

	if err := store.SetDefault("plugins", string(pluginsJSON)); err != nil {
		t.Fatalf("SetDefault(plugins): %v", err)
	}

	ctx := resource.Context{DataStore: store}

	res, err := h.Get(ctx, "global-defaults")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	yamlStr := string(yamlBytes)

	// RED: loadGlobalDefaults() never reads "plugins" → this fails.
	if !strings.Contains(yamlStr, "plugins:") {
		t.Errorf(
			"ToYAML() output must contain 'plugins:' field when 'plugins' default is set\n"+
				"Bug: globalDefaultsSpec has no Plugins field and loadGlobalDefaults() "+
				"does not read 'plugins' from the defaults table\ngot YAML:\n%s",
			yamlStr,
		)
	}

	if !strings.Contains(yamlStr, "nvim-telescope") {
		t.Errorf(
			"ToYAML() output must contain plugin name 'nvim-telescope'\ngot YAML:\n%s",
			yamlStr,
		)
	}
}

// =============================================================================
// Test 4: ToYAML must include all 5 registry type defaults
//
// RED: globalDefaultsSpec has no RegistryOCI, RegistryPyPI, RegistryNPM,
// RegistryGo, or RegistryHTTP fields. loadGlobalDefaults() never reads any
// of the "registry-*" keys from the defaults table.
// =============================================================================

func TestGlobalDefaults_ExportsRegistryDefaults(t *testing.T) {
	h := NewGlobalDefaultsHandler()
	store := db.NewMockDataStore()

	registryDefaults := map[string]string{
		"registry-oci":  "my-zot-registry",
		"registry-pypi": "my-devpi-registry",
		"registry-npm":  "my-verdaccio-registry",
		"registry-go":   "my-athens-registry",
		"registry-http": "my-squid-proxy",
	}

	for key, val := range registryDefaults {
		if err := store.SetDefault(key, val); err != nil {
			t.Fatalf("SetDefault(%s): %v", key, err)
		}
	}

	ctx := resource.Context{DataStore: store}

	res, err := h.Get(ctx, "global-defaults")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	yamlStr := string(yamlBytes)

	// RED: loadGlobalDefaults() never reads any "registry-*" keys → all fail.
	tests := []struct {
		yamlKey string
		value   string
	}{
		{"registryOci:", "my-zot-registry"},
		{"registryPypi:", "my-devpi-registry"},
		{"registryNpm:", "my-verdaccio-registry"},
		{"registryGo:", "my-athens-registry"},
		{"registryHttp:", "my-squid-proxy"},
	}

	for _, tt := range tests {
		if !strings.Contains(yamlStr, tt.yamlKey) {
			t.Errorf(
				"ToYAML() output must contain %q field when the corresponding registry default is set\n"+
					"Bug: globalDefaultsSpec has no field for this registry type and loadGlobalDefaults() "+
					"does not read it from the defaults table\ngot YAML:\n%s",
				tt.yamlKey, yamlStr,
			)
		}
		if !strings.Contains(yamlStr, tt.value) {
			t.Errorf(
				"ToYAML() output must contain registry value %q\ngot YAML:\n%s",
				tt.value, yamlStr,
			)
		}
	}
}

// =============================================================================
// Test 5: ToYAML must include registry-idle-timeout when set
//
// RED: globalDefaultsSpec has no RegistryIdleTimeout field. loadGlobalDefaults()
// never reads "registry-idle-timeout". ToYAML() never emits it.
// =============================================================================

func TestGlobalDefaults_ExportsRegistryIdleTimeout(t *testing.T) {
	h := NewGlobalDefaultsHandler()
	store := db.NewMockDataStore()

	if err := store.SetDefault("registry-idle-timeout", "1h30m"); err != nil {
		t.Fatalf("SetDefault(registry-idle-timeout): %v", err)
	}

	ctx := resource.Context{DataStore: store}

	res, err := h.Get(ctx, "global-defaults")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	yamlStr := string(yamlBytes)

	// RED: loadGlobalDefaults() never reads "registry-idle-timeout" → this fails.
	if !strings.Contains(yamlStr, "registryIdleTimeout:") {
		t.Errorf(
			"ToYAML() output must contain 'registryIdleTimeout:' field when 'registry-idle-timeout' is set\n"+
				"Bug: globalDefaultsSpec has no RegistryIdleTimeout field and loadGlobalDefaults() "+
				"does not read 'registry-idle-timeout' from the defaults table\ngot YAML:\n%s",
			yamlStr,
		)
	}

	if !strings.Contains(yamlStr, "1h30m") {
		t.Errorf(
			"ToYAML() output must contain the registry-idle-timeout value '1h30m'\ngot YAML:\n%s",
			yamlStr,
		)
	}
}

// =============================================================================
// Test 6: Apply must restore ALL keys from a full GlobalDefaults YAML
//
// RED: globalDefaultsSpec only has Theme, BuildArgs, CACerts. yaml.Unmarshal
// silently drops all the new fields. Apply() never calls SetDefault() for
// nvimPackage, terminalPackage, plugins, registry-*, or registry-idle-timeout.
// =============================================================================

func TestGlobalDefaults_ApplyRestoresAllKeys(t *testing.T) {
	h := NewGlobalDefaultsHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	// A full GlobalDefaults YAML document with ALL fields.
	// After the fix, Apply() must write every one of these to the defaults store.
	yamlInput := []byte(`apiVersion: devopsmaestro.io/v1
kind: GlobalDefaults
metadata:
  name: global-defaults
spec:
  theme: catppuccin
  buildArgs:
    GOENV: production
    NODE_ENV: production
  caCerts:
    - name: my-ca
      vaultSecret: secret/ca
  nvimPackage: astronvim
  terminalPackage: wezterm
  plugins:
    - nvim-telescope
    - nvim-treesitter
  registryOci: my-zot
  registryPypi: my-devpi
  registryNpm: my-verdaccio
  registryGo: my-athens
  registryHttp: my-squid
  registryIdleTimeout: 45m
`)

	_, err := h.Apply(ctx, yamlInput)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// Verify every key was written to the defaults store
	checks := []struct {
		key      string
		contains string
	}{
		{"theme", "catppuccin"},
		{"build-args", "GOENV"},
		{"ca-certs", "my-ca"},
		{"nvim-package", "astronvim"},
		{"terminal-package", "wezterm"},
		{"plugins", "nvim-telescope"},
		{"registry-oci", "my-zot"},
		{"registry-pypi", "my-devpi"},
		{"registry-npm", "my-verdaccio"},
		{"registry-go", "my-athens"},
		{"registry-http", "my-squid"},
		{"registry-idle-timeout", "45m"},
	}

	for _, c := range checks {
		got, err := store.GetDefault(c.key)
		if err != nil {
			t.Errorf("GetDefault(%q) after Apply() error = %v", c.key, err)
			continue
		}
		// RED: For all new keys, Apply() never calls SetDefault → got == ""
		if !strings.Contains(got, c.contains) {
			t.Errorf(
				"Apply() must restore %q to defaults store\n"+
					"Bug: Apply() does not handle this field (globalDefaultsSpec missing field or Apply() not calling SetDefault)\n"+
					"got defaults[%q] = %q, want value containing %q",
				c.key, c.key, got, c.contains,
			)
		}
	}
}

// =============================================================================
// Test 7: Delete must clear ALL keys, not just build-args and ca-certs
//
// RED: Delete() currently only calls SetDefault("build-args", "") and
// SetDefault("ca-certs", ""). It misses theme, nvim-package, terminal-package,
// plugins, all registry-* keys, and registry-idle-timeout.
// =============================================================================

func TestGlobalDefaults_DeleteClearsAllKeys(t *testing.T) {
	h := NewGlobalDefaultsHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	// Set every key
	allKeys := map[string]string{
		"theme":                 "catppuccin",
		"build-args":            `{"GOENV":"production"}`,
		"ca-certs":              `[{"name":"my-ca","vaultSecret":"secret/ca"}]`,
		"nvim-package":          "astronvim",
		"terminal-package":      "wezterm",
		"plugins":               `["nvim-telescope","nvim-treesitter"]`,
		"registry-oci":          "my-zot",
		"registry-pypi":         "my-devpi",
		"registry-npm":          "my-verdaccio",
		"registry-go":           "my-athens",
		"registry-http":         "my-squid",
		"registry-idle-timeout": "45m",
	}

	for key, val := range allKeys {
		if err := store.SetDefault(key, val); err != nil {
			t.Fatalf("SetDefault(%q): %v", key, err)
		}
	}

	// Delete all global defaults
	if err := h.Delete(ctx, "global-defaults"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify ALL keys were cleared
	for key := range allKeys {
		got, err := store.GetDefault(key)
		if err != nil {
			t.Errorf("GetDefault(%q) after Delete() error = %v", key, err)
			continue
		}
		// RED: Delete() only clears build-args and ca-certs; all other keys remain.
		if got != "" {
			t.Errorf(
				"Delete() must clear defaults[%q]\n"+
					"Bug: Delete() does not call SetDefault(%q, \"\") — key was not cleared\n"+
					"got defaults[%q] = %q, want \"\"",
				key, key, key, got,
			)
		}
	}
}

// =============================================================================
// Test 8: Full round-trip — all keys preserved through export → wipe → apply
//
// RED: Multiple failures expected:
//   (a) ToYAML() silently drops all new fields (no struct fields for them)
//   (b) Apply() silently drops all new fields from the parsed YAML
//   (c) Delete() leaves behind theme, nvim-package, terminal-package, etc.
//
// This test documents the complete round-trip failure for all missing fields.
// =============================================================================

func TestGlobalDefaults_RoundTrip(t *testing.T) {
	h := NewGlobalDefaultsHandler()

	// ── SOURCE store: all keys set ───────────────────────────────────────────
	srcStore := db.NewMockDataStore()

	pluginsJSON, _ := json.Marshal([]string{"nvim-telescope", "nvim-treesitter"})
	buildArgsJSON, _ := json.Marshal(map[string]string{"GOENV": "production"})
	caCertsJSON, _ := json.Marshal([]map[string]string{{"name": "my-ca", "vaultSecret": "secret/ca"}})

	sourceDefaults := map[string]string{
		"theme":                 "catppuccin",
		"build-args":            string(buildArgsJSON),
		"ca-certs":              string(caCertsJSON),
		"nvim-package":          "astronvim",
		"terminal-package":      "wezterm",
		"plugins":               string(pluginsJSON),
		"registry-oci":          "my-zot",
		"registry-pypi":         "my-devpi",
		"registry-npm":          "my-verdaccio",
		"registry-go":           "my-athens",
		"registry-http":         "my-squid",
		"registry-idle-timeout": "45m",
	}

	for key, val := range sourceDefaults {
		if err := srcStore.SetDefault(key, val); err != nil {
			t.Fatalf("srcStore.SetDefault(%q): %v", key, err)
		}
	}

	srcCtx := resource.Context{DataStore: srcStore}

	// ── Export: Get resource → serialize to YAML ─────────────────────────────
	res, err := h.Get(srcCtx, "global-defaults")
	if err != nil {
		t.Fatalf("Get() from source store error = %v", err)
	}

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	yamlStr := string(yamlBytes)

	// Verify all new fields appear in the exported YAML
	// RED: none of these will be present until the fix is implemented
	exportChecks := []string{
		"nvimPackage:",
		"terminalPackage:",
		"plugins:",
		"registryOci:",
		"registryPypi:",
		"registryNpm:",
		"registryGo:",
		"registryHttp:",
		"registryIdleTimeout:",
	}

	for _, field := range exportChecks {
		if !strings.Contains(yamlStr, field) {
			t.Errorf(
				"ToYAML() round-trip: exported YAML must contain %q\n"+
					"Bug: this field is not in globalDefaultsSpec / not read by loadGlobalDefaults()\ngot YAML:\n%s",
				field, yamlStr,
			)
		}
	}

	// ── Destination store: fresh, apply the exported YAML ───────────────────
	dstStore := db.NewMockDataStore()
	dstCtx := resource.Context{DataStore: dstStore}

	_, err = h.Apply(dstCtx, yamlBytes)
	if err != nil {
		t.Fatalf("Apply() to destination store error = %v", err)
	}

	// Verify all keys were restored in the destination store
	restoreChecks := []struct {
		key      string
		contains string
	}{
		{"theme", "catppuccin"},
		{"nvim-package", "astronvim"},
		{"terminal-package", "wezterm"},
		{"plugins", "nvim-telescope"},
		{"registry-oci", "my-zot"},
		{"registry-pypi", "my-devpi"},
		{"registry-npm", "my-verdaccio"},
		{"registry-go", "my-athens"},
		{"registry-http", "my-squid"},
		{"registry-idle-timeout", "45m"},
	}

	for _, c := range restoreChecks {
		got, err := dstStore.GetDefault(c.key)
		if err != nil {
			t.Errorf("dstStore.GetDefault(%q) after Apply() error = %v", c.key, err)
			continue
		}
		// RED: Apply() doesn't handle these fields → got == "" for all of them
		if !strings.Contains(got, c.contains) {
			t.Errorf(
				"Round-trip Apply() must restore %q in destination store\n"+
					"Bug: Apply() does not handle this field from the YAML spec\n"+
					"got defaults[%q] = %q, want value containing %q",
				c.key, c.key, got, c.contains,
			)
		}
	}
}
