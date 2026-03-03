package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidateRegistryURL_ValidHTTP tests valid HTTP localhost URLs
func TestValidateRegistryURL_ValidHTTP(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "http localhost with port",
			url:  "http://localhost:5000",
		},
		{
			name: "http 127.0.0.1 with port",
			url:  "http://127.0.0.1:5000",
		},
		{
			name: "http localhost without port",
			url:  "http://localhost",
		},
		{
			name: "http host.docker.internal",
			url:  "http://host.docker.internal:5000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegistryURL(tt.url)
			assert.NoError(t, err, "Valid HTTP URL should pass validation")
		})
	}
}

// TestValidateRegistryURL_ValidHTTPS tests valid HTTPS localhost URLs
func TestValidateRegistryURL_ValidHTTPS(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "https localhost with port",
			url:  "https://localhost:5000",
		},
		{
			name: "https 127.0.0.1 with port",
			url:  "https://127.0.0.1:5000",
		},
		{
			name: "https host.docker.internal",
			url:  "https://host.docker.internal:5000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegistryURL(tt.url)
			assert.NoError(t, err, "Valid HTTPS URL should pass validation")
		})
	}
}

// TestValidateRegistryURL_InvalidScheme tests rejection of non-http/https schemes
func TestValidateRegistryURL_InvalidScheme(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		errMsg string
	}{
		{
			name:   "ftp scheme",
			url:    "ftp://localhost:5000",
			errMsg: "scheme",
		},
		{
			name:   "file scheme",
			url:    "file:///var/lib/registry",
			errMsg: "scheme",
		},
		{
			name:   "ssh scheme",
			url:    "ssh://localhost:5000",
			errMsg: "scheme",
		},
		{
			name:   "no scheme",
			url:    "localhost:5000",
			errMsg: "scheme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegistryURL(tt.url)
			assert.Error(t, err, "Invalid scheme should fail validation")
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

// TestValidateRegistryURL_EmbeddedCredentials tests rejection of URLs with credentials
func TestValidateRegistryURL_EmbeddedCredentials(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		errMsg string
	}{
		{
			name:   "username and password",
			url:    "http://user:pass@localhost:5000",
			errMsg: "credentials",
		},
		{
			name:   "username only",
			url:    "http://user@localhost:5000",
			errMsg: "credentials",
		},
		{
			name:   "password only (unusual but should reject)",
			url:    "http://:pass@localhost:5000",
			errMsg: "credentials",
		},
		{
			name:   "credentials in https",
			url:    "https://admin:secret@localhost:5000",
			errMsg: "credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegistryURL(tt.url)
			assert.Error(t, err, "Embedded credentials should fail validation")
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

// TestValidateRegistryURL_ExternalWarning tests warning for external URLs
func TestValidateRegistryURL_ExternalWarning(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		shouldWarn bool
	}{
		{
			name:       "external domain",
			url:        "http://registry.example.com:5000",
			shouldWarn: true,
		},
		{
			name:       "external IP",
			url:        "http://192.168.1.100:5000",
			shouldWarn: true,
		},
		{
			name:       "public domain",
			url:        "https://registry.hub.docker.com",
			shouldWarn: true,
		},
		{
			name:       "localhost should not warn",
			url:        "http://localhost:5000",
			shouldWarn: false,
		},
		{
			name:       "127.0.0.1 should not warn",
			url:        "http://127.0.0.1:5000",
			shouldWarn: false,
		},
		{
			name:       "host.docker.internal should not warn",
			url:        "http://host.docker.internal:5000",
			shouldWarn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warning := ValidateRegistryURLWithWarning(tt.url)

			if tt.shouldWarn {
				assert.NotEmpty(t, warning, "External URL should produce warning")
				assert.Contains(t, warning, "external", "Warning should mention external")
			} else {
				assert.Empty(t, warning, "Local URL should not produce warning")
			}
		})
	}
}

// TestValidateRegistryURL_HostDockerInternal tests host.docker.internal is treated as local
func TestValidateRegistryURL_HostDockerInternal(t *testing.T) {
	url := "http://host.docker.internal:5000"

	err := ValidateRegistryURL(url)
	assert.NoError(t, err, "host.docker.internal should be valid")

	warning := ValidateRegistryURLWithWarning(url)
	assert.Empty(t, warning, "host.docker.internal should not warn (it's local)")
}

// TestValidateRegistryURL_EmptyURL tests empty URL handling
func TestValidateRegistryURL_EmptyURL(t *testing.T) {
	err := ValidateRegistryURL("")
	assert.Error(t, err, "Empty URL should fail validation")
	assert.Contains(t, err.Error(), "empty")
}

// TestValidateRegistryURL_MalformedURL tests malformed URL handling
func TestValidateRegistryURL_MalformedURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "invalid characters",
			url:  "http://local host:5000",
		},
		{
			name: "missing host",
			url:  "http://:5000",
		},
		{
			name: "invalid port",
			url:  "http://localhost:99999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegistryURL(tt.url)
			assert.Error(t, err, "Malformed URL should fail validation")
		})
	}
}

