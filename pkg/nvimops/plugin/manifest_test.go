package plugin

import (
	"testing"
)

func TestResolveManifest(t *testing.T) {
	tests := []struct {
		name    string
		plugins []*Plugin
		want    PluginFeatures
		wantLen int
	}{
		{
			name: "detects Mason from repo",
			plugins: []*Plugin{
				{Name: "mason", Repo: "williamboman/mason.nvim", Enabled: true},
			},
			want: PluginFeatures{
				HasMason:      true,
				HasTreesitter: false,
				HasTelescope:  false,
				HasLSPConfig:  false,
			},
			wantLen: 1,
		},
		{
			name: "detects Treesitter from name",
			plugins: []*Plugin{
				{Name: "nvim-treesitter", Repo: "nvim-treesitter/nvim-treesitter", Enabled: true},
			},
			want: PluginFeatures{
				HasMason:      false,
				HasTreesitter: true,
				HasTelescope:  false,
				HasLSPConfig:  false,
			},
			wantLen: 1,
		},
		{
			name: "detects Telescope from repo",
			plugins: []*Plugin{
				{Name: "telescope", Repo: "nvim-telescope/telescope.nvim", Enabled: true},
			},
			want: PluginFeatures{
				HasMason:      false,
				HasTreesitter: false,
				HasTelescope:  true,
				HasLSPConfig:  false,
			},
			wantLen: 1,
		},
		{
			name: "detects LSPConfig",
			plugins: []*Plugin{
				{Name: "lspconfig", Repo: "neovim/nvim-lspconfig", Enabled: true},
			},
			want: PluginFeatures{
				HasMason:      false,
				HasTreesitter: false,
				HasTelescope:  false,
				HasLSPConfig:  true,
			},
			wantLen: 1,
		},
		{
			name: "detects multiple features",
			plugins: []*Plugin{
				{Name: "mason", Repo: "williamboman/mason.nvim", Enabled: true},
				{Name: "nvim-treesitter", Repo: "nvim-treesitter/nvim-treesitter", Enabled: true},
				{Name: "telescope", Repo: "nvim-telescope/telescope.nvim", Enabled: true},
				{Name: "lspconfig", Repo: "neovim/nvim-lspconfig", Enabled: true},
			},
			want: PluginFeatures{
				HasMason:      true,
				HasTreesitter: true,
				HasTelescope:  true,
				HasLSPConfig:  true,
			},
			wantLen: 4,
		},
		{
			name: "skips disabled plugins",
			plugins: []*Plugin{
				{Name: "mason", Repo: "williamboman/mason.nvim", Enabled: false},
				{Name: "plenary", Repo: "nvim-lua/plenary.nvim", Enabled: true},
			},
			want: PluginFeatures{
				HasMason:      false,
				HasTreesitter: false,
				HasTelescope:  false,
				HasLSPConfig:  false,
			},
			wantLen: 1,
		},
		{
			name:    "empty plugin list",
			plugins: []*Plugin{},
			want: PluginFeatures{
				HasMason:      false,
				HasTreesitter: false,
				HasTelescope:  false,
				HasLSPConfig:  false,
			},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := ResolveManifest(tt.plugins)

			if len(manifest.InstalledPlugins) != tt.wantLen {
				t.Errorf("InstalledPlugins length = %d, want %d", len(manifest.InstalledPlugins), tt.wantLen)
			}

			if manifest.Features.HasMason != tt.want.HasMason {
				t.Errorf("HasMason = %v, want %v", manifest.Features.HasMason, tt.want.HasMason)
			}

			if manifest.Features.HasTreesitter != tt.want.HasTreesitter {
				t.Errorf("HasTreesitter = %v, want %v", manifest.Features.HasTreesitter, tt.want.HasTreesitter)
			}

			if manifest.Features.HasTelescope != tt.want.HasTelescope {
				t.Errorf("HasTelescope = %v, want %v", manifest.Features.HasTelescope, tt.want.HasTelescope)
			}

			if manifest.Features.HasLSPConfig != tt.want.HasLSPConfig {
				t.Errorf("HasLSPConfig = %v, want %v", manifest.Features.HasLSPConfig, tt.want.HasLSPConfig)
			}
		})
	}
}

func TestResolveManifestFromNames(t *testing.T) {
	tests := []struct {
		name        string
		pluginNames []string
		want        PluginFeatures
	}{
		{
			name:        "detects Mason from name",
			pluginNames: []string{"mason"},
			want: PluginFeatures{
				HasMason:      true,
				HasTreesitter: false,
				HasTelescope:  false,
				HasLSPConfig:  false,
			},
		},
		{
			name:        "detects multiple features from names",
			pluginNames: []string{"mason", "nvim-treesitter", "telescope", "lspconfig"},
			want: PluginFeatures{
				HasMason:      true,
				HasTreesitter: true,
				HasTelescope:  true,
				HasLSPConfig:  true,
			},
		},
		{
			name:        "case insensitive detection",
			pluginNames: []string{"Mason", "TREESITTER", "Telescope", "LspConfig"},
			want: PluginFeatures{
				HasMason:      true,
				HasTreesitter: true,
				HasTelescope:  true,
				HasLSPConfig:  true,
			},
		},
		{
			name:        "empty list",
			pluginNames: []string{},
			want: PluginFeatures{
				HasMason:      false,
				HasTreesitter: false,
				HasTelescope:  false,
				HasLSPConfig:  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := ResolveManifestFromNames(tt.pluginNames)

			if len(manifest.InstalledPlugins) != len(tt.pluginNames) {
				t.Errorf("InstalledPlugins length = %d, want %d", len(manifest.InstalledPlugins), len(tt.pluginNames))
			}

			if manifest.Features.HasMason != tt.want.HasMason {
				t.Errorf("HasMason = %v, want %v", manifest.Features.HasMason, tt.want.HasMason)
			}

			if manifest.Features.HasTreesitter != tt.want.HasTreesitter {
				t.Errorf("HasTreesitter = %v, want %v", manifest.Features.HasTreesitter, tt.want.HasTreesitter)
			}

			if manifest.Features.HasTelescope != tt.want.HasTelescope {
				t.Errorf("HasTelescope = %v, want %v", manifest.Features.HasTelescope, tt.want.HasTelescope)
			}

			if manifest.Features.HasLSPConfig != tt.want.HasLSPConfig {
				t.Errorf("HasLSPConfig = %v, want %v", manifest.Features.HasLSPConfig, tt.want.HasLSPConfig)
			}
		})
	}
}
