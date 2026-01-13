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
	cron      *cron.Cron
	db        *gorm.DB
	ctx       context.Context
	cancel    context.CancelFunc
	isRunning bool
	mu        sync.RWMutex
	// Performance optimizations
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

	// Schedule morning notifications at 9:00 AM
	_, err := s.cron.AddFunc("0 0 0 * * *", func() {
		if err := s.CleanUp(); err != nil {
			err = errors.Wrap(err, errors.InternalError, "Failed to clean up course entries")
			return
		}

		if err := s.GetCourseEntries(); err != nil {
			err = errors.Wrap(err, errors.InternalError, "Failed to get course entries")
			return
		}

		if err := s.ScheduleCourseNotifications(s.courseEntries); err != nil {
			err = errors.Wrap(err, errors.InternalError, "Failed to schedule course notifications")
			return
		}
	})

	_, err = s.cron.AddFunc("@every 30m", func() {
		if err := s.UpdateCourseEntries(); err != nil {
			err = errors.Wrap(err, errors.InternalError, "Failed to update course entries")
			return
		}
	})
	if err != nil {
		return err
	}

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
	now := time.Now()
	var entries map[uint]*CourseEntry = make(map[uint]*CourseEntry)

	courses, err := models.GetActiveCourses(s.db)
	if err != nil {
		return errors.HandleDBReadError(err)
	}

	for _, course := range courses {
		schedule, err := models.ParseSchedule(course.Schedule)
		if err != nil {
			return errors.Wrap(err, errors.InternalError, "Failed to parse schedule")
		}

		startTime := time.Date(now.Year(), now.Month(), now.Day(), schedule.StartTime, schedule.StartMinute, 0, 0, now.Location())
		notifTime := startTime.Add(-30 * time.Minute)

		entries[course.ID] = &CourseEntry{
			course:  course,
			pattern: fmt.Sprintf("0 %d %d * * *", notifTime.Format("15"), notifTime.Format("4")),
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

		if entryID, err := s.cron.AddFunc(entry.pattern, func() {
			if err := beeep.Notify(entry.message.Title, entry.message.Message, ""); err != nil {
				log.Printf("[Scheduler] Error sending system notification: %v", err)
			}
		}); err != nil {
			entry.ID = entryID
			return errors.Wrap(err, errors.InternalError, "Failed to add cron entry")
		}
	}
	return nil
}

func (s *Scheduler) RemoveCourseEntry(entryID cron.EntryID) {
	s.cron.Remove(entryID)
}

func (s *Scheduler) UpdateCourseEntries() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
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
			schedule, err := models.ParseSchedule(course.Schedule)
			if err != nil {
				return errors.Wrap(err, errors.InternalError, "Failed to parse schedule")
			}
			startTime := time.Date(now.Year(), now.Month(), now.Day(), schedule.StartTime, schedule.StartMinute, 0, 0, now.Location())
			notifTime := startTime.Add(-30 * time.Minute)
			entry.pattern = fmt.Sprintf("0 %d %d * * *", notifTime.Format("15"), notifTime.Format("04"))
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

func (s *Scheduler) CleanUp() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, entry := range s.courseEntries {
		s.RemoveCourseEntry(entry.ID)
	}
	s.courseEntries = make(map[uint]*CourseEntry)
	return nil
}
