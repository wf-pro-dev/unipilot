package notifications

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/services/utils"

	"github.com/gen2brain/beeep"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// Scheduler handles scheduled notification tasks
type Scheduler struct {
	cron          *cron.Cron
	db            *gorm.DB
	ctx           context.Context
	cancel        context.CancelFunc
	isRunning     bool
	mu            sync.RWMutex
	courseEntries map[uint]*CourseEntry
}

// NewScheduler creates a new notification scheduler
func NewScheduler() (*Scheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Get database connection
	db, err := utils.GetUserDB()
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, errors.DBConnectionFailed, "Failed to get user database")
	}

	scheduler := &Scheduler{
		cron:          cron.New(cron.WithSeconds()),
		db:            db,
		ctx:           ctx,
		cancel:        cancel,
		isRunning:     false,
		courseEntries: make(map[uint]*CourseEntry),
	}

	return scheduler, nil
}

// StartScheduler initializes and starts the cron scheduler
func (s *Scheduler) StartScheduler() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return errors.NewAppError(errors.ValidationInvalid, "Scheduler is already running", nil)
	}

	var err error
	// Schedule morning notifications at 8:00 AM
	_, err = s.cron.AddFunc("0 0 8 * * *", func() {
		if err = s.CleanUp(); err != nil {
			err = errors.Wrap(err, errors.InternalError, "Failed to clean up course entries")
			return
		}

		if err = s.GetCourseEntries(); err != nil {
			err = errors.Wrap(err, errors.InternalError, "Failed to get course entries")
			return
		}

		if err = s.ScheduleCourseNotifications(s.courseEntries); err != nil {
			err = errors.Wrap(err, errors.InternalError, "Failed to schedule course notifications")
			return
		}

		log.Println(err)

	})

	_, err = s.cron.AddFunc("@every 30m", func() {
		if err = s.UpdateCourseEntries(); err != nil {
			err = errors.Wrap(err, errors.InternalError, "Failed to update course entries")
			return
		}
	})
	if err != nil {
		return err
	}

	s.cron.Start()
	s.isRunning = true

	return nil
}

// StopScheduler gracefully stops the scheduler and cleans up resources
func (s *Scheduler) StopScheduler() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	// Clean up course entries
	if err := s.CleanUp(); err != nil {
		return errors.Wrap(err, errors.InternalError, "Failed to clean up course entries")
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

type CourseEntry struct {
	ID      cron.EntryID
	course  models.LocalCourse
	pattern string
	message models.Message
}

func (s *Scheduler) GetCourseEntries() error {
	var entries map[uint]*CourseEntry = make(map[uint]*CourseEntry)

	courses, err := models.GetActiveCourses(s.db)
	if err != nil {
		return errors.HandleDBReadError(err)
	}

	for _, course := range courses {

		notifTime, err := GetNotificationTime(course.Schedule)
		if err != nil {
			return errors.Wrap(err, errors.InternalError, "Failed to get notification time")
		}

		entries[course.ID] = &CourseEntry{
			course:  course,
			pattern: fmt.Sprintf("0 %d %d * * *", notifTime.Minute(), notifTime.Hour()),
			message: models.Message{
				Type:    models.MessageCourse,
				Title:   fmt.Sprintf("%s - %s", course.Code, course.Name),
				Message: fmt.Sprintf("You have a class at %v in 30 minutes", notifTime.Format("15:04")),
			},
		}
	}
	s.courseEntries = entries

	return nil
}

func (s *Scheduler) ScheduleCourseNotifications(entries map[uint]*CourseEntry) error {
	for _, entry := range entries {

		entryID, err := s.cron.AddFunc(entry.pattern, func() {
			if err := beeep.Notify(entry.message.Title, entry.message.Message, ""); err != nil {
				log.Printf("[Scheduler] Error sending system notification: %v", err)
			}
		})
		if err != nil {
			log.Printf("Schedule: %v", err)
			return errors.Wrap(err, errors.InternalError, "Failed to add cron entry")
		}
		entry.ID = entryID

	}
	return nil
}

func (s *Scheduler) RemoveCourseEntry(entryID cron.EntryID) {
	s.cron.Remove(entryID)
}

func (s *Scheduler) UpdateCourseEntries() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	courses, err := models.GetActiveCourses(s.db)
	if err != nil {
		return errors.HandleDBReadError(err)
	}
	for _, course := range courses {
		entry, ok := s.courseEntries[course.ID]
		if !ok {
			continue
		}
		if course.Schedule != entry.course.Schedule {
			// Remove from cache entries
			delete(s.courseEntries, entry.course.ID)

			// remove from cron
			s.RemoveCourseEntry(entry.ID)

			// Compute new entry

			notifTime, err := GetNotificationTime(course.Schedule)
			if err != nil {
				return errors.Wrap(err, errors.InternalError, "Failed to get notification time")
			}

			entry.pattern = fmt.Sprintf("%d %d * * *", notifTime.Minute(), notifTime.Hour())
			entry.message = models.Message{
				Type:    models.MessageCourse,
				Title:   fmt.Sprintf("%s - %s", course.Code, course.Name),
				Message: fmt.Sprintf("You have a class at %v in 30 minutes", notifTime.Format("15:04")),
			}

			// Schedule new entry
			var entryID cron.EntryID
			if entryID, err = s.cron.AddFunc(entry.pattern, func() {
				if err := beeep.Notify(entry.message.Title, entry.message.Message, ""); err != nil {
					log.Printf("[Scheduler] Error sending system notification: %v", err)
				}
			}); err != nil {
				return errors.Wrap(err, errors.InternalError, "Failed to schedule new cron entry")
			}
			entry.ID = entryID
			s.courseEntries[course.ID] = entry
		}
	}
	return nil
}

// ListCronEntries lists all cron entries
func (s *Scheduler) ListCronEntries() ([]cron.Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cron.Entries(), nil
}

// CleanUp removes all course entries from the cron scheduler
func (s *Scheduler) CleanUp() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, entry := range s.courseEntries {
		s.RemoveCourseEntry(entry.ID)
	}
	s.courseEntries = make(map[uint]*CourseEntry)
	return nil
}

// Helpers

// GetNotificationTime returns the time 30 minutes before the start of the course
func GetNotificationTime(schedule string) (time.Time, error) {

	parsedSchedule, err := models.ParseSchedule(schedule)
	if err != nil {
		return time.Time{}, errors.Wrap(err, errors.InternalError, "Failed to parse schedule")
	}

	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), parsedSchedule.StartTime, parsedSchedule.StartMinute, 0, 0, now.Location())
	notifTime := startTime.Add(-30 * time.Minute)
	return notifTime, nil
}
