package models

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"unipilot/internal/errors"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type BaseCourse struct {
	Code            string    `validate:"required,min=1,max=12"`
	Name            string    `gorm:"not null" validate:"required,min=3,max=100"`
	Color           string    `gorm:"default:bg-blue-500" validate:"required"`
	Location        string    `validate:"required,min=3,max=100"`
	StartDate       time.Time `validate:"required"`
	EndDate         time.Time `validate:"required,gtfield=StartDate"`
	Schedule        string    `validate:"required,min=3,max=100"`
	Credits         int       `validate:"required,min=1,max=10"`
	Semester        string    `validate:"required,min=1,max=20"`
	Instructor      string    `validate:"required,min=3,max=100"`
	InstructorEmail string    `validate:"required,email"`
	ParentID        uint      `gorm:"default:null"`
}

// Course represents a school course
type Course struct {
	gorm.Model
	BaseCourse

	UserID uint `gorm:"not null;index" validate:"required"`
	// Common fields

	// Relationships
	Parent   *Course   `gorm:"foreignKey:ParentID;references:ID" validate:"-"`
	Children []*Course `gorm:"foreignKey:ParentID"`

	User        *User         `gorm:"foreignKey:UserID;references:ID" validate:"-"`
	Assignments []*Assignment `gorm:"foreignKey:CourseID;references:ID"`
	Notes       []*Note       `gorm:"foreignKey:CourseID;references:ID"`
}

// LocalCourse represents a course in the local database
// This is for local operations and caching remote metadata
type LocalCourse struct {
	gorm.Model
	BaseCourse

	RemoteID uint `gorm:"unique;default:null"`

	Assignments []*LocalAssignment `gorm:"foreignKey:CourseID;references:ID" validate:"-"`
	Notes       []*LocalNote       `gorm:"foreignKey:CourseID;references:ID" validate:"-"`
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

// END: Conversion Functions

// START: GORM Hooks

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

	assignments, err := c.GetCourseAssignments(tx)
	if err != nil {
		return errors.Wrap(err, errors.DBQueryFailed, "Error getting assignments by course")
	}

	notes, err := c.GetCourseNotes(tx)
	if err != nil {
		return errors.Wrap(err, errors.DBQueryFailed, "Error getting notes by course")
	}

	// Batch delete assignments
	tx.Delete(assignments) //  Uses QdrantClient from transaction context

	// Batch delete notes
	tx.Delete(notes)

	return nil
}

func (lc *LocalCourse) BeforeDelete(tx *gorm.DB) error {
	assignments, err := lc.GetAssignmentsByCourse(tx)
	if err != nil {
		return errors.Wrap(err, errors.DBQueryFailed, "Error getting assignments by course")
	}

	notes, err := lc.GetNotesByCourse(tx)
	if err != nil {
		return errors.Wrap(err, errors.DBQueryFailed, "Error getting notes by course")
	}

	// Batch delete assignments
	tx.Delete(assignments)

	// Batch delete notes
	tx.Delete(notes)

	return nil
}

// END: GORM Hooks

// START: Validation Functions

func (bc *BaseCourse) Validate() error {
	bc.Code = strings.TrimLeft(bc.Code, " ")
	bc.Code = strings.TrimRight(bc.Code, " ")
	bc.Code = strings.ToUpper(bc.Code)

	bc.Name = strings.TrimRight(bc.Name, " ")
	bc.Location = strings.TrimRight(bc.Location, " ")

	bc.Schedule = strings.TrimRight(bc.Schedule, " ")
	bc.Schedule = strings.TrimLeft(bc.Schedule, " ")

	bc.Semester = strings.TrimRight(bc.Semester, " ")
	bc.Semester = strings.TrimLeft(bc.Semester, " ")
	bc.Semester = strings.ToUpper(bc.Semester)

	bc.Instructor = strings.TrimRight(bc.Instructor, " ")
	bc.Instructor = strings.TrimLeft(bc.Instructor, " ")

	bc.InstructorEmail = strings.TrimSpace(bc.InstructorEmail)

	return nil
}

func (c *Course) Validate() error {

	if err := c.BaseCourse.Validate(); err != nil {
		return err
	}

	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return errors.Wrap(err, errors.ValidationInvalid, "Course validation failed")
	}

	if err := isValidCode(c.Code); err != nil {
		return err
	}

	return nil
}

func (lc *LocalCourse) Validate() error {
	if err := lc.BaseCourse.Validate(); err != nil {
		return err
	}

	validate := validator.New()
	if err := validate.Struct(lc); err != nil {
		return errors.Wrap(err, errors.ValidationInvalid, "LocalCourse validation failed")
	}

	if err := isValidCode(lc.Code); err != nil {
		return err
	}

	return nil
}