// TestIsLocalHost tests localhost detection
func TestIsLocalHost(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		isLocal bool
	}{
		{
			name:    "localhost",
			host:    "localhost",
			isLocal: true,
		},
		{
			name:    "127.0.0.1",
			host:    "127.0.0.1",
			isLocal: true,
		},
		{
			name:    "0.0.0.0",
			host:    "0.0.0.0",
			isLocal: true,
		},
		{
			name:    "::1 (IPv6 localhost)",
			host:    "::1",
			isLocal: true,
		},
		{
			name:    "host.docker.internal",
			host:    "host.docker.internal",
			isLocal: true,
		},
		{
			name:    "external domain",
			host:    "registry.example.com",
			isLocal: false,
		},
		{
			name:    "external IP",
			host:    "192.168.1.100",
			isLocal: false,
		},
		{
			name:    "localhost with port (should strip port)",
			host:    "localhost:5000",
			isLocal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsLocalHost(tt.host)
			assert.Equal(t, tt.isLocal, result, "IsLocalHost(%q) should return %v", tt.host, tt.isLocal)
		})
	}
}

// TestValidateRegistryURL_CaseSensitivity tests case insensitivity
func TestValidateRegistryURL_CaseSensitivity(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "HTTP uppercase",
			url:  "HTTP://localhost:5000",
		},
		{
			name: "HTTPS uppercase",
			url:  "HTTPS://localhost:5000",
		},
		{
			name: "localhost uppercase",
			url:  "http://LOCALHOST:5000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegistryURL(tt.url)
			assert.NoError(t, err, "URLs should be case-insensitive")
		})
	}
}

// TestValidateRegistryURL_IPv6 tests IPv6 address handling
func TestValidateRegistryURL_IPv6(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "IPv6 localhost",
			url:     "http://[::1]:5000",
			wantErr: false,
		},
		{
			name:    "IPv6 link-local",
			url:     "http://[fe80::1]:5000",
			wantErr: false, // Should allow but warn
		},
		{
			name:    "IPv6 without brackets (invalid)",
			url:     "http://::1:5000",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegistryURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateRegistryURL_PortRange tests valid port range
func TestValidateRegistryURL_PortRange(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "port 80 (privileged but valid)",
			url:     "http://localhost:80",
			wantErr: false, // Allow even if privileged
		},
		{
			name:    "port 1024 (first unprivileged)",
			url:     "http://localhost:1024",
			wantErr: false,
		},
		{
			name:    "port 65535 (max valid)",
			url:     "http://localhost:65535",
			wantErr: false,
		},
		{
			name:    "port 0 (invalid)",
			url:     "http://localhost:0",
			wantErr: true,
		},
		{
			name:    "port 65536 (out of range)",
			url:     "http://localhost:65536",
			wantErr: true,
		},
		{
			name:    "negative port",
			url:     "http://localhost:-1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegistryURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err, "Invalid port should fail validation")
			} else {
				assert.NoError(t, err, "Valid port should pass validation")
			}
		})
	}
}
