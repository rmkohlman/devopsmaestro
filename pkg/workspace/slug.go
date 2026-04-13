package workspace

import (
	"fmt"
	"strings"
)

// GenerateSlug creates a workspace slug from hierarchy names.
// Format: {ecosystem}-{domain}-{system}-{app}-{workspace}
// System is optional — when empty, the segment is omitted.
// Names are sanitized: lowercased, spaces/underscores converted to hyphens.
func GenerateSlug(ecosystemName, domainName, systemName, appName, workspaceName string) string {
	parts := []string{
		sanitizeName(ecosystemName),
		sanitizeName(domainName),
	}

	// Include system segment only if non-empty
	if systemName != "" {
		parts = append(parts, sanitizeName(systemName))
	}

	parts = append(parts, sanitizeName(appName), sanitizeName(workspaceName))

	return fmt.Sprintf("%s", strings.Join(parts, "-"))
}

func sanitizeName(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	return s
}
