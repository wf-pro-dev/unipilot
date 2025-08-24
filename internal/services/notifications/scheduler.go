package notifications

import (
	"context"
	"fmt"
	"log"
	"slices"
	"sync"
	"time"

	"unipilot/internal/models"
	"unipilot/internal/models/assignment"
	"unipilot/internal/models/course"
	"unipilot/internal/storage"

	"github.com/gen2brain/beeep"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// Scheduler handles scheduled notification tasks
type Scheduler struct {
	cron      *cron.Cron
	db        *gorm.DB
	ctx       context.Context
	cancel    context.CancelFunc
	isRunning bool
	mu        sync.RWMutex
	userID    uint
}

// NewScheduler creates a new notification scheduler
func NewScheduler() (*Scheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Get database connection
	db, _, err := storage.GetLocalDB()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	scheduler := &Scheduler{
		cron:      cron.New(cron.WithSeconds()),
		db:        db,
		ctx:       ctx,
		cancel:    cancel,
		isRunning: false,
	}

	return scheduler, nil
}

// InitializeForDaemon sets up the scheduler to run as a daemon
func (s *Scheduler) InitializeForDaemon(userID uint) error {
	// Set the user ID for this daemon instance
	s.userID = userID

	// Initialize any daemon-specific configurations
	log.Printf("[Scheduler] Initialized daemon for user %d", userID)
	return nil
}

// StartScheduler initializes and starts the cron scheduler
func (s *Scheduler) StartScheduler() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("scheduler is already running")
	}

	// Schedule morning notifications at 8:00 AM
	morningEntry, err := s.cron.AddFunc("0 0 8 * * *", func() {
		log.Printf("[Scheduler] Starting morning notifications at %s", time.Now().Format("15:04:05"))
		s.SendDailyNotifications("morning")
	})
	if err != nil {
		return fmt.Errorf("failed to schedule morning notifications: %w", err)
	}

	// Schedule evening notifications at 8:00 PM
	eveningEntry, err := s.cron.AddFunc("0 18 10 * * *", func() {
		log.Printf("[Scheduler] Starting evening notifications at %s", time.Now().Format("15:04:05"))
		s.SendDailyNotifications("evening")
	})
	if err != nil {
		return fmt.Errorf("failed to schedule evening notifications: %w", err)
	}

	// Schedule course entry notifications every 5 minutes
	courseEntry, err := s.cron.AddFunc("@every 5m", func() {
		log.Printf("[Scheduler] Starting course entry notifications at %s", time.Now().Format("15:04:05"))
		s.SendCourseEntryNotifications()
	})
	if err != nil {
		return fmt.Errorf("failed to schedule course entry notifications: %w", err)
	}

	s.cron.Start()
	s.isRunning = true

	log.Printf("[Scheduler] Started with morning entry ID: %d, evening entry ID: %d, course entry ID: %d", morningEntry, eveningEntry, courseEntry)
	log.Printf("[Scheduler] Next morning notification: %s", s.cron.Entry(morningEntry).Next.Format("2006-01-02 15:04:05"))
	log.Printf("[Scheduler] Next evening notification: %s", s.cron.Entry(eveningEntry).Next.Format("2006-01-02 15:04:05"))

	return nil
}

// StopScheduler gracefully stops the scheduler and cleans up resources
func (s *Scheduler) StopScheduler() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	// Stop the cron scheduler
	s.cron.Stop()

	// Cancel the context to stop any running operations
	s.cancel()

	s.isRunning = false
	log.Printf("[Scheduler] Stopped successfully")
	return nil
}

// IsSchedulerRunning checks if the scheduler is currently active
func (s *Scheduler) IsSchedulerRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// SendDailyNotifications is the main entry point for daily notification processing
func (s *Scheduler) SendDailyNotifications(session string) {
	log.Printf("[Scheduler] Processing %s daily notifications", session)

	// Get weekly assignments
	assignments, err := s.getWeeklyAssignments()
	if err != nil {
		log.Printf("[Scheduler] Error getting weekly assignments: %v", err)
		return
	}

	if len(assignments) == 0 {
		log.Printf("[Scheduler] No assignments found for %s session", session)
		return
	}

	log.Printf("[Scheduler] Found %d assignments for %s session", len(assignments), session)

	// Group assignments by course
	grouped := s.groupAssignmentsByCourse(assignments)

	// Send course-based notifications
	s.sendCourseBasedNotifications(grouped, session)
}

// getWeeklyAssignments queries database for assignments due in the next 7 days
func (s *Scheduler) getWeeklyAssignments() ([]assignment.LocalAssignment, error) {
	now := time.Now()
	startDate := now.AddDate(0, 0, -1)
	endDate := now.AddDate(0, 0, 7)

	// Query for assignments due in the next 7 days that are not completed
	query := fmt.Sprintf(
		"deadline BETWEEN '%s' AND '%s' AND status_name != 'Done'",
		startDate.Format(time.RFC3339),
		endDate.Format(time.RFC3339),
	)

	var assignments []assignment.LocalAssignment
	err := s.db.Model(&assignment.LocalAssignment{}).Where(query).Order("deadline ASC").Find(&assignments).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query assignments: %w", err)
	}

	log.Printf("[Scheduler] Retrieved %d assignments for user %d", len(assignments), s.userID)
	return assignments, nil
}

