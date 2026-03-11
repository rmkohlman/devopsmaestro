package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Wave 1A: GenerateSlug Tests (Sprint 4.3)
// Tests for workspace slug generation from hierarchy names.
// Format: {ecosystem}-{domain}-{app}-{workspace}
// =============================================================================

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name          string
		ecosystemName string
		domainName    string
		appName       string
		workspaceName string
		want          string
	}{
		{
			name:          "all lowercase names produces clean slug",
			ecosystemName: "myeco",
			domainName:    "mydomain",
			appName:       "myapp",
			workspaceName: "dev",
			want:          "myeco-mydomain-myapp-dev",
		},
		{
			name:          "mixed case names are lowercased",
			ecosystemName: "MyEco",
			domainName:    "MyDomain",
			appName:       "MyApp",
			workspaceName: "Dev",
			want:          "myeco-mydomain-myapp-dev",
		},
		{
			name:          "all uppercase names are lowercased",
			ecosystemName: "CORP",
			domainName:    "PAYMENTS",
			appName:       "GATEWAY",
			workspaceName: "STAGING",
			want:          "corp-payments-gateway-staging",
		},
		{
			name:          "spaces in names become hyphens",
			ecosystemName: "my eco",
			domainName:    "my domain",
			appName:       "my app",
			workspaceName: "my ws",
			want:          "my-eco-my-domain-my-app-my-ws",
		},
		{
			name:          "underscores in names become hyphens",
			ecosystemName: "my_eco",
			domainName:    "my_domain",
			appName:       "my_app",
			workspaceName: "my_ws",
			want:          "my-eco-my-domain-my-app-my-ws",
		},
		{
			name:          "mixed spaces underscores and uppercase are all sanitized",
			ecosystemName: "My_Eco System",
			domainName:    "My Domain_Area",
			appName:       "My_App Service",
			workspaceName: "Dev_Workspace One",
			want:          "my-eco-system-my-domain-area-my-app-service-dev-workspace-one",
		},
		{
			name:          "all empty strings produces triple-hyphen separator",
			ecosystemName: "",
			domainName:    "",
			appName:       "",
			workspaceName: "",
			want:          "---",
		},
		{
			name:          "numeric names are preserved as-is",
			ecosystemName: "eco1",
			domainName:    "domain2",
			appName:       "app3",
			workspaceName: "ws4",
			want:          "eco1-domain2-app3-ws4",
		},
		{
			name:          "already hyphenated names are preserved",
			ecosystemName: "my-eco",
			domainName:    "my-domain",
			appName:       "my-app",
			workspaceName: "dev-ws",
			want:          "my-eco-my-domain-my-app-dev-ws",
		},
		{
			name:          "single character names are valid",
			ecosystemName: "e",
			domainName:    "d",
			appName:       "a",
			workspaceName: "w",
			want:          "e-d-a-w",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSlug(tt.ecosystemName, tt.domainName, tt.appName, tt.workspaceName)
			assert.Equal(t, tt.want, got)
		})
	}
}
