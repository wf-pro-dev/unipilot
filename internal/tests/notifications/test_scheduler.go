package main

import (
	"log"
	"unipilot/internal/services/notifications"
)

func main() {
	scheduler, err := notifications.NewScheduler()
	if err != nil {
		log.Fatalf("Failed to create scheduler: %v", err)
	}

	if err := scheduler.GetCourseEntries(); err != nil {
		log.Fatalf("Failed to get course entries: %v", err)
	}
}
