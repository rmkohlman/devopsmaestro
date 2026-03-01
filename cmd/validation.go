package cmd

import (
	"fmt"
	"strings"
)

// ValidateResourceName validates a resource name is not empty or whitespace.
// Returns an error if the name is empty, only whitespace, or otherwise invalid.
func ValidateResourceName(name, resourceType string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%s name cannot be empty", resourceType)
	}
	return nil
}
