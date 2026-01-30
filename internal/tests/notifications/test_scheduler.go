package main

import (
	"log"
	"time"
	"unipilot/internal/services/notifications"
)

func main() {
	scheduler, err := notifications.NewScheduler()
	if err != nil {
		log.Fatalf("Failed to create scheduler: %v", err)
	}

	log.Println("=== Testing Cron Scheduler ===")

	// Test 1: Check if we can get course entries
	log.Println("\n[Test 1] Getting course entries...")
	if err := scheduler.GetCourseEntries(); err != nil {
		log.Printf("Error: %v", err)
	}

	// Test 2: Start the scheduler
	log.Println("\n[Test 2] Starting scheduler...")
	if err := scheduler.StartScheduler(); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}
	log.Println("Scheduler started successfully")

	// Test 3: List all cron entries
	log.Println("\n[Test 3] Listing cron entries...")
	entries, err := scheduler.ListCronEntries()
	if err != nil {
		log.Fatalf("Error listing entries: %v", err)
	}

	log.Printf("Total cron entries: %d\n", len(entries))
	for i, entry := range entries {
		log.Printf("Entry %d:", i+1)
		log.Printf("  - ID: %d", entry.ID)
		log.Printf("  - Next run: %v", entry.Next)
		log.Printf("  - Previous run: %v", entry.Prev)
		log.Printf("  - Time until next: %v", time.Until(entry.Next))
	}

	// Test 4: Wait and observe
	log.Println("\n[Test 4] Waiting for cron jobs to execute...")
	log.Println("Will wait for 2 minutes. Watch for notifications...")

	ticker := time.NewTicker(10 * time.Second)
	timeout := time.After(2 * time.Minute)

	for {
		select {
		case <-ticker.C:
			log.Printf("Still running... Current time: %v", time.Now().Format("15:04:05"))

			// Re-check entries to see updates
			entries, _ := scheduler.ListCronEntries()
			for i, entry := range entries {
				log.Printf("  Entry %d next: %v (in %v)",
					i+1,
					entry.Next.Format("15:04:05"),
					time.Until(entry.Next).Round(time.Second))
			}

		case <-timeout:
			log.Println("\n=== Test Complete ===")
			scheduler.StopScheduler()
			return
		}
	}
}
