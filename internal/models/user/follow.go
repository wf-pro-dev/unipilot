package user

import (
	Errors "errors"

	"gorm.io/gorm"

	"unipilot/internal/errors"
)

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

// FollowStats represents cached follow statistics for a user
type FollowStats struct {
	gorm.Model
	UserID         uint `gorm:"unique;not null"`
	FollowersCount int  `gorm:"default:0"`
	FollowingCount int  `gorm:"default:0"`

	// Foreign key relationship
	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
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
	return errors.HandleDBWriteError(err)
}

// Get followers list for a user
func GetFollowers(userID uint, limit, offset int, db *gorm.DB) ([]User, error) {
	var followers []User
	err := db.Joins("JOIN follows ON users.id = follows.follower_id").
		Where("follows.followed_id = ? AND follows.deleted_at is NULL", userID).
		Limit(limit).Offset(offset).
		Find(&followers).Error
	return followers, errors.HandleDBReadError(err)
}

// Get following list for a user
func GetFollowing(userID uint, limit, offset int, db *gorm.DB) ([]User, error) {
	var following []User
	err := db.Joins("JOIN follows ON users.id = follows.followed_id").
		Where("follows.follower_id = ? AND follows.deleted_at is NULL", userID).
		Limit(limit).Offset(offset).
		Find(&following).Error
	return following, errors.HandleDBReadError(err)
}

// Update cached follow statistics for a user
func UpdateFollowStats(userID uint, db *gorm.DB) error {
	followersCount, err := GetFollowersCount(userID, db)
	if err != nil {
		return errors.HandleDBReadError(err)
	}

	followingCount, err := GetFollowingCount(userID, db)
	if err != nil {
		return errors.HandleDBReadError(err)
	}

	// Upsert follow stats
	var stats FollowStats
	err = db.Where("user_id = ?", userID).First(&stats).Error
	if err != nil {
		if Errors.Is(err, gorm.ErrRecordNotFound) {
			stats = FollowStats{
				UserID:         userID,
				FollowersCount: followersCount,
				FollowingCount: followingCount,
			}
			err = db.Create(&stats).Error
			if err != nil {
				return errors.HandleDBCreateError(err)
			}
			return nil
		}
		return errors.HandleDBReadError(err)
	}

	stats.FollowersCount = followersCount
	stats.FollowingCount = followingCount
	err = db.Save(&stats).Error
	if err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}