func isValidCode(code string) error {
	pattern := `^[A-Z0-9\s\-_.]+$`
	re := regexp.MustCompile(pattern)
	if !re.MatchString(code) {
		return errors.Wrap(fmt.Errorf("code invalid"), errors.ValidationInvalid, "Code invalid")
	}
	return nil
}

// END: Validation Functions

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

func GetCoursesByIDs(courseIDs []uint, db *gorm.DB) ([]*Course, error) {
	var courses []*Course
	err := db.Where(courseIDs).Find(&courses).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return courses, nil
}

func (c *Course) IsInCluster(db *gorm.DB) bool {
	// If the course is a parent, it is in a cluster
	if c.ParentID != 0 {
		return true
	}
	// Check if the course is a parent of a cluster
	var courses []Course
	err := db.Model(&Course{}).Where("parent_id = ?", c.ID).Find(&courses).Error
	if err != nil {
		return false
	}

	return len(courses) > 0
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

func GetClusterCourses(rootID uint, db *gorm.DB) ([]uint, error) {
	var courseIDs []uint

	// Find the Root itself and all its Children
	err := db.Model(&Course{}).
		Where("id = ? OR parent_id = ?", rootID, rootID).
		Pluck("id", &courseIDs).Error

	return courseIDs, err
}

func GetClusterUserIDs(rootID uint, db *gorm.DB) ([]uint, error) {
	var userIDs []uint

	err := db.Raw(`
        SELECT DISTINCT user_id FROM (
            SELECT owner_id AS user_id 
            FROM course_invitations
            WHERE course_id = ? AND status = ? AND deleted_at IS NULL
            UNION
            SELECT receiver_id AS user_id 
            FROM course_invitations
            WHERE course_id = ? AND status = ? AND deleted_at IS NULL
        ) AS combined_users
    `, rootID, InvitationAccepted, rootID, InvitationAccepted).
		Pluck("user_id", &userIDs).Error

	return userIDs, err
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

type InvitationStatus string

const (
	InvitationPending  InvitationStatus = "pending"
	InvitationAccepted InvitationStatus = "accepted"
)

// Course Link Request
type CourseInvitation struct {
	gorm.Model
	OwnerID    uint             `gorm:"not null;index" validate:"required,min=1"`
	ReceiverID uint             `gorm:"not null;index;uniqueIndex:idx_invitations,where:deleted_at IS NULL" validate:"required,min=1"`
	SenderID   uint             `gorm:"not null;index;uniqueIndex:idx_invitations,where:deleted_at IS NULL" validate:"required,min=1"`
	CourseID   uint             `gorm:"not null;index;uniqueIndex:idx_invitations,where:deleted_at IS NULL" validate:"required,min=1"`
	CourseCode string           `gorm:"not null" validate:"required,min=3,max=12"`
	Status     InvitationStatus `gorm:"not null;default:pending" validate:"required,oneof=pending accepted"`

	Owner    *User   `gorm:"foreignKey:OwnerID;references:ID" validate:"-"`
	Receiver *User   `gorm:"foreignKey:ReceiverID;references:ID" validate:"-"`
	Sender   *User   `gorm:"foreignKey:SenderID;references:ID" validate:"-"`
	Course   *Course `gorm:"foreignKey:CourseID;references:ID" validate:"-"`
}

func (ci *CourseInvitation) BeforeCreate(tx *gorm.DB) error {
	var course Course
	if err := tx.Select("parent_id").First(&course, ci.CourseID).Error; err != nil {
		return err
	}

	if course.ParentID != 0 {
		return errors.NewAppError(errors.ValidationInvalid,
			"CourseID must be a parent course, not a child", nil)
	}

	if exists := InvitationExists(ci.OwnerID, ci.ReceiverID, ci.CourseID, tx); exists {
		return errors.NewAppError(errors.ValidationInvalid,
			"An Invitation is still active", nil)
	}

	return nil
}

func (ci *CourseInvitation) Validate() error {
	validate := validator.New()
	if err := validate.Struct(ci); err != nil {
		return errors.Wrap(err, errors.ValidationInvalid, "CourseInvitation validation failed")
	}
	return nil
}

func GetCourseInvitation(id uint, db *gorm.DB) (*CourseInvitation, error) {
	invitation := &CourseInvitation{}
	err := db.First(&invitation, id).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return invitation, nil
}

func InvitationExists(ownerID, receiverID, courseID uint, db *gorm.DB) bool {
	threshold := time.Now().Add(-48 * time.Hour)
	var invitation CourseInvitation
	err := db.Model(&CourseInvitation{}).
		Where("owner_id = ? AND receiver_id = ? AND course_id = ? AND deleted_at > ?", ownerID, receiverID, courseID, threshold).
		First(&invitation).
		Error
	return err == nil
}
