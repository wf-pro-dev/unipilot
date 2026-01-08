package models

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"unipilot/internal/errors"
)

// User represents the application user
type User struct {
	gorm.Model
	Username string `gorm:"unique;not null" validate:"required,min=3,max=30"`
	Email    string `gorm:"unique;not null" validate:"required,email"`

	Password     string `gorm:"-" validate:"required,min=8,max=100"`
	PasswordHash string `gorm:"not null"`

	Avatar     string
	University string `validate:"required,min=3,max=100"`
	Semester   string `validate:"required,min=1,max=20"`
	Year       string `validate:"required,min=4,max=4"`

	IsVerified bool   `gorm:"default:false"`
	Language   string `gorm:"default:'en'" validate:"required,oneof=en fr es de it pt nl ru tr ja zh ko ar he"`

	CoursesCode []string `gorm:"-"`
	LastSync    *time.Time

	// Follow relationships
	Courses          []*Course           `gorm:"foreignKey:UserID;references:ID"`
	Assignments      []*Assignment       `gorm:"foreignKey:UserID;references:ID"`
	Notes            []*Note             `gorm:"foreignKey:UserID;references:ID"`
	OwnerRequests    []*CourseInvitation `gorm:"foreignKey:OwnerID;references:ID"`
	ReceiverRequests []*CourseInvitation `gorm:"foreignKey:ReceiverID;references:ID"`
	Followers        []*User             `gorm:"many2many:follows;foreignKey:ID;joinForeignKey:FollowerID;References:ID;joinReferences:FollowedID"`
	Following        []*User             `gorm:"many2many:follows;foreignKey:ID;joinForeignKey:FollowedID;References:ID;joinReferences:FollowerID"`
}

// START; TO MAP FUNCTIONS

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

// END; TO MAP FUNCTIONS

// START; GET FUNCTIONS

func GetUser(id uint, db *gorm.DB) (*User, error) {
	var user User
	err := db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return &user, nil
}

func GetUsers(db *gorm.DB) ([]User, error) {
	var users []User
	err := db.Find(&users).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return users, nil
}

type UserCourseCodes struct {
	UserID uint   `gorm:"column:user_id"`
	Code   string `gorm:"column:code"`
}

func GetUsersCourseCodes(userIDs []uint, db *gorm.DB) ([]UserCourseCodes, error) {
	var courseCodes []UserCourseCodes
	if err := db.Model(&Course{}).
		Select("user_id, code").
		Where("user_id IN ? AND deleted_at IS NULL", userIDs).
		Find(&courseCodes).Error; err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return courseCodes, nil
}

func GetCourseUsers(courseID uint, db *gorm.DB) ([]uint, error) {
	var userIDs []uint

	var parentCourse Course
	if err := db.Select("user_id").First(&parentCourse, courseID).Error; err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	userIDs = append(userIDs, parentCourse.UserID)

	err := db.Model(&Course{}).
		Where("parent_id = ?", courseID).
		Pluck("user_id", &userIDs).Error

	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}

	return userIDs, nil
}

func GetUserClusterIDs(userID uint, db *gorm.DB) ([]uint, error) {
	var clusterIDs []uint

	// The CourseID in the invitation is ALWAYS the Root ID by your design
	err := db.Model(&CourseInvitation{}).
		Where("(owner_id = ? OR receiver_id = ?) AND status = 'accepted'", userID, userID).
		Distinct("course_id").
		Pluck("course_id", &clusterIDs).Error

	return clusterIDs, err
}

// END; GET FUNCTIONS

// START; VALIDATION FUNCTIONS

func (u *User) Validate() error {
	// Username can contain spaces
	u.Email = strings.TrimSpace(u.Email)
	u.Password = strings.TrimSpace(u.Password)
	u.University = strings.TrimSpace(u.University)
	u.Semester = strings.TrimSpace(u.Semester)
	u.Year = strings.TrimSpace(u.Year)
	u.Language = strings.TrimSpace(u.Language)

	validate := validator.New()
	if err := validate.Struct(u); err != nil {
		return errors.Wrap(err, errors.ValidationInvalid, "User Validation failed")
	}

	err := isValidUsername(u.Username)
	if err != nil {
		return err
	}
	err = isValidEmail(u.Email)
	if err != nil {
		return err
	}
	err = isValidPassword(u.Password)
	if err != nil {
		return err
	}

	return nil
}

func isValidUsername(username string) error {
	pattern := `^[a-zA-Z0-9_\s-]+$`
	re := regexp.MustCompile(pattern)
	if !re.MatchString(username) {
		return errors.Wrap(fmt.Errorf("username invalid"), errors.ValidationInvalid, "Username invalid")
	}
	return nil
}

func isValidEmail(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return errors.Wrap(err, errors.ValidationInvalid, "Email invalid")
	}
	return nil
}

