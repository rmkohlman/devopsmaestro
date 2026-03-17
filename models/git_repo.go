package models

import (
	"database/sql"
	"time"
)

// KindGitRepoModel is the Kind identifier for GitRepo resources in YAML.
const KindGitRepoModel = "GitRepo"

// GitRepoDB represents a git repository configuration in the database.
type GitRepoDB struct {
	ID                  int            `db:"id"`
	Name                string         `db:"name"`
	URL                 string         `db:"url"`
	Slug                string         `db:"slug"`
	DefaultRef          string         `db:"default_ref"`
	AuthType            string         `db:"auth_type"`
	CredentialID        sql.NullInt64  `db:"credential_id"`
	AutoSync            bool           `db:"auto_sync"`
	SyncIntervalMinutes int            `db:"sync_interval_minutes"`
	LastSyncedAt        sql.NullTime   `db:"last_synced_at"`
	SyncStatus          string         `db:"sync_status"`
	SyncError           sql.NullString `db:"sync_error"`
	CreatedAt           time.Time      `db:"created_at"`
	UpdatedAt           time.Time      `db:"updated_at"`
}

// GitRepoYAML represents the YAML serialization format for a git repository.
type GitRepoYAML struct {
	APIVersion string          `yaml:"apiVersion"`
	Kind       string          `yaml:"kind"`
	Metadata   GitRepoMetadata `yaml:"metadata"`
	Spec       GitRepoSpec     `yaml:"spec"`
}

// GitRepoMetadata contains git repository metadata.
type GitRepoMetadata struct {
	Name        string            `yaml:"name"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// GitRepoSpec contains git repository specification.
type GitRepoSpec struct {
	URL                 string `yaml:"url"`
	DefaultRef          string `yaml:"defaultRef,omitempty"`
	AuthType            string `yaml:"authType,omitempty"`
	AutoSync            bool   `yaml:"autoSync,omitempty"`
	SyncIntervalMinutes int    `yaml:"syncIntervalMinutes,omitempty"`
}

// ToYAML converts a GitRepoDB to YAML format.
func (r *GitRepoDB) ToYAML() GitRepoYAML {
	return GitRepoYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       KindGitRepoModel,
		Metadata: GitRepoMetadata{
			Name:        r.Name,
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
		Spec: GitRepoSpec{
			URL:                 r.URL,
			DefaultRef:          r.DefaultRef,
			AuthType:            r.AuthType,
			AutoSync:            r.AutoSync,
			SyncIntervalMinutes: r.SyncIntervalMinutes,
		},
	}
}

// FromYAML populates a GitRepoDB from YAML format.
func (r *GitRepoDB) FromYAML(y GitRepoYAML) {
	r.Name = y.Metadata.Name
	r.URL = y.Spec.URL

	r.DefaultRef = y.Spec.DefaultRef
	if r.DefaultRef == "" {
		r.DefaultRef = "main"
	}

	r.AuthType = y.Spec.AuthType
	if r.AuthType == "" {
		r.AuthType = "none"
	}

	r.AutoSync = y.Spec.AutoSync
	r.SyncIntervalMinutes = y.Spec.SyncIntervalMinutes
	r.SyncStatus = "pending"
}
