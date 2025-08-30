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
	"unipilot/internal/models/user"
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
	user      user.User

	// Performance optimizations
	courseCache       []course.LocalCourse
	cacheLastUpdate   time.Time
	cacheMutex        sync.RWMutex
	notificationCache map[string]time.Time // Track sent notifications to avoid duplicates
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
		cron:              cron.New(cron.WithSeconds()),
		db:                db,
		ctx:               ctx,
		cancel:            cancel,
		isRunning:         false,
		notificationCache: make(map[string]time.Time),
	}

	return scheduler, nil
}

// InitializeForDaemon sets up the scheduler to run as a daemon
func (s *Scheduler) InitializeForDaemon(user *user.User) error {
	// Set the user ID for this daemon instance
	s.user = *user

	// Initialize any daemon-specific configurations
	log.Printf("[Scheduler] Initialized daemon for user %d", user.ID)
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
	eveningEntry, err := s.cron.AddFunc("0 0 20 * * *", func() {
		log.Printf("[Scheduler] Starting evening notifications at %s", time.Now().Format("15:04:05"))
		s.SendDailyNotifications("evening")
	})
	if err != nil {
		return fmt.Errorf("failed to schedule evening notifications: %w", err)
	}

	// Schedule course entry notifications every 5 minutes instead of every minute
	// This reduces database load by 80%
	courseEntry, err := s.cron.AddFunc("@every 5m", func() {
		log.Printf("[Scheduler] Starting course entry notifications at %s", time.Now().Format("15:04:05"))
		s.SendCourseEntryNotifications()
	})
	if err != nil {
		return fmt.Errorf("failed to schedule course entry notifications: %w", err)
	}

	// Schedule cache refresh every hour to keep course data fresh
	cacheRefresh, err := s.cron.AddFunc("@every 1h", func() {
		log.Printf("[Scheduler] Refreshing course cache at %s", time.Now().Format("15:04:05"))
		s.refreshCourseCache()
	})
	if err != nil {
		return fmt.Errorf("failed to schedule cache refresh: %w", err)
	}

	s.cron.Start()
	s.isRunning = true

	log.Printf("[Scheduler] Started with morning entry ID: %d, evening entry ID: %d, course entry ID: %d, cache refresh ID: %d",
		morningEntry, eveningEntry, courseEntry, cacheRefresh)
	log.Printf("[Scheduler] Next morning notification: %s", s.cron.Entry(morningEntry).Next.Format("2006-01-02 15:04:05"))
	log.Printf("[Scheduler] Next evening notification: %s", s.cron.Entry(eveningEntry).Next.Format("2006-01-02 15:04:05"))
	log.Printf("[Scheduler] Next course entry notification: %s", s.cron.Entry(courseEntry).Next.Format("2006-01-02 15:04:05"))
	log.Printf("[Scheduler] Next cache refresh: %s", s.cron.Entry(cacheRefresh).Next.Format("2006-01-02 15:04:05"))

	return nil
}

// refreshCourseCache updates the cached course data
func (s *Scheduler) refreshCourseCache() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	var courses []course.LocalCourse
	if err := s.db.Where("semester = ?", s.user.Semester).Find(&courses).Error; err != nil {
		log.Printf("[Scheduler] Error refreshing course cache: %v", err)
		return
	}

	s.courseCache = courses
	s.cacheLastUpdate = time.Now()
	log.Printf("[Scheduler] Refreshed course cache with %d courses", len(courses))
}

