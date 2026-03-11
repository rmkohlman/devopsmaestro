package workspace

import (
	"fmt"
	"strings"
)

// GenerateSlug creates a workspace slug from hierarchy names.
// Format: {ecosystem}-{domain}-{app}-{workspace}
// Names are sanitized: lowercased, spaces/underscores converted to hyphens.
func GenerateSlug(ecosystemName, domainName, appName, workspaceName string) string {
	return fmt.Sprintf("%s-%s-%s-%s",
		sanitizeName(ecosystemName),
		sanitizeName(domainName),
		sanitizeName(appName),
		sanitizeName(workspaceName),
	)
}

func sanitizeName(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	return s
}
