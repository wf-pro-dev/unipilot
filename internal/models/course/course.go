package course

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"unipilot/internal/models/user"

	"gorm.io/gorm"
)

// Course represents a school course
type Course struct {
	gorm.Model
	UserID          uint      `gorm:"not null"`
	LocalID         uint      `gorm:"not null"`
	User            user.User `gorm:"foreignKey:UserID;references:ID"`
	NotionID        string
	Code            string `gorm:"index:idx_courses_user_code_active,unique,where:deleted_at IS NULL;not null"`
	Name            string `gorm:"not null"`
	Color           string `gorm:"default:bg-blue-500"`
	Duration        string
	RoomNumber      string
	StartDate       time.Time
	EndDate         time.Time
	Schedule        string
	Credits         int
	Semester        string
	Instructor      string
	InstructorEmail string
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
		return fmt.Errorf("course with code '%s' already exists for this user", c.Code)
	}
	return nil
}

func NewCourse() *Course {
	fmt.Println("===== Creating new Course =====")

	course := &Course{}
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("The Name: ")
	scanner.Scan()
	course.Name = scanner.Text()

	fmt.Printf("The Code: ")
	scanner.Scan()
	course.Code = scanner.Text()

	fmt.Printf("The Room Number: ")
	scanner.Scan()
	course.RoomNumber = scanner.Text()

	fmt.Printf("The Duration: ")
	scanner.Scan()
	course.Duration = scanner.Text()

	return course
}

func Get_Course_byId(id uint, db *gorm.DB) (*Course, error) {
	course := &Course{}
	err := db.Preload("User").
		Where("id = ?", id).
		First(course).Error

	if err != nil {
		return nil, err
	}
	return course, nil
}
func Get_Course_byLocalId(id uint, db *gorm.DB) (*Course, error) {
	course := &Course{}
	err := db.Preload("User").
		Where("local_id = ?", id).
		First(course).Error

	if err != nil {
		return nil, err
	}
	return course, nil
}
func Get_Course_byCode(code, user_id string, db *gorm.DB) (*Course, error) {
	course := &Course{}
	err := db.Where("code = ? AND user_id = ?", code, user_id).First(course).Error
	if err != nil {
		return nil, err
	}

	return course, nil
}

func Get_Course_byNotionID(notion_id string, db *gorm.DB) *Course {

	course := &Course{}
	err := db.Where("notion_id = ?", notion_id).First(course).Error
	if err != nil {
		log.Fatalln("Error getting course with notion id: ", err)
		return nil
	}

	return course
}

func (c *Course) ToMap() map[string]string {
	return map[string]string{
		"id":               strconv.Itoa(int(c.ID)),
		"local_id":         strconv.Itoa(int(c.LocalID)),
		"user_id":          strconv.Itoa(int(c.UserID)),
		"notion_id":        c.NotionID,
		"name":             c.Name,
		"code":             c.Code,
		"room_number":      c.RoomNumber,
		"duration":         c.Duration,
		"color":            c.Color,
		"start_date":       c.StartDate.Format(time.DateOnly),
		"end_date":         c.EndDate.Format(time.DateOnly),
		"schedule":         c.Schedule,
		"semester":         c.Semester,
		"instructor":       c.Instructor,
		"instructor_email": c.InstructorEmail,
		"credits":          strconv.Itoa(int(c.Credits)),
		"created_at":       c.CreatedAt.Format(time.DateOnly),
		"updated_at":       c.UpdatedAt.Format(time.DateOnly),
	}
}
