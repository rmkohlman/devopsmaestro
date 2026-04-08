package main

import (
	"testing"

	"devopsmaestro/pkg/terminalbridge/shellgen"
)

func TestParseEnvFlags(t *testing.T) {
	tests := []struct {
		name    string
		flags   []string
		wantErr bool
		wantEnv map[string]string
	}{
		{
			name:    "empty flags",
			flags:   nil,
			wantErr: false,
			wantEnv: nil,
		},
		{
			name:    "single env var",
			flags:   []string{"KEY=value"},
			wantErr: false,
			wantEnv: map[string]string{"KEY": "value"},
		},
		{
			name:    "multiple env vars",
			flags:   []string{"APP_ENV=dev", "PORT=8080"},
			wantErr: false,
			wantEnv: map[string]string{"APP_ENV": "dev", "PORT": "8080"},
		},
		{
			name:    "value with equals sign",
			flags:   []string{"URL=http://host:8080?a=b"},
			wantErr: false,
			wantEnv: map[string]string{"URL": "http://host:8080?a=b"},
		},
		{
			name:    "empty value",
			flags:   []string{"EMPTY="},
			wantErr: false,
			wantEnv: map[string]string{"EMPTY": ""},
		},
		{
			name:    "missing equals",
			flags:   []string{"INVALID"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &shellgen.ShellConfig{}
			err := parseEnvFlags(tt.flags, config)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantEnv == nil {
				return
			}

			for k, want := range tt.wantEnv {
				got, ok := config.EnvVars[k]
				if !ok {
					t.Errorf("missing env var %s", k)
					continue
				}
				if got != want {
					t.Errorf("env var %s = %q, want %q", k, got, want)
				}
			}
		})
	}
}

func TestIndexOf(t *testing.T) {
	tests := []struct {
		input    string
		char     byte
		expected int
	}{
		{"KEY=VALUE", '=', 3},
		{"NOEQUALS", '=', -1},
		{"=START", '=', 0},
		{"END=", '=', 3},
		{"A=B=C", '=', 1},
	}

	for _, tt := range tests {
		got := indexOf(tt.input, tt.char)
		if got != tt.expected {
			t.Errorf("indexOf(%q, %c) = %d, want %d", tt.input, tt.char, got, tt.expected)
		}
	}
}
