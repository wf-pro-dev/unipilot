package database

import (
	"fmt"

	"unipilot/internal/models/assignment"
	"unipilot/internal/models/course"
	"unipilot/internal/models/note"
	"unipilot/internal/models/notifications"
	"unipilot/internal/models/user"

	"gorm.io/gorm"
)

// GetAssignment retrieves an assignment by ID
func (h *Database) GetAssignment(id uint) (*assignment.LocalAssignment, error) {
	return assignment.Get_Local_Assignment_byId(id, h.db)
}

// GetAssignments retrieves all assignments for a user
func (h *Database) GetAssignments() ([]assignment.LocalAssignment, error) {
	var LocalAssignment []assignment.LocalAssignment
	err := h.db.Preload("Course").Preload("Type").Preload("Status").Order("deadline DESC").Order("created_at DESC").Find(&LocalAssignment).Error
	return LocalAssignment, err
}

// GetCourse retrieves a course by ID
func (h *Database) GetCourse(id uint) (*course.Course, error) {
	return course.Get_Course_byId(id, h.db)
}

// GetCourses retrieves all courses for a user
func (h *Database) GetCourses() ([]course.LocalCourse, error) {
	var LocalCourse []course.LocalCourse
	err := h.db.Find(&LocalCourse).Error
	return LocalCourse, err
}

// GetUser retrieves a user by ID
func (h *Database) GetUser(id uint) (*user.User, error) {
	var u user.User
	err := h.db.First(&u, id).Error
	return &u, err
}

// GetDB returns the database connection
func (h *Database) GetDB() *gorm.DB {
	return h.db
}

// CreateAssignment creates a new assignment
func (h *Database) CreateAssignment(assignment *assignment.LocalAssignment) error {
	return h.db.Create(assignment).Error
}

// UpdateAssignment updates an existing assignment
func (h *Database) UpdateAssignment(LocalAssignment *assignment.LocalAssignment, column, value string) error {
	// Only update the assignment fields, not the related course data
	return h.db.Exec(fmt.Sprintf("UPDATE local_assignments SET %s = '%s', updated_at = CURRENT_TIMESTAMP WHERE id = '%d'", column, value, LocalAssignment.ID)).Error
}

// DeleteAssignment deletes an assignment
func (h *Database) DeleteAssignment(assignment *assignment.LocalAssignment) error {
	return h.db.Delete(assignment).Error
}

// CreateCourse creates a new course
func (h *Database) CreateCourse(course *course.Course) error {

	return h.db.Create(course).Error
}

// UpdateCourse updates an existing course
func (h *Database) UpdateCourse(LocalCourse *course.LocalCourse, column, value string) error {
	// Only update the assignment fields, not the related course data
	return h.db.Exec(fmt.Sprintf("UPDATE local_courses SET %s = '%s', updated_at = CURRENT_TIMESTAMP WHERE id = '%d'", column, value, LocalCourse.ID)).Error
}

// DeleteCourse deletes a course
func (h *Database) DeleteCourse(course *course.LocalCourse) error {
	return h.db.Delete(course).Error
}

// GetNotes returns all notes for the current user
func (h *Database) GetNotes() ([]note.LocalNote, error) {
	var LocalNote []note.LocalNote
	err := h.db.Preload("Course").Find(&LocalNote).Order("created_at DESC").Error
	return LocalNote, err
}

// CreateNote creates a new note
func (h *Database) CreateNote(note *note.LocalNote) error {
	return h.db.Create(note).Error
}

// UpdateNote updates an existing note
func (h *Database) UpdateNote(LocalNote *note.LocalNote, column, value string) error {
	return h.db.Exec(fmt.Sprintf("UPDATE local_notes SET %s = '%s', updated_at = CURRENT_TIMESTAMP WHERE id = '%d'", column, value, LocalNote.ID)).Error
}

// DeleteNote deletes a note
func (h *Database) DeleteNote(note *note.LocalNote) error {
	return h.db.Delete(note).Error
}

// GetNotifications returns all notifications for the current user
func (h *Database) GetNotifications() ([]notifications.LocalNotification, error) {
	h.db = h.db.Debug()
	var LocalNotification []notifications.LocalNotification
	err := h.db.
		Where("type = ?", notifications.NotificationFollow).Or("type = ?", notifications.NotificationSync).
		Find(&LocalNotification).
		Order("created_at DESC").Error
	return LocalNotification, err
}

func (h *Database) DeleteNotification(notification *notifications.LocalNotification) error {
	return h.db.Delete(notification).Error
}
