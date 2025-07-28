package network

import (
        "time"
        "net/http"
)

var onlineStatus bool
var lastChecked time.Time

func IsOnline() bool {
    // Cache status for 30 seconds to avoid repeated checks
    if time.Since(lastChecked) < 30*time.Second {
        return onlineStatus
    }

    // Simple check - adjust as needed
    client := http.Client{Timeout: 3 * time.Second}
    _, err := client.Get("https://newsroom.dedyn.io")
    
    onlineStatus = err == nil
    lastChecked = time.Now()
    return onlineStatus
} 
