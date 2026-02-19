package shell

// GetDefaults returns the default shell configuration settings
func GetDefaults() map[string]interface{} {
	return map[string]interface{}{
		"type":      "zsh",
		"framework": "oh-my-zsh",
		"theme":     "starship",
	}
}
