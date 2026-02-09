package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// App represents a codebase/application within a domain.
// The App is "the thing you build and run". It has a path to source code
// and can run in dev mode (Workspace) or live mode (managed by Operator).
//
// Hierarchy: Ecosystem -> Domain -> App -> Workspace
type App struct {
	ID          int            `db:"id" json:"id" yaml:"-"`
	DomainID    int            `db:"domain_id" json:"domain_id" yaml:"-"`
	Name        string         `db:"name" json:"name" yaml:"name"`
	Path        string         `db:"path" json:"path" yaml:"path"`
	Description sql.NullString `db:"description" json:"description,omitempty" yaml:"description,omitempty"`
	// Language and build config stored as JSON in database
	Language    sql.NullString `db:"language" json:"language,omitempty" yaml:"-"`
	BuildConfig sql.NullString `db:"build_config" json:"build_config,omitempty" yaml:"-"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at" yaml:"-"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at" yaml:"-"`
}

// AppYAML represents the YAML serialization format for an app
type AppYAML struct {
	APIVersion string      `yaml:"apiVersion"`
	Kind       string      `yaml:"kind"`
	Metadata   AppMetadata `yaml:"metadata"`
	Spec       AppSpec     `yaml:"spec"`
}

// AppMetadata contains app metadata
type AppMetadata struct {
	Name        string            `yaml:"name"`
	Domain      string            `yaml:"domain"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// AppSpec contains app specification - everything about the codebase
type AppSpec struct {
	Path         string             `yaml:"path"`
	Language     AppLanguageConfig  `yaml:"language,omitempty"`
	Build        AppBuildConfig     `yaml:"build,omitempty"`
	Dependencies AppDependencies    `yaml:"dependencies,omitempty"`
	Services     []AppServiceConfig `yaml:"services,omitempty"`
	Env          map[string]string  `yaml:"env,omitempty"`
	Ports        []string           `yaml:"ports,omitempty"`
	Workspaces   []string           `yaml:"workspaces,omitempty"`
}

// AppLanguageConfig defines the primary language/runtime for the app
type AppLanguageConfig struct {
	Name    string `yaml:"name"`              // go, python, node, rust, java, etc.
	Version string `yaml:"version,omitempty"` // 1.22, 3.11, 20, etc.
}

// AppBuildConfig defines how to build the app
type AppBuildConfig struct {
	Dockerfile string            `yaml:"dockerfile,omitempty"` // Path to Dockerfile (if exists)
	Buildpack  string            `yaml:"buildpack,omitempty"`  // Auto-detect, go, python, etc.
	Args       map[string]string `yaml:"args,omitempty"`       // Build arguments
	Target     string            `yaml:"target,omitempty"`     // Build target stage
	Context    string            `yaml:"context,omitempty"`    // Build context path
}

// AppDependencies defines where the app's dependencies come from
type AppDependencies struct {
	File    string   `yaml:"file,omitempty"`    // go.mod, requirements.txt, package.json
	Install string   `yaml:"install,omitempty"` // Command to install deps
	Extra   []string `yaml:"extra,omitempty"`   // Additional dependencies to install
}

// AppServiceConfig defines services the app needs (databases, caches, etc.)
type AppServiceConfig struct {
	Name    string            `yaml:"name"`              // postgres, redis, mongodb, etc.
	Image   string            `yaml:"image,omitempty"`   // Custom image (default: official)
	Version string            `yaml:"version,omitempty"` // Service version
	Port    int               `yaml:"port,omitempty"`    // Port to expose
	Env     map[string]string `yaml:"env,omitempty"`     // Service environment variables
}

// ToYAML converts an App to YAML format
func (a *App) ToYAML(domainName string) AppYAML {
	description := ""
	if a.Description.Valid {
		description = a.Description.String
	}

	annotations := make(map[string]string)
	if description != "" {
		annotations["description"] = description
	}

	// Parse language from stored JSON if available
	var langConfig AppLanguageConfig
	if a.Language.Valid && a.Language.String != "" {
		_ = json.Unmarshal([]byte(a.Language.String), &langConfig)
	}

	// Parse build config from stored JSON if available
	var buildConfig AppBuildConfig
	if a.BuildConfig.Valid && a.BuildConfig.String != "" {
		_ = json.Unmarshal([]byte(a.BuildConfig.String), &buildConfig)
	}

	return AppYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "App",
		Metadata: AppMetadata{
			Name:        a.Name,
			Domain:      domainName,
			Labels:      make(map[string]string),
			Annotations: annotations,
		},
		Spec: AppSpec{
			Path:       a.Path,
			Language:   langConfig,
			Build:      buildConfig,
			Workspaces: []string{},
		},
	}
}

// FromYAML converts YAML format to an App
func (a *App) FromYAML(yaml AppYAML) {
	a.Name = yaml.Metadata.Name
	a.Path = yaml.Spec.Path

	if desc, ok := yaml.Metadata.Annotations["description"]; ok {
		a.Description = sql.NullString{String: desc, Valid: true}
	}

	// Store language config as JSON
	if yaml.Spec.Language.Name != "" {
		if langJSON, err := json.Marshal(yaml.Spec.Language); err == nil {
			a.Language = sql.NullString{String: string(langJSON), Valid: true}
		}
	}

	// Store build config as JSON
	if yaml.Spec.Build.Dockerfile != "" || yaml.Spec.Build.Buildpack != "" {
		if buildJSON, err := json.Marshal(yaml.Spec.Build); err == nil {
			a.BuildConfig = sql.NullString{String: string(buildJSON), Valid: true}
		}
	}
}
