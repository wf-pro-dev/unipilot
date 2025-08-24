package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"unipilot/internal/services/notifications"
)

func main() {
	var (
		userID  = flag.Uint("user", 0, "User ID to monitor")
		logFile = flag.String("log", "", "Log file path")
	)
	flag.Parse()

	if *userID == 0 {
		log.Fatal("User ID is required")
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

	if err := scheduler.InitializeForDaemon(*userID); err != nil {
		log.Fatalf("Failed to initialize scheduler: %v", err)
	}

	if err := scheduler.StartScheduler(); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}

	log.Printf("Notification daemon started for user %d", *userID)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down notification daemon...")
	scheduler.StopScheduler()
}
