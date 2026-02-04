package source

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestIsURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"http://example.com/file.yaml", true},
		{"https://example.com/file.yaml", true},
		{"github:user/repo/path/file.yaml", true},
		{"./local-file.yaml", false},
		{"/absolute/path.yaml", false},
		{"relative/path.yaml", false},
		{"-", false},
		{"", false},
		{"httpx://not-a-url", false},
		{"github:", true}, // Technically starts with github:
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := IsURL(tt.input)
			if got != tt.want {
				t.Errorf("IsURL(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolve(t *testing.T) {
	tests := []struct {
		input    string
		wantType string
	}{
		{"-", "stdin"},
		{"http://example.com/file.yaml", "url"},
		{"https://example.com/file.yaml", "url"},
		{"github:user/repo/path/file.yaml", "github"},
		{"./local-file.yaml", "file"},
		{"/absolute/path.yaml", "file"},
		{"relative/path.yaml", "file"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			src := Resolve(tt.input)
			if src.Type() != tt.wantType {
				t.Errorf("Resolve(%q).Type() = %v, want %v", tt.input, src.Type(), tt.wantType)
			}
		})
	}
}

func TestFileSource_Read(t *testing.T) {
	// Create a temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")
	content := []byte("kind: NvimPlugin\nname: test")
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	t.Run("read existing file", func(t *testing.T) {
		src := &FileSource{Path: tmpFile}
		data, displayName, err := src.Read()
		if err != nil {
			t.Fatalf("Read() error = %v", err)
		}
		if string(data) != string(content) {
			t.Errorf("Read() data = %q, want %q", data, content)
		}
		if displayName != tmpFile {
			t.Errorf("Read() displayName = %q, want %q", displayName, tmpFile)
		}
	})

	t.Run("read non-existent file", func(t *testing.T) {
		src := &FileSource{Path: "/non/existent/file.yaml"}
		_, _, err := src.Read()
		if err == nil {
			t.Error("Read() should return error for non-existent file")
		}
	})
}

func TestURLSource_Read(t *testing.T) {
	t.Run("successful fetch", func(t *testing.T) {
		content := "kind: NvimPlugin\nname: test"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(content))
		}))
		defer server.Close()

		src := &URLSource{URL: server.URL}
		data, displayName, err := src.Read()
		if err != nil {
			t.Fatalf("Read() error = %v", err)
		}
		if string(data) != content {
			t.Errorf("Read() data = %q, want %q", data, content)
		}
		if displayName != server.URL {
			t.Errorf("Read() displayName = %q, want %q", displayName, server.URL)
		}
	})

	t.Run("404 error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		src := &URLSource{URL: server.URL}
		_, _, err := src.Read()
		if err == nil {
			t.Error("Read() should return error for 404")
		}
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		src := &URLSource{URL: server.URL}
		_, _, err := src.Read()
		if err == nil {
			t.Error("Read() should return error for 500")
		}
	})
}

func TestGitHubSource(t *testing.T) {
	t.Run("shorthand conversion", func(t *testing.T) {
		src := NewGitHubSource("github:rmkohlman/nvim-yaml-plugins/plugins/telescope.yaml")

		expectedURL := "https://raw.githubusercontent.com/rmkohlman/nvim-yaml-plugins/main/plugins/telescope.yaml"
		if src.URL != expectedURL {
			t.Errorf("URL = %q, want %q", src.URL, expectedURL)
		}
		if src.Original != "github:rmkohlman/nvim-yaml-plugins/plugins/telescope.yaml" {
			t.Errorf("Original = %q, want original shorthand", src.Original)
		}
		if src.Type() != "github" {
			t.Errorf("Type() = %q, want github", src.Type())
		}
	})

	t.Run("nested path conversion", func(t *testing.T) {
		src := NewGitHubSource("github:user/repo/path/to/deep/file.yaml")

		expectedURL := "https://raw.githubusercontent.com/user/repo/main/path/to/deep/file.yaml"
		if src.URL != expectedURL {
			t.Errorf("URL = %q, want %q", src.URL, expectedURL)
		}
	})

	t.Run("actual fetch via httptest", func(t *testing.T) {
		content := "kind: NvimPlugin\nname: test"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(content))
		}))
		defer server.Close()

		// Create a GitHubSource but override the inner URL
		src := &GitHubSource{
			Original: "github:test/repo/file.yaml",
			URL:      server.URL,
			inner:    URLSource{URL: server.URL},
		}

		data, displayName, err := src.Read()
		if err != nil {
			t.Fatalf("Read() error = %v", err)
		}
		if string(data) != content {
			t.Errorf("Read() data = %q, want %q", data, content)
		}
		if displayName != server.URL {
			t.Errorf("Read() displayName = %q, want %q", displayName, server.URL)
		}
	})
}

func TestReadAll(t *testing.T) {
	// Create temp files
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.yaml")
	file2 := filepath.Join(tmpDir, "file2.yaml")
	os.WriteFile(file1, []byte("content1"), 0644)
	os.WriteFile(file2, []byte("content2"), 0644)

	results := ReadAll([]string{file1, file2})

	if len(results) != 2 {
		t.Fatalf("ReadAll() returned %d results, want 2", len(results))
	}

	if results[0].Err != nil {
		t.Errorf("results[0].Err = %v, want nil", results[0].Err)
	}
	if string(results[0].Data) != "content1" {
		t.Errorf("results[0].Data = %q, want content1", results[0].Data)
	}

	if results[1].Err != nil {
		t.Errorf("results[1].Err = %v, want nil", results[1].Err)
	}
	if string(results[1].Data) != "content2" {
		t.Errorf("results[1].Data = %q, want content2", results[1].Data)
	}
}

func TestReadAll_WithErrors(t *testing.T) {
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "exists.yaml")
	os.WriteFile(existingFile, []byte("content"), 0644)

	results := ReadAll([]string{existingFile, "/non/existent/file.yaml"})

	if len(results) != 2 {
		t.Fatalf("ReadAll() returned %d results, want 2", len(results))
	}

	// First should succeed
	if results[0].Err != nil {
		t.Errorf("results[0].Err = %v, want nil", results[0].Err)
	}

	// Second should fail
	if results[1].Err == nil {
		t.Error("results[1].Err should not be nil for non-existent file")
	}
}
