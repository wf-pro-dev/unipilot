package app

import (
	"fmt"
	"unipilot/internal/models/assignment"
	"unipilot/internal/models/course"
	"unipilot/internal/models/user"
	"unipilot/internal/storage"

	"gorm.io/gorm"
)

// DatabaseHelper provides database operations for the exposed structs
type DatabaseHelper struct {
	db     *gorm.DB
	userID uint
}

// NewDatabaseHelper creates a new database helper
func NewDatabaseHelper() (*DatabaseHelper, error) {
	db, userID, err := storage.GetLocalDB()
	if err != nil {
		return nil, err
	}

	return &DatabaseHelper{db: db, userID: userID}, nil
}

// GetAssignment retrieves an assignment by ID
func (h *DatabaseHelper) GetAssignment(id uint) (*assignment.LocalAssignment, error) {
	return assignment.Get_Local_Assignment_byId(id, h.db)
}

// GetAssignments retrieves all assignments for a user
func (h *DatabaseHelper) GetAssignments() ([]assignment.LocalAssignment, error) {
	var LocalAssignment []assignment.LocalAssignment
	err := h.db.Preload("Course").Preload("Type").Preload("Status").Order("deadline DESC").Order("created_at DESC").Find(&LocalAssignment).Error
	return LocalAssignment, err
}

// GetCourse retrieves a course by ID
func (h *DatabaseHelper) GetCourse(id uint) (*course.Course, error) {
	return course.Get_Course_byId(id, h.db)
}

// GetCourses retrieves all courses for a user
func (h *DatabaseHelper) GetCourses() ([]course.LocalCourse, error) {
	var LocalCourse []course.LocalCourse
	err := h.db.Find(&LocalCourse).Error
	return LocalCourse, err
}

// GetUser retrieves a user by ID
func (h *DatabaseHelper) GetUser(id uint) (*user.User, error) {
	var u user.User
	err := h.db.First(&u, id).Error
	return &u, err
}

// GetDB returns the database connection
func (h *DatabaseHelper) GetDB() *gorm.DB {
	return h.db
}

// GetCurrentUserID returns the current user ID
func (h *DatabaseHelper) GetCurrentUserID() uint {
	return h.userID
}

// CreateAssignment creates a new assignment
func (h *DatabaseHelper) CreateAssignment(assignment *assignment.LocalAssignment) error {
	h.db = h.db.Debug()
	return h.db.Create(assignment).Error
}

// UpdateAssignment updates an existing assignment
func (h *DatabaseHelper) UpdateAssignment(LocalAssignment *assignment.LocalAssignment, column, value string) error {
	// Only update the assignment fields, not the related course data
	return h.db.Exec(fmt.Sprintf("UPDATE local_assignments SET %s = '%s' WHERE id = '%d'", column, value, LocalAssignment.ID)).Error
}

// DeleteAssignment deletes an assignment
func (h *DatabaseHelper) DeleteAssignment(assignment *assignment.LocalAssignment) error {
	return h.db.Delete(assignment).Error
}

// CreateCourse creates a new course
func (h *DatabaseHelper) CreateCourse(course *course.Course) error {
	course.UserID = h.userID
	return h.db.Create(course).Error
}

// UpdateCourse updates an existing course
func (h *DatabaseHelper) UpdateCourse(LocalCourse *course.LocalCourse, column, value string) error {
	// Only update the assignment fields, not the related course data
	return h.db.Exec(fmt.Sprintf("UPDATE local_courses SET %s = '%s' WHERE id = '%d'", column, value, LocalCourse.ID)).Error
}

// DeleteCourse deletes a course
func (h *DatabaseHelper) DeleteCourse(course *course.LocalCourse) error {
	// Get all assignment related to the course
	assignments := []assignment.LocalAssignment{}
	err := h.db.Where("course_code = ?", course.Code).Find(&assignments).Error
	if err != nil {
		return err
	}

	// Delete all assignment related to the course
	for _, assignment := range assignments {
		if err := h.DeleteAssignment(&assignment); err != nil {
			return err
		}
	}

	// Delete all notes related to the course

	return h.db.Delete(course).Error
}
