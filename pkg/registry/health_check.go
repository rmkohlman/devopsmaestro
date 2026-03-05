package registry

import (
	"net/http"
	"time"
)

// healthCheckClient is the shared HTTP client for health-check requests in
// waitForReady methods.  It has a short timeout to prevent hung requests from
// blocking service startup indefinitely.
var healthCheckClient = &http.Client{Timeout: 2 * time.Second}
