package builders

// GetContainerDefaults returns the default container configuration settings
func GetContainerDefaults() map[string]interface{} {
	return map[string]interface{}{
		"user":       "dev",
		"uid":        1000,
		"gid":        1000,
		"workingDir": "/workspace",
		"command":    []string{"/bin/zsh", "-l"},
	}
}
