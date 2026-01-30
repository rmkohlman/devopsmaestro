package plugin

import (
	"strings"
	"testing"
)

// MockGenerator is a test implementation of LuaGenerator.
type MockGenerator struct {
	GenerateLuaCalls     int
	GenerateLuaFileCalls int
	ReturnError          error
	ReturnLua            string
}

func (m *MockGenerator) GenerateLua(p *Plugin) (string, error) {
	m.GenerateLuaCalls++
	if m.ReturnError != nil {
		return "", m.ReturnError
	}
	if m.ReturnLua != "" {
		return m.ReturnLua, nil
	}
	return "return { \"" + p.Repo + "\" }", nil
}

func (m *MockGenerator) GenerateLuaFile(p *Plugin) (string, error) {
	m.GenerateLuaFileCalls++
	if m.ReturnError != nil {
		return "", m.ReturnError
	}
	lua, _ := m.GenerateLua(p)
	return "-- " + p.Name + "\n" + lua, nil
}

// Verify MockGenerator implements LuaGenerator
var _ LuaGenerator = (*MockGenerator)(nil)

// TestGeneratorInterface verifies the Generator interface contract.
func TestGeneratorInterface(t *testing.T) {
	t.Run("Generator implements LuaGenerator", func(t *testing.T) {
		var _ LuaGenerator = (*Generator)(nil)
		var _ LuaGenerator = NewGenerator()
	})

	t.Run("MockGenerator implements LuaGenerator", func(t *testing.T) {
		var _ LuaGenerator = (*MockGenerator)(nil)
	})
}

// TestGeneratorSwappability verifies different generators can be used interchangeably.
func TestGeneratorSwappability(t *testing.T) {
	testPlugin := &Plugin{
		Name:        "test-plugin",
		Description: "A test plugin",
		Repo:        "test/plugin",
		Branch:      "main",
	}

	// Function that works with any LuaGenerator
	generateLua := func(g LuaGenerator, p *Plugin) (string, error) {
		return g.GenerateLua(p)
	}

	t.Run("DefaultGenerator", func(t *testing.T) {
		g := NewGenerator()
		lua, err := generateLua(g, testPlugin)
		if err != nil {
			t.Fatalf("GenerateLua failed: %v", err)
		}
		if !strings.Contains(lua, "test/plugin") {
			t.Error("Generated Lua should contain repo")
		}
		if !strings.Contains(lua, "branch") {
			t.Error("Generated Lua should contain branch")
		}
	})

	t.Run("MockGenerator", func(t *testing.T) {
		mock := &MockGenerator{ReturnLua: "-- custom lua"}
		lua, err := generateLua(mock, testPlugin)
		if err != nil {
			t.Fatalf("GenerateLua failed: %v", err)
		}
		if lua != "-- custom lua" {
			t.Errorf("Expected custom lua, got: %s", lua)
		}
		if mock.GenerateLuaCalls != 1 {
			t.Errorf("GenerateLua was called %d times, want 1", mock.GenerateLuaCalls)
		}
	})

	t.Run("CustomGenerator", func(t *testing.T) {
		// Custom generator that produces minimal output
		custom := &minimalGenerator{}
		lua, err := generateLua(custom, testPlugin)
		if err != nil {
			t.Fatalf("GenerateLua failed: %v", err)
		}
		if !strings.HasPrefix(lua, "return") {
			t.Error("Custom generator should produce valid Lua")
		}
	})
}

// minimalGenerator is a custom implementation for testing.
type minimalGenerator struct{}

func (g *minimalGenerator) GenerateLua(p *Plugin) (string, error) {
	return "return { \"" + p.Repo + "\" }", nil
}

func (g *minimalGenerator) GenerateLuaFile(p *Plugin) (string, error) {
	lua, err := g.GenerateLua(p)
	if err != nil {
		return "", err
	}
	return "-- Minimal: " + p.Name + "\n" + lua, nil
}

var _ LuaGenerator = (*minimalGenerator)(nil)

// TestGeneratorOptions verifies generator configuration works.
func TestGeneratorOptions(t *testing.T) {
	t.Run("Custom indent size", func(t *testing.T) {
		g := &Generator{IndentSize: 4}
		p := &Plugin{
			Name: "test",
			Repo: "test/test",
		}
		lua, err := g.GenerateLua(p)
		if err != nil {
			t.Fatalf("GenerateLua failed: %v", err)
		}
		// Check for 4-space indentation
		if !strings.Contains(lua, "    \"test/test\"") {
			t.Error("Expected 4-space indentation")
		}
	})
}

// TestGenerateLuaFileFormat verifies the file output format.
func TestGenerateLuaFileFormat(t *testing.T) {
	g := NewGenerator()
	p := &Plugin{
		Name:        "my-plugin",
		Description: "My awesome plugin",
		Repo:        "author/my-plugin",
	}

	lua, err := g.GenerateLuaFile(p)
	if err != nil {
		t.Fatalf("GenerateLuaFile failed: %v", err)
	}

	// Check header format
	lines := strings.Split(lua, "\n")
	if len(lines) < 3 {
		t.Fatal("Generated file should have at least 3 lines")
	}
	if !strings.HasPrefix(lines[0], "-- my-plugin") {
		t.Errorf("First line should be plugin name comment, got: %s", lines[0])
	}
	if !strings.HasPrefix(lines[1], "-- My awesome plugin") {
		t.Errorf("Second line should be description comment, got: %s", lines[1])
	}
	if !strings.Contains(lines[2], "nvim-manager") {
		t.Errorf("Third line should mention nvim-manager, got: %s", lines[2])
	}
}

// TestGenerateLuaComplex tests generation of complex plugin configurations.
func TestGenerateLuaComplex(t *testing.T) {
	p := &Plugin{
		Name:     "complex",
		Repo:     "author/complex",
		Branch:   "main",
		Version:  "v1.0.0",
		Priority: 1000,
		Event:    []string{"BufRead", "BufNewFile"},
		Ft:       []string{"lua", "go"},
		Cmd:      []string{"Complex", "ComplexSetup"},
		Dependencies: []Dependency{
			{Repo: "dep/one"},
			{Repo: "dep/two", Build: "make", Config: true},
		},
		Keys: []Keymap{
			{Key: "<leader>c", Mode: []string{"n", "v"}, Action: ":Complex<cr>", Desc: "Run complex"},
		},
		Build:  ":TSUpdate",
		Config: "require('complex').setup({})",
		Init:   "vim.g.complex = true",
		Opts:   map[string]interface{}{"option1": true, "option2": "value"},
	}

	g := NewGenerator()
	lua, err := g.GenerateLua(p)
	if err != nil {
		t.Fatalf("GenerateLua failed: %v", err)
	}

	// Verify all complex features are present
	checks := []string{
		`"author/complex"`,
		`branch = "main"`,
		`version = "v1.0.0"`,
		`priority = 1000`,
		`event = { "BufRead", "BufNewFile" }`,
		`ft = { "lua", "go" }`,
		`cmd = { "Complex", "ComplexSetup" }`,
		`dependencies = {`,
		`"dep/one"`,
		`build = "make"`,
		`config = true`,
		`keys = {`,
		`"<leader>c"`,
		`mode = { "n", "v" }`,
		`desc = "Run complex"`,
		`build = ":TSUpdate"`,
		`config = function()`,
		`init = function()`,
		`opts = {`,
	}

	for _, check := range checks {
		if !strings.Contains(lua, check) {
			t.Errorf("Generated Lua missing: %q", check)
		}
	}
}
