package user

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// User represents the application user
type User struct {
	gorm.Model
	Username     string `gorm:"unique;not null"`
	Email        string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"`

	Avatar     string
	University string
	Semester   string
	Year       string

	IsVerified bool   `gorm:"default:false"`
	Language   string `gorm:"default:'en'"`

	CoursesCode []string `gorm:"-"`
	LastSync    *time.Time

	// Follow relationships
	Followers []User `gorm:"many2many:follows;foreignKey:ID;joinForeignKey:FollowerID;References:ID;joinReferences:FollowedID"`
	Following []User `gorm:"many2many:follows;foreignKey:ID;joinForeignKey:FollowedID;References:ID;joinReferences:FollowerID"`
}

func (u *User) ToMap() map[string]interface{} {
	if u == nil {
		return nil
	}

	return map[string]interface{}{
		"id":            u.ID,
		"username":      u.Username,
		"email":         u.Email,
		"avatar":        u.Avatar,
		"university":    u.University,
		"semester":      u.Semester,
		"year":          u.Year,
		"is_verified":   u.IsVerified,
		"language":      u.Language,
		"courses_codes": u.CoursesCode,
		"last_sync":     u.LastSync,
		"created_at":    u.CreatedAt,
		"updated_at":    u.UpdatedAt,
	}
}

func Get_User_by_NotionID(notion_id string, db *gorm.DB) (*User, error) {
	u := &User{}
	err := db.Where("notion_id = ?", notion_id).First(u).Error
	if err != nil {
		return nil, fmt.Errorf("Error getting user with notion id: ", err)
	}
	return u, nil
}
