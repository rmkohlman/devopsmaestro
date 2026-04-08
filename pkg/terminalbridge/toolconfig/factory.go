package toolconfig

import "fmt"

// registry maps tool names to their generator constructors.
var registry = map[string]func() ToolConfigGenerator{
	"bat":       func() ToolConfigGenerator { return &BatGenerator{} },
	"delta":     func() ToolConfigGenerator { return &DeltaGenerator{} },
	"fzf":       func() ToolConfigGenerator { return &FzfGenerator{} },
	"dircolors": func() ToolConfigGenerator { return &DircolorsGenerator{} },
}

// NewToolConfigGenerator returns a generator for the named tool.
// Returns an error if the tool name is not recognized.
func NewToolConfigGenerator(name string) (ToolConfigGenerator, error) {
	ctor, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown tool %q (available: %v)", name, AvailableTools())
	}
	return ctor(), nil
}

// AllGenerators returns a generator for every registered tool.
func AllGenerators() []ToolConfigGenerator {
	gens := make([]ToolConfigGenerator, 0, len(registry))
	for _, ctor := range registry {
		gens = append(gens, ctor())
	}
	return gens
}

// AvailableTools returns the names of all registered tools in sorted order.
func AvailableTools() []string {
	// Return in a deterministic order
	tools := []string{"bat", "delta", "dircolors", "fzf"}
	return tools
}
