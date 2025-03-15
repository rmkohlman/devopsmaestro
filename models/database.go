package models

// Metadata contains metadata information for the context (e.g., for backups or snapshots).
type Metadata struct {
	Timestamp string `yaml:"timestamp"`
	Type      string `yaml:"type"`
}

// DatabaseContext represents the entire state of the database.
type DatabaseContext struct {
	Metadata     Metadata     `yaml:"metadata"`
	Projects     []Project    `yaml:"projects"`
	Workspaces   []Workspace  `yaml:"workspaces"`
	Dependencies []Dependency `yaml:"dependencies"`
	// Add other relevant fields here
}