// groupAssignmentsByCourse organizes assignments by course
func (s *Scheduler) groupAssignmentsByCourse(assignments []assignment.LocalAssignment) map[string][]assignment.LocalAssignment {
	grouped := make(map[string][]assignment.LocalAssignment)

	for _, assignment := range assignments {
		grouped[assignment.CourseCode] = append(grouped[assignment.CourseCode], assignment)
	}

	log.Printf("[Scheduler] Grouped assignments by course: %d courses", len(grouped))
	return grouped
}

// sendCourseBasedNotifications sends course-based notifications
func (s *Scheduler) sendCourseBasedNotifications(grouped map[string][]assignment.LocalAssignment, session string) {
	log.Printf("[Scheduler] Sending notifications for %d courses", len(grouped))

	for courseCode, assignments := range grouped {
		s.createAssignmentNotification(courseCode, assignments)
		time.Sleep(2 * time.Second) // Delay between notifications
	}
}

// createAssignmentNotification creates a notification for a course with multiple assignments
func (s *Scheduler) createAssignmentNotification(courseCode string, assignments []assignment.LocalAssignment) {
	title := s.formatTitle(courseCode)
	message := s.formatMessage(len(assignments))

	// Create notification in database
	notification := models.LocalNotification{
		Type:      models.NotificationAssignment,
		Title:     title,
		Message:   message,
		SenderID:  s.userID,
		Read:      false,
		ExpiresAt: &time.Time{},
	}

	if err := s.db.Create(&notification).Error; err != nil {
		log.Printf("[Scheduler] Error saving notification: %v", err)
		return
	}

	// Send system notification
	if err := beeep.Notify(title, message, ""); err != nil {
		log.Printf("[Scheduler] Error sending system notification: %v", err)
	} else {
		log.Printf("[Scheduler] Sent course notification: %s", title)
	}
}

// formatTitle creates course-based notification titles
func (s *Scheduler) formatTitle(courseCode string) string {
	return courseCode
}

// formatMessage creates detailed course notification messages
func (s *Scheduler) formatMessage(count int) string {
	return fmt.Sprintf("%d assignment(s) due this week", count)
}

// SendCourseEntryNotifications sends course entry notifications
func (s *Scheduler) SendCourseEntryNotifications() {
	log.Printf("[Scheduler] Processing course entry notifications")

	// Get courses
	courses, err := s.getCourseEntries()
	if err != nil {
		log.Printf("[Scheduler] Error getting courses: %v", err)
		return
	}

	for _, course := range courses {
		s.createCourseNotification(course)
		time.Sleep(2 * time.Second) // Delay between notifications
	}

}

type UntilType string

const (
	UntilHiigh UntilType = "an hour"
	UntilLow   UntilType = "10 minutes"
)

type CourseEntry struct {
	course course.LocalCourse
	until  UntilType
}

func (s *Scheduler) getCourseEntries() ([]CourseEntry, error) {
	now := time.Now()
	inHour := now.Add(time.Hour)
	in10Minutes := now.Add(10 * time.Minute)
	var results []CourseEntry
	var courses []course.LocalCourse
	err := s.db.Find(&courses).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query courses: %w", err)
	}

	for _, course := range courses {
		schedule, err := course.ParseSchedule(course.Schedule)
		if err != nil {
			log.Printf("[Scheduler] Error parsing schedule for course %s: %v", course.Code, err)
			continue
		}

		// If the class is currently in session, skip
		if duringClass(now, schedule.StartTime, schedule.StartMinute) {
			continue
		}

		// If the class is scheduled to start in the next hour, add a notification
		if slices.Contains(schedule.Days, int(now.Weekday())) &&
			duringClass(inHour, schedule.StartTime, schedule.StartMinute) {
			results = append(results, CourseEntry{course: course, until: UntilHiigh})
		}

		// If the class is scheduled to start in the next 10 minutes, add a notification
		if slices.Contains(schedule.Days, int(now.Weekday())) &&
			duringClass(in10Minutes, schedule.StartTime, schedule.StartMinute) {
			results = append(results, CourseEntry{course: course, until: UntilLow})
		}
	}

	return results, nil
}

func duringClass(time time.Time, hour int, minute int) bool {
	return time.Hour() >= hour && time.Hour() <= hour &&
		time.Minute() >= minute && time.Minute() <= minute
}

func (s *Scheduler) createCourseNotification(courseEntry CourseEntry) {
	title := s.formatTitle(courseEntry.course.Code)
	message := fmt.Sprintf("You have a class in %s", courseEntry.until)

	// Create notification in database
	notification := models.LocalNotification{
		Type:      models.NotificationCourse,
		Title:     title,
		Message:   message,
		SenderID:  s.userID,
		Read:      false,
		ExpiresAt: &time.Time{},
	}

	if err := s.db.Create(&notification).Error; err != nil {
		log.Printf("[Scheduler] Error saving notification: %v", err)
		return
	}

	// Send system notification
	if err := beeep.Notify(title, message, ""); err != nil {
		log.Printf("[Scheduler] Error sending system notification: %v", err)
	} else {
		log.Printf("[Scheduler] Sent course notification: %s", title)
	}
}
