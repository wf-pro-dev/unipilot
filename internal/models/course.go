package models

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"unipilot/internal/errors"

	"gorm.io/gorm"
)

type BaseCourse struct {
	Code            string `gorm:"index:idx_courses_user_code_active,unique,where:deleted_at IS NULL;not null"`
	Name            string `gorm:"not null"`
	Color           string `gorm:"default:bg-blue-500"`
	Location        string
	StartDate       time.Time
	EndDate         time.Time
	Schedule        string
	Credits         int
	Semester        string
	Instructor      string
	InstructorEmail string
	ParentID        uint `gorm:"default:0"`
}

// Course represents a school course
type Course struct {
	gorm.Model
	BaseCourse

	UserID uint `gorm:"not null;index"`
	// Common fields

	// Relationships
	Parent   *Course  `gorm:"foreignKey:ParentID;references:ID"`
	Children []Course `gorm:"foreignKey:ParentID"`

	User        User         `gorm:"foreignKey:UserID;references:ID"`
	Assignments []Assignment `gorm:"foreignKey:CourseID;references:ID"`
	Notes       []Note       `gorm:"foreignKey:CourseID;references:ID"`
}

// LocalCourse represents a course in the local database
// This is for local operations and caching remote metadata
type LocalCourse struct {
	gorm.Model
	BaseCourse

	RemoteID uint `gorm:"unique"`

	Assignments []LocalAssignment `gorm:"foreignKey:CourseID;references:ID"`
	Notes       []LocalNote       `gorm:"foreignKey:CourseID;references:ID"`
}

// Course Link Request
type CourseLinkRequest struct {
	ID         uint `gorm:"primaryKey"`
	OwnerID    uint `gorm:"not null;index"`
	ReceiverID uint `gorm:"not null;index"`
	CourseID   uint `gorm:"not null;index"`

	Owner    User   `gorm:"foreignKey:OwnerID;references:ID"`
	Receiver User   `gorm:"foreignKey:ReceiverID;references:ID"`
	Course   Course `gorm:"foreignKey:CourseID;references:ID"`
}

func (c *BaseCourse) ToMap() map[string]string {
	return map[string]string{
		"name":             c.Name,
		"code":             c.Code,
		"color":            c.Color,
		"location":         c.Location,
		"start_date":       c.StartDate.Format(time.DateOnly),
		"end_date":         c.EndDate.Format(time.DateOnly),
		"schedule":         c.Schedule,
		"semester":         c.Semester,
		"instructor":       c.Instructor,
		"instructor_email": c.InstructorEmail,
		"credits":          strconv.Itoa(int(c.Credits)),
	}
}

func (c *Course) ToMap() map[string]string {
	cMap := c.BaseCourse.ToMap()
	cMap["user_id"] = strconv.Itoa(int(c.UserID))
	return cMap
}

func (c *LocalCourse) ToMap() map[string]string {
	cMap := c.BaseCourse.ToMap()
	cMap["remote_id"] = strconv.Itoa(int(c.RemoteID))
	return cMap
}

func (c *Course) ToLocal() *LocalCourse {
	localCourse := &LocalCourse{
		BaseCourse: c.BaseCourse,
		RemoteID:   c.ID,
	}
	return localCourse
}

func (lc *LocalCourse) ToRemote() *Course {

	c := &Course{
		BaseCourse: lc.BaseCourse,
	}

	return c
}

// BeforeCreate is a GORM hook that runs before creating a record
func (c *Course) BeforeCreate(tx *gorm.DB) error {
	// Check if a course with the same code exists for this user (including soft-deleted ones)
	var existingCourse Course
	if err := tx.Unscoped().Where("code = ? AND user_id = ?", c.Code, c.UserID).First(&existingCourse).Error; err == nil {
		// A course with this code exists for this user, check if it's soft-deleted
		if existingCourse.DeletedAt.Valid {
			// If it's soft-deleted, we can reuse the code
			return nil
		}
		// If it's not soft-deleted, return an error
		return errors.NewAppError(errors.DBConstraintViolation, "Course with code already exists for this user", nil)
	}
	return nil
}

func (c *Course) BeforeDelete(tx *gorm.DB) error {

	assignments, err := GetAssignmentsByCourse(c.ID, tx)
	if err != nil {
		return errors.Wrap(err, errors.DBQueryFailed, "Error getting assignments by course")
	}

	notes, err := GetNotesByCourse(c.ID, tx)
	if err != nil {
		return errors.Wrap(err, errors.DBQueryFailed, "Error getting notes by course")
	}

	for _, assignment := range assignments {
		err := tx.Delete(&assignment).Error
		if err != nil {
			return errors.Wrap(err, errors.DBQueryFailed, "Error deleting assignment")
		}
	}

	for _, note := range notes {
		err := tx.Delete(&note).Error
		if err != nil {
			return errors.Wrap(err, errors.DBQueryFailed, "Error deleting note")
		}
	}

	// Retrieve qdrantClient from transaction context
	return nil
}

