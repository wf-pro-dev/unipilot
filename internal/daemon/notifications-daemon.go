package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"unipilot/internal/services/notifications"
	"unipilot/internal/services/utils"
	"unipilot/internal/sse"
)

func main() {
	var (
		userID  = flag.Uint("user", 0, "User ID to monitor")
		logFile = flag.String("log", "", "Log file path")
	)
	flag.Parse()

	user, err := utils.GetUserFromFile()
	if err != nil {
		log.Fatalf("Failed to get current user: %v", err)
	}

	// Setup logging if specified
	if *logFile != "" {
		file, err := os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		defer file.Close()
		log.SetOutput(file)
	}

	// Create and start scheduler
	scheduler, err := notifications.NewScheduler()
	if err != nil {
		log.Fatalf("Failed to create scheduler: %v", err)
	}

	if err := scheduler.InitializeForDaemon(user); err != nil {
		log.Fatalf("Failed to initialize scheduler: %v", err)
	}

	if err := scheduler.StartScheduler(); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}

	// Create and start event handler
	eventHandler, err := notifications.NewEventHandler(user.ID)
	if err != nil {
		log.Fatalf("Failed to create event handler: %v", err)
	}

	// Try to get authenticated HTTP client from stored credentials
	httpClient, err := sse.NewSSEClient()
	if err != nil {
		log.Printf("Warning: Could not create authenticated HTTP client: %v", err)
		log.Printf("Event handler will not receive real-time notifications")
	} else {
		// Create SSE client for the daemon with authentication
		sseClient := sse.NewSSE()

		// Start SSE connection in background with authenticated client
		go func() {
			log.Printf("[Daemon] Starting authenticated SSE connection for event handler")
			sseClient.Connect(httpClient)
		}()

		// Start event handler with the SSE client
		if err := eventHandler.StartEventHandler(sseClient); err != nil {
			log.Printf("Warning: Failed to start event handler: %v", err)
		} else {
			log.Printf("Event handler started for user %d", *userID)
		}
	}

	log.Printf("Notification daemon started for user %d", *userID)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down notification daemon...")

	// Stop both scheduler and event handler
	scheduler.StopScheduler()
	eventHandler.StopEventHandler()
}