// getCachedCourses returns courses from cache, refreshing if needed
func (s *Scheduler) getCachedCourses() ([]course.LocalCourse, error) {
	s.cacheMutex.RLock()

	// If cache is empty or older than 1 hour, refresh it
	if len(s.courseCache) == 0 || time.Since(s.cacheLastUpdate) > time.Hour {
		s.cacheMutex.RUnlock()
		s.refreshCourseCache()
		s.cacheMutex.RLock()
	}

	courses := make([]course.LocalCourse, len(s.courseCache))
	copy(courses, s.courseCache)
	s.cacheMutex.RUnlock()

	return courses, nil
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

	log.Printf("[Scheduler] Retrieved %d assignments for user %d", len(assignments), s.user.ID)
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
		SenderID:  s.user.ID,
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

	// Get courses from cache instead of database
	courses, err := s.getCachedCourses()
	if err != nil {
		log.Printf("[Scheduler] Error getting courses: %v", err)
		return
	}

	// Only process courses that have classes today
	today := time.Now().Weekday()
	relevantCourses := s.filterCoursesForToday(courses, today)

	if len(relevantCourses) == 0 {
		log.Printf("[Scheduler] No courses scheduled for today (%s)", today.String())
		return
	}

	log.Printf("[Scheduler] Processing %d courses scheduled for today", len(relevantCourses))

	for _, course := range relevantCourses {
		s.processCourseForNotification(course)
		time.Sleep(1 * time.Second) // Reduced delay since we're processing fewer courses
	}
}

// filterCoursesForToday returns only courses that have classes on the given day
func (s *Scheduler) filterCoursesForToday(courses []course.LocalCourse, day time.Weekday) []course.LocalCourse {
	var relevant []course.LocalCourse

	for _, course := range courses {
		schedule, err := course.ParseSchedule(course.Schedule)
		if err != nil {
			continue // Skip courses with invalid schedules
		}

		if slices.Contains(schedule.Days, int(day)) {
			relevant = append(relevant, course)
		}
	}

	return relevant
}

// processCourseForNotification checks if a course needs a notification and sends it
func (s *Scheduler) processCourseForNotification(course course.LocalCourse) {
	now := time.Now()
	in30Minutes := now.Add(30 * time.Minute)

	schedule, err := course.ParseSchedule(course.Schedule)
	if err != nil {
		log.Printf("[Scheduler] Error parsing schedule for course %s: %v", course.Code, err)
		return
	}

	startTime := time.Date(now.Year(), now.Month(), now.Day(), schedule.StartTime, schedule.StartMinute, 0, 0, now.Location())
	endTime := time.Date(now.Year(), now.Month(), now.Day(), schedule.EndTime, schedule.EndMinute, 0, 0, now.Location())

	courseEntry := CourseEntry{
		course:    course,
		schedule:  schedule,
		startTime: startTime,
		endTime:   endTime,
	}

	// If the class is currently in session, skip
	if duringClass(now, courseEntry) {
		return
	}

	// If the class is scheduled to start in the next 30 minutes, send notification
	if duringClass(in30Minutes, courseEntry) {
		// Check if we already sent a notification for this course today
		cacheKey := fmt.Sprintf("%s_%s", course.Code, now.Format("2006-01-02"))
		if lastSent, exists := s.notificationCache[cacheKey]; exists &&
			time.Since(lastSent) < 30*time.Minute {
			return // Already sent notification recently
		}

		s.createCourseNotification(courseEntry)
		s.notificationCache[cacheKey] = now
	}
}

type CourseEntry struct {
	course    course.LocalCourse
	schedule  *course.ParsedSchedule
	startTime time.Time
	endTime   time.Time
}

func duringClass(t time.Time, courseEntry CourseEntry) bool {
	return courseEntry.startTime.Before(t) && courseEntry.endTime.After(t)
}

func (s *Scheduler) createCourseNotification(courseEntry CourseEntry) {
	title := fmt.Sprintf("%s - %s", courseEntry.course.Code, courseEntry.course.Name)
	message := fmt.Sprintf("You have a class at %v in 30 minutes", courseEntry.startTime.Format("15:04"))

	// Create notification in database
	notification := models.LocalNotification{
		Type:      models.NotificationCourse,
		Title:     title,
		Message:   message,
		SenderID:  s.user.ID,
		Read:      false,
		ExpiresAt: &courseEntry.endTime,
	}

	// Check if notification already exists (avoid duplicates)
	if err := s.db.Where("title = ? AND message = ? AND sender_id = ? AND expires_at > ?",
		title, message, s.user.ID, time.Now()).First(&notification).Error; err == nil {
		log.Printf("[Scheduler] Notification already exists: %s", title)
		return
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
