package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// CacheReadiness.FormatSummary — emoji and content tests
// =============================================================================

func TestFormatSummary_NoRegistriesEnabled(t *testing.T) {
	r := &CacheReadiness{TotalEnabled: 0}
	assert.Equal(t, "Cache: no registries enabled", r.FormatSummary())
}

func TestFormatSummary_AllHealthy_HasCheckmarkEmoji(t *testing.T) {
	r := &CacheReadiness{
		TotalEnabled: 3,
		HealthyCount: 3,
		AllHealthy:   true,
	}
	summary := r.FormatSummary()
	assert.Contains(t, summary, "✅", "healthy summary must contain checkmark emoji")
	assert.Contains(t, summary, "3/3")
}

func TestFormatSummary_SomeUnhealthy_HasWarningEmoji(t *testing.T) {
	r := &CacheReadiness{
		TotalEnabled: 3,
		HealthyCount: 2,
		AllHealthy:   false,
		Unhealthy:    []string{"squid-proxy"},
	}
	summary := r.FormatSummary()
	assert.Contains(t, summary, "⚠️", "degraded summary must contain warning emoji")
	assert.Contains(t, summary, "2/3")
	assert.Contains(t, summary, "squid-proxy")
}

func TestFormatSummary_AllUnhealthy_HasWarningEmoji(t *testing.T) {
	r := &CacheReadiness{
		TotalEnabled: 2,
		HealthyCount: 0,
		AllHealthy:   false,
		Unhealthy:    []string{"squid-proxy", "pip-proxy"},
	}
	summary := r.FormatSummary()
	assert.Contains(t, summary, "⚠️")
	assert.Contains(t, summary, "0/2")
	assert.Contains(t, summary, "squid-proxy")
	assert.Contains(t, summary, "pip-proxy")
}

func TestFormatSummary_AllHealthy_NoWarningEmoji(t *testing.T) {
	r := &CacheReadiness{
		TotalEnabled: 1,
		HealthyCount: 1,
		AllHealthy:   true,
	}
	summary := r.FormatSummary()
	assert.NotContains(t, summary, "⚠️", "healthy summary must not contain warning emoji")
}

func TestFormatSummary_Unhealthy_NoCheckmarkEmoji(t *testing.T) {
	r := &CacheReadiness{
		TotalEnabled: 1,
		HealthyCount: 0,
		AllHealthy:   false,
		Unhealthy:    []string{"squid-proxy"},
	}
	summary := r.FormatSummary()
	assert.NotContains(t, summary, "✅", "degraded summary must not contain checkmark emoji")
}
