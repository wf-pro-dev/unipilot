package notifications

import (
	"fmt"
	"log"
	"math"
	"os/exec"
	"time"

	"unipilot/internal/models/assignment"
	"unipilot/internal/storage"
)

func getUntilDeadline(deadline time.Time) string {

	today, err := time.Parse(time.DateOnly, time.Now().Format(time.DateOnly))

	if err != nil {
		return ""
	}

	if deadline.Equal(today) {
		return "Today"
	}

	if deadline.Equal(today.AddDate(0, 0, 1)) {
		return "Tomorrow"
	}

	if deadline.Equal(today.AddDate(0, 0, 7)) {
		return "in a week"
	}

	daysRemaining := int(math.Ceil(deadline.Sub(today).Hours() / 24))
	return fmt.Sprintf("in %d days", daysRemaining)
}

func UseNotifier(args []string) error {
	// Find terminal-notifier in common locations
	locations := []string{
		"/usr/local/bin/terminal-notifier",                         // Homebrew default
		"/opt/homebrew/bin/terminal-notifier",                      // Apple Silicon Homebrew
		"./terminal-notifier.app/Contents/MacOS/terminal-notifier", // Local copy
	}

	var cmd *exec.Cmd
	var err error
	// Try different locations until we find the binary
	for _, path := range locations {

		cmd = exec.Command(path, args...)
		err = cmd.Run()
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to send notification: %w", err)
}

// sendNotification sends a clickable notification that opens a URL when clicked
func sendNotification(assign map[string]string) error {
	title := fmt.Sprintf("%s: %s", assign["course_code"], assign["title"])

	// Parse deadline and calculate days remaining
	deadline, err := time.Parse(time.DateOnly, assign["deadline"][:10])
	if err != nil {
		return fmt.Errorf("failed to parse deadline: %w", err)
	}

	subtitle := fmt.Sprintf("Due %s (%s)",
		deadline.Format("Jan 2, 2006"),
		getUntilDeadline(deadline))

	// Build the terminal-notifier command
	args := []string{
		"-group", assign["notion_id"],
		"-title", title,
		"-subtitle", subtitle,
		"-message", assign["todo"],
		"-sound", "Frog",
		"-timeout", "10", // Notification stays for 30 seconds
	}

	// Add click action if URL exists
	if link, exists := assign["link"]; exists && link != "" {
		args = append(args, "-open", link)
	}

	// Remove the notification if it already exists
	err = UseNotifier([]string{"-remove", assign["notion_id"]})
	if err != nil {
		return fmt.Errorf("failed to remove notification: %w", err)
	}

	// Send notification if the assignment is not done

	err = UseNotifier(args)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

// scheduleNotifications checks for upcoming assignments and notifies
func ScheduleNotifications() error {

	db, _, err := storage.GetLocalDB()
	if err != nil {
		return err
	}

	now := time.Now()

	query := fmt.Sprintf(

		"deadline BETWEEN '%s' AND '%s' AND status_name != 'Done'",
		now.AddDate(0, 0, -1).Format(time.RFC3339),
		now.AddDate(0, 0, 7).Format(time.RFC3339),
	)

	assignments := []assignment.LocalAssignment{}
	db.Model(&assignment.LocalAssignment{}).Where(query).Order("deadline ASC").Find(&assignments)

	for _, assign := range assignments {
		if err := sendNotification(assign.ToMap()); err != nil {
			fmt.Printf("Error notifying for assignment %s: %v\n", assign.Title, err)
		}
		time.Sleep(5 * time.Second) // Space out notifications
	}

	return nil
}

func main() {
	err := ScheduleNotifications()
	if err != nil {
		log.Fatal(err)
	}
}
