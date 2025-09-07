package course

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ParsedSchedule represents a parsed course schedule
type ParsedSchedule struct {
	Days            []int  `json:"days"`
	StartTime       int    `json:"startTime"`
	EndTime         int    `json:"endTime"`
	StartMinute     int    `json:"startMinute"`
	EndMinute       int    `json:"endMinute"`
	StartTimeString string `json:"startTimeString"`
	EndTimeString   string `json:"endTimeString"`
}

type LocalCourse struct {
	gorm.Model
	RemoteID        uint
	Code            string `gorm:"index:idx_course_code_active,unique,where:deleted_at IS NULL"`
	Name            string `gorm:"not null"`
	Location        string
	Color           string
	StartDate       time.Time
	EndDate         time.Time
	Credits         int
	Schedule        string
	Semester        string
	Instructor      string
	InstructorEmail string
	LinkID          uuid.UUID
}

// BeforeCreate is a GORM hook that runs before creating a record
func (c *LocalCourse) BeforeCreate(tx *gorm.DB) error {
	// Check if a course with the same code exists (including soft-deleted ones)
	var existingCourse LocalCourse
	if err := tx.Unscoped().Where("code = ?", c.Code).First(&existingCourse).Error; err == nil {
		// A course with this code exists, check if it's soft-deleted
		if existingCourse.DeletedAt.Valid {
			// If it's soft-deleted, we can reuse the code
			return nil
		}
		// If it's not soft-deleted, return an error
		return fmt.Errorf("course with code '%s' already exists", c.Code)
	}
	return nil
}

func (c *LocalCourse) ToMap() map[string]string {
	return map[string]string{
		"remote_id":        strconv.Itoa(int(c.RemoteID)),
		"code":             c.Code,
		"name":             c.Name,
		"location":         c.Location,
		"color":            c.Color,
		"start_date":       c.StartDate.Format(time.DateOnly),
		"end_date":         c.EndDate.Format(time.DateOnly),
		"schedule":         c.Schedule,
		"credits":          strconv.Itoa(int(c.Credits)),
		"semester":         c.Semester,
		"instructor":       c.Instructor,
		"instructor_email": c.InstructorEmail,
		"link_id":          c.LinkID.String(),
	}
}

// ParseSchedule parses a schedule string and returns a ParsedSchedule struct
func (c *LocalCourse) ParseSchedule(schedule string) (*ParsedSchedule, error) {
	if schedule == "" || schedule == "Async" || schedule == "Asynchronous" {
		return nil, fmt.Errorf("non parsable schedule: %s", schedule)
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

	// Time pattern regex: "1:00 PM - 2:00 PM"
	timePattern := regexp.MustCompile(`(\d{1,2}):(\d{2})\s*(AM|PM)\s*-\s*(\d{1,2}):(\d{2})\s*(AM|PM)`)
	matches := timePattern.FindStringSubmatch(c.Schedule)
	if matches == nil {
		return nil, fmt.Errorf("invalid time format in schedule: %s", c.Schedule)
	}

	// Extract time components
	startHour, _ := strconv.Atoi(matches[1])
	startMin, _ := strconv.Atoi(matches[2])
	startPeriod := matches[3]
	endHour, _ := strconv.Atoi(matches[4])
	endMin, _ := strconv.Atoi(matches[5])
	endPeriod := matches[6]

	// Convert to 24-hour format
	start24 := startHour
	end24 := endHour

	if strings.ToUpper(startPeriod) == "PM" && start24 != 12 {
		start24 += 12
	}
	if strings.ToUpper(startPeriod) == "AM" && start24 == 12 {
		start24 = 0
	}
	if strings.ToUpper(endPeriod) == "PM" && end24 != 12 {
		end24 += 12
	}
	if strings.ToUpper(endPeriod) == "AM" && end24 == 12 {
		end24 = 0
	}

	// Parse days (everything before the time)
	timeIndex := timePattern.FindStringIndex(c.Schedule)
	if timeIndex == nil {
		return nil, fmt.Errorf("could not find time pattern in schedule: %s", c.Schedule)
	}

	daysPart := strings.TrimSpace(schedule[:timeIndex[0]])
	if daysPart == "" {
		return nil, fmt.Errorf("no days found in schedule: %s", c.Schedule)
	}

	// Split days by comma and space
	dayTokens := strings.FieldsFunc(daysPart, func(r rune) bool {
		return r == ',' || r == ' '
	})

	var days []int
	for _, token := range dayTokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if day, exists := dayMapping[token]; exists {
			days = append(days, day)
		}
	}

	if len(days) == 0 {
		return nil, fmt.Errorf("no valid days found in schedule: %s", c.Schedule)
	}

	return &ParsedSchedule{
		Days:            days,
		StartTime:       start24,
		EndTime:         end24,
		StartMinute:     startMin,
		EndMinute:       endMin,
		StartTimeString: fmt.Sprintf("%d:%02d %s", startHour, startMin, startPeriod),
		EndTimeString:   fmt.Sprintf("%d:%02d %s", endHour, endMin, endPeriod),
	}, nil
}
