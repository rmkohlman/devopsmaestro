package nvimops

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultNvimConfig(t *testing.T) {
	config := DefaultNvimConfig()

	assert.Equal(t, DefaultStructure, config.Structure)
	assert.Equal(t, "lazyvim", config.Structure)

	assert.Equal(t, DefaultPackage, config.PluginPackage)
	assert.Equal(t, "core", config.PluginPackage)

	assert.Equal(t, "", config.Theme) // Let theme resolution handle via cascade
	assert.Equal(t, "append", config.MergeMode)
	assert.Nil(t, config.Plugins) // Use package plugins only
}

func TestDefaultConstants(t *testing.T) {
	assert.Equal(t, "lazyvim", DefaultStructure)
	assert.Equal(t, "core", DefaultPackage)
}
