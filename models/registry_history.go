package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// RegistryHistory represents a registry configuration history entry
type RegistryHistory struct {
	ID               int64
	RegistryID       int
	Revision         int
	Config           string // JSON snapshot
	Enabled          bool
	Lifecycle        string
	Port             int
	Storage          string
	IdleTimeout      sql.NullInt64
	Action           string // 'start', 'restart', 'config_change', 'rollback'
	Status           string // 'success', 'failed', 'in_progress'
	User             sql.NullString
	ErrorMessage     sql.NullString
	PreviousRevision sql.NullInt64
	RegistryVersion  sql.NullString
	CreatedAt        time.Time
	CompletedAt      sql.NullTime
}

// RegistryHistoryYAML is the clean DTO for JSON/YAML serialization of RegistryHistory.
// It replaces sql.Null* types with plain Go types so marshaled output is clean.
type RegistryHistoryYAML struct {
	ID               int64       `json:"id" yaml:"id"`
	RegistryID       int         `json:"registryId" yaml:"registryId"`
	Revision         int         `json:"revision" yaml:"revision"`
	Config           interface{} `json:"config,omitempty" yaml:"config,omitempty"`
	Enabled          bool        `json:"enabled" yaml:"enabled"`
	Lifecycle        string      `json:"lifecycle" yaml:"lifecycle"`
	Port             int         `json:"port" yaml:"port"`
	Storage          string      `json:"storage" yaml:"storage"`
	IdleTimeout      int64       `json:"idleTimeout,omitempty" yaml:"idleTimeout,omitempty"`
	Action           string      `json:"action" yaml:"action"`
	Status           string      `json:"status" yaml:"status"`
	User             string      `json:"user,omitempty" yaml:"user,omitempty"`
	ErrorMessage     string      `json:"errorMessage,omitempty" yaml:"errorMessage,omitempty"`
	PreviousRevision int64       `json:"previousRevision,omitempty" yaml:"previousRevision,omitempty"`
	RegistryVersion  string      `json:"registryVersion,omitempty" yaml:"registryVersion,omitempty"`
	CreatedAt        time.Time   `json:"createdAt" yaml:"createdAt"`
	CompletedAt      time.Time   `json:"completedAt,omitempty" yaml:"completedAt,omitempty"`
}

// ToYAML converts RegistryHistory to a clean RegistryHistoryYAML DTO
func (h *RegistryHistory) ToYAML() RegistryHistoryYAML {
	y := RegistryHistoryYAML{
		ID:         h.ID,
		RegistryID: h.RegistryID,
		Revision:   h.Revision,
		Enabled:    h.Enabled,
		Lifecycle:  h.Lifecycle,
		Port:       h.Port,
		Storage:    h.Storage,
		Action:     h.Action,
		Status:     h.Status,
		CreatedAt:  h.CreatedAt,
	}

	// Parse Config JSON string into map[string]interface{}
	if h.Config != "" {
		var configMap map[string]interface{}
		if err := json.Unmarshal([]byte(h.Config), &configMap); err == nil {
			y.Config = configMap
		}
	}

	// Map sql.NullInt64 fields
	if h.IdleTimeout.Valid {
		y.IdleTimeout = h.IdleTimeout.Int64
	}
	if h.PreviousRevision.Valid {
		y.PreviousRevision = h.PreviousRevision.Int64
	}

	// Map sql.NullString fields
	if h.User.Valid {
		y.User = h.User.String
	}
	if h.ErrorMessage.Valid {
		y.ErrorMessage = h.ErrorMessage.String
	}
	if h.RegistryVersion.Valid {
		y.RegistryVersion = h.RegistryVersion.String
	}

	// Map sql.NullTime fields
	if h.CompletedAt.Valid {
		y.CompletedAt = h.CompletedAt.Time
	}

	return y
}
