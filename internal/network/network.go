package network

import (
	"fmt"
	"net/http"
	"time"
	"unipilot/internal/secrets"
)

var onlineStatus bool
var lastChecked time.Time

func IsOnline() bool {
	// Cache status for 30 seconds to avoid repeated checks
	if time.Since(lastChecked) < 5*time.Second {
		return onlineStatus
	}

	api_url := secrets.CONSTANTS["API_URL"]

	// Simple check - adjust as needed
	client := http.Client{Timeout: 1 * time.Second}
	_, err := client.Get(fmt.Sprintf("%s/health", api_url))

	onlineStatus = err == nil
	lastChecked = time.Now()
	return onlineStatus
}