func (lc *LocalCourse) BeforeDelete(tx *gorm.DB) error {
	assignments, err := GetLAssignmentsByCourse(lc.Code, tx)
	if err != nil {
		return errors.Wrap(err, errors.DBQueryFailed, "Error getting assignments by course")
	}

	notes, err := GetLNotesByCourse(lc.ID, tx)
	if err != nil {
		return errors.Wrap(err, errors.DBQueryFailed, "Error getting notes by course")
	}

	for _, assignment := range assignments {
		err := tx.Delete(&assignment).Error
		if err != nil {
			return errors.Wrap(err, errors.DBQueryFailed, "Error deleting assignment")
		}
	}

	for _, note := range notes {
		err := tx.Delete(&note).Error
		if err != nil {
			return errors.Wrap(err, errors.DBQueryFailed, "Error deleting note")
		}
	}
	return nil
}

// GET Operations

func GetCourse(id uint, db *gorm.DB) (*Course, error) {
	course := &Course{}
	err := db.First(&course, id).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return course, nil
}

func GetLCourse(id uint, db *gorm.DB) (*LocalCourse, error) {
	course := &LocalCourse{}
	err := db.First(&course, id).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return course, nil
}

func GetCourses(userID uint, db *gorm.DB) ([]Course, error) {
	var courses []Course
	err := db.Where("user_id = ?", userID).Find(&courses).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return courses, nil
}

func GetLCourses(db *gorm.DB) ([]LocalCourse, error) {
	var courses []LocalCourse
	err := db.Find(&courses).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return courses, nil
}

func GetCoursesLinked(courseID uint, db *gorm.DB) ([]Course, error) {
	var courses []Course
	err := db.Where("id = ? AND EXISTS (SELECT 1 FROM courses AS c WHERE c.parent_id = courses.id)", courseID).
		Find(&courses).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return courses, err
}

// Other Operations

// ParsedSchedule represents a parsed course schedule
type ParsedSchedule struct {
	Days        []int
	StartTime   int // 24-hour format
	StartMinute int
	EndTime     int // 24-hour format
	EndMinute   int
}

// ParseSchedule parses a schedule string into a ParsedSchedule object
// Schedule format: "Mon Wed Fri 1:00 PM - 2:00 PM" or similar
func ParseSchedule(schedule string) (*ParsedSchedule, error) {
	if schedule == "" {
		return nil, fmt.Errorf("empty schedule")
	}

	// Day abbreviations mapping
	dayMapping := map[string]int{
		"M": 1, "Mo": 1, "Mon": 1,
		"T": 2, "Tu": 2, "Tue": 2,
		"W": 3, "We": 3, "Wed": 3,
		"Th": 4, "Thu": 4,
		"F": 5, "Fr": 5, "Fri": 5,
		"S": 6, "Sa": 6, "Sat": 6,
		"Su": 0, "Sun": 0,
	}

	// Parse time pattern: "1:00 PM - 2:00 PM"
	timePattern := `(\d{1,2}):(\d{2})\s*(AM|PM)\s*-\s*(\d{1,2}):(\d{2})\s*(AM|PM)`
	re := regexp.MustCompile(timePattern)
	matches := re.FindStringSubmatch(schedule)
	if len(matches) < 7 {
		return nil, fmt.Errorf("invalid schedule format")
	}

	startHour, _ := strconv.Atoi(matches[1])
	startMin, _ := strconv.Atoi(matches[2])
	startPeriod := matches[3]
	endHour, _ := strconv.Atoi(matches[4])
	endMin, _ := strconv.Atoi(matches[5])
	endPeriod := matches[6]

	// Convert to 24-hour format
	if startPeriod == "PM" && startHour != 12 {
		startHour += 12
	}
	if startPeriod == "AM" && startHour == 12 {
		startHour = 0
	}
	if endPeriod == "PM" && endHour != 12 {
		endHour += 12
	}
	if endPeriod == "AM" && endHour == 12 {
		endHour = 0
	}

	// Parse days (everything before the time)
	daysPart := re.Split(schedule, 2)[0]
	dayTokens := strings.Fields(daysPart)
	days := []int{}
	for _, token := range dayTokens {
		if day, ok := dayMapping[token]; ok {
			days = append(days, day)
		}
	}

	return &ParsedSchedule{
		Days:        days,
		StartTime:   startHour,
		StartMinute: startMin,
		EndTime:     endHour,
		EndMinute:   endMin,
	}, nil
}
