package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"unipilot/internal/errors"
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

func GetUserById(id uint, db *gorm.DB) (*User, error) {
	u := &User{}
	err := db.Where("id = ?", id).First(u).Omit("password_hash").Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return u, nil
}

func GetUserIdsByLinkID(linkID uuid.UUID, userID uint, db *gorm.DB) ([]uint, error) {
	var userIDs []uint
	err := db.Model(&User{}).Where("link_id = ? AND id != ?", linkID, userID).Pluck("id", &userIDs).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return userIDs, nil
}

// Follow represents a follow relationship between users
type Follow struct {
	gorm.Model
	FollowerID uint `gorm:"not null;index"` // User who is following
	FollowedID uint `gorm:"not null;index"` // User who is being followed

	// Foreign key relationships
	Follower User `gorm:"foreignKey:FollowerID;references:ID;constraint:OnDelete:CASCADE"`
	Followed User `gorm:"foreignKey:FollowedID;references:ID;constraint:OnDelete:CASCADE"`

	// Ensure unique follow relationship
	UniqueFollow string `gorm:"uniqueIndex:idx_follower_followed"`
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
		Order("users.name ASC").
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
		Order("users.name ASC").
		Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return following, nil
}
