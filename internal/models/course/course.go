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
	NotionID        string    `gorm:"unique;not null"`
	Code            string    `gorm:"unique;not null"`
	Name            string    `gorm:"not null"`
	Color           string    `gorm:"default:bg-blue-500"`
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