func isValidPassword(password string) error {
	uppercase := regexp.MustCompile(`[A-Z]`)
	lowercase := regexp.MustCompile(`[a-z]`)
	number := regexp.MustCompile(`[0-9]`)
	special := regexp.MustCompile(`[^a-zA-Z0-9]`)

	if len(password) > 100 {
		return errors.Wrap(fmt.Errorf("password too long"), errors.ValidationInvalid, "Password must be less than 100 characters")
	}

	if !uppercase.MatchString(password) {
		return errors.Wrap(fmt.Errorf("password must contain at least one uppercase letter"), errors.ValidationInvalid, "Password must contain at least one uppercase letter")
	}
	if !lowercase.MatchString(password) {
		return errors.Wrap(fmt.Errorf("password must contain at least one lowercase letter"), errors.ValidationInvalid, "Password must contain at least one lowercase letter")
	}
	if !number.MatchString(password) {
		return errors.Wrap(fmt.Errorf("password must contain at least one number"), errors.ValidationInvalid, "Password must contain at least one number")
	}
	if !special.MatchString(password) {
		return errors.Wrap(fmt.Errorf("password must contain at least one special character"), errors.ValidationInvalid, "Password must contain at least one special character")
	}
	return nil
}

// END; VALIDATION FUNCTIONS

// Follow represents a follow relationship between users
type Follow struct {
	gorm.Model
	FollowerID uint `gorm:"not null;index"` // User who is following
	FollowedID uint `gorm:"not null;index"` // User who is being followed

	// Foreign key relationships
	Follower User `gorm:"foreignKey:FollowerID;references:ID;constraint:OnDelete:CASCADE"`
	Followed User `gorm:"foreignKey:FollowedID;references:ID;constraint:OnDelete:CASCADE"`
}

// Helper functions
func (f *Follow) ToMap() map[string]interface{} {
	if f == nil {
		return nil
	}

	return map[string]interface{}{
		"id":          f.ID,
		"follower_id": f.FollowerID,
		"followed_id": f.FollowedID,
		"created_at":  f.CreatedAt,
		"updated_at":  f.UpdatedAt,
	}
}

// Check if user A is following user B
func IsFollowing(followerID, followedID uint, db *gorm.DB) (bool, error) {
	var count int64
	err := db.Model(&Follow{}).Where("follower_id = ? AND followed_id = ?", followerID, followedID).Count(&count).Error
	if err != nil {
		return false, errors.HandleDBReadError(err)
	}
	return count > 0, nil
}

// Get followers count for a user
func GetFollowersCount(userID uint, db *gorm.DB) (int, error) {
	var count int64
	err := db.Model(&Follow{}).Where("followed_id = ?", userID).Count(&count).Error
	if err != nil {
		return 0, errors.HandleDBReadError(err)
	}
	return int(count), nil
}

// Get following count for a user
func GetFollowingCount(userID uint, db *gorm.DB) (int, error) {
	var count int64
	err := db.Model(&Follow{}).Where("follower_id = ?", userID).Count(&count).Error
	if err != nil {
		return 0, errors.HandleDBReadError(err)
	}
	return int(count), nil
}

// Create a follow relationship
func CreateFollow(followerID, followedID uint, db *gorm.DB) error {
	// Check if already following
	isFollowing, err := IsFollowing(followerID, followedID, db)
	if err != nil {
		return errors.HandleDBReadError(err)
	}
	if isFollowing {
		return errors.NewAppError(errors.DBConstraintViolation, "Already following", nil)
	}

	follow := &Follow{
		FollowerID: followerID,
		FollowedID: followedID,
	}
	if err := db.Create(follow).Error; err != nil {
		return errors.HandleDBCreateError(err)
	}

	return nil
}

// Remove a follow relationship
func RemoveFollow(followerID, followedID uint, db *gorm.DB) error {
	err := db.Where("follower_id = ? AND followed_id = ?", followerID, followedID).Delete(&Follow{}).Error
	if err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}

// Get followers list for a user
func GetFollowers(userID uint, limit, offset int, db *gorm.DB) ([]User, error) {
	var followers []User
	err := db.Joins("JOIN follows ON users.id = follows.follower_id").
		Where("follows.followed_id = ? AND follows.deleted_at is NULL", userID).
		Limit(limit).Offset(offset).
		Find(&followers).
		Order("users.username ASC").
		Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return followers, nil
}

// Get following list for a user
func GetFollowing(userID uint, limit, offset int, db *gorm.DB) ([]User, error) {
	var following []User
	err := db.Joins("JOIN follows ON users.id = follows.followed_id").
		Where("follows.follower_id = ? AND follows.deleted_at is NULL", userID).
		Limit(limit).Offset(offset).
		Find(&following).
		Order("users.username ASC").
		Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return following, nil
}

func GetUserCourseInvitations(userID uint, db *gorm.DB) ([]CourseInvitation, error) {
	var invitations []CourseInvitation
	err := db.Preload("Course").Where("receiver_id = ? OR owner_id = ?", userID, userID).Find(&invitations).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return invitations, nil
}
