package envinjector

import (
	"devopsmaestro/models"
	"fmt"
	"strings"
)

// EnvironmentInjector generates environment variables for registry injection
type EnvironmentInjector struct{}

// NewEnvironmentInjector creates a new EnvironmentInjector
func NewEnvironmentInjector() *EnvironmentInjector {
	return &EnvironmentInjector{}
}

// InjectForAttach generates env vars for attaching to a workspace
func (ei *EnvironmentInjector) InjectForAttach(registry *models.Registry) map[string]string {
	if registry == nil {
		return map[string]string{}
	}
	return ei.InjectForAttachWithHost(registry, "localhost")
}

// InjectForBuild generates env vars for Docker build-time
// Uses host.docker.internal instead of localhost for Docker builds
func (ei *EnvironmentInjector) InjectForBuild(registry *models.Registry) map[string]string {
	if registry == nil {
		return map[string]string{}
	}
	return ei.InjectForAttachWithHost(registry, "host.docker.internal")
}

// InjectForAttachWithHost generates env vars with a custom host
func (ei *EnvironmentInjector) InjectForAttachWithHost(registry *models.Registry, host string) map[string]string {
	if registry == nil {
		return map[string]string{}
	}

	envVars := make(map[string]string)

	switch registry.Type {
	case "devpi":
		// PyPI registry
		url := fmt.Sprintf("http://%s:%d/root/pypi/+simple/", host, registry.Port)
		envVars["PIP_INDEX_URL"] = url

		// Only set PIP_TRUSTED_HOST for localhost (security requirement)
		if isLocalHost(host) {
			envVars["PIP_TRUSTED_HOST"] = host
		}

	case "verdaccio":
		// NPM registry
		url := fmt.Sprintf("http://%s:%d/", host, registry.Port)
		envVars["NPM_CONFIG_REGISTRY"] = url
		envVars["npm_config_registry"] = url // lowercase variant for compatibility

	case "athens":
		// Go module proxy
		url := fmt.Sprintf("http://%s:%d", host, registry.Port)
		envVars["GOPROXY"] = url
		envVars["GONOSUMDB"] = "*" // Disable checksum verification for privacy
		envVars["GOPRIVATE"] = "*" // Treat all modules as private

	case "squid":
		// HTTP proxy
		url := fmt.Sprintf("http://%s:%d", host, registry.Port)
		envVars["HTTP_PROXY"] = url
		envVars["HTTPS_PROXY"] = url
		envVars["http_proxy"] = url // lowercase variants
		envVars["https_proxy"] = url

		// NO_PROXY list - exclude local addresses and RFC1918 private ranges
		noProxy := []string{"localhost", "127.0.0.1", "::1", ".local", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}
		envVars["NO_PROXY"] = strings.Join(noProxy, ",")
		envVars["no_proxy"] = strings.Join(noProxy, ",")

	case "zot":
		// OCI registry - provide endpoint for convenience
		envVars["DVM_OCI_REGISTRY"] = fmt.Sprintf("%s:%d", host, registry.Port)

	default:
		// Unknown registry type - return empty map
	}

	return envVars
}

// InjectForAttachMultiple generates env vars for multiple registries
func (ei *EnvironmentInjector) InjectForAttachMultiple(registries []*models.Registry) map[string]string {
	envVars := make(map[string]string)

	for _, registry := range registries {
		if registry == nil {
			continue
		}
		// Merge env vars from each registry
		for k, v := range ei.InjectForAttach(registry) {
			envVars[k] = v
		}
	}

	return envVars
}

// InjectForDocker generates Docker-specific env vars
// Uses host.docker.internal for all registries
func (ei *EnvironmentInjector) InjectForDocker(registries []*models.Registry) map[string]string {
	envVars := make(map[string]string)

	for _, registry := range registries {
		if registry == nil {
			continue
		}
		// Use host.docker.internal for Docker context
		for k, v := range ei.InjectForAttachWithHost(registry, "host.docker.internal") {
			envVars[k] = v
		}
	}

	return envVars
}

// isLocalHost checks if a host is localhost (for PIP_TRUSTED_HOST security)
func isLocalHost(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}
