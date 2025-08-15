package user

import (
	"gorm.io/gorm"
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
	return count > 0, err
}

// Get followers count for a user
func GetFollowersCount(userID uint, db *gorm.DB) (int, error) {
	var count int64
	err := db.Model(&Follow{}).Where("followed_id = ?", userID).Count(&count).Error
	return int(count), err
}

// Get following count for a user
func GetFollowingCount(userID uint, db *gorm.DB) (int, error) {
	var count int64
	err := db.Model(&Follow{}).Where("follower_id = ?", userID).Count(&count).Error
	return int(count), err
}

// Create a follow relationship
func CreateFollow(followerID, followedID uint, db *gorm.DB) error {
	// Check if already following
	isFollowing, err := IsFollowing(followerID, followedID, db)
	if err != nil {
		return err
	}
	if isFollowing {
		return nil // Already following
	}

	follow := &Follow{
		FollowerID: followerID,
		FollowedID: followedID,
	}

	return db.Create(follow).Error
}

// Remove a follow relationship
func RemoveFollow(followerID, followedID uint, db *gorm.DB) error {
	return db.Where("follower_id = ? AND followed_id = ?", followerID, followedID).Delete(&Follow{}).Error
}

// Get followers list for a user
func GetFollowers(userID uint, limit, offset int, db *gorm.DB) ([]User, error) {
	var followers []User
	err := db.Joins("JOIN follows ON users.id = follows.follower_id").
		Where("follows.followed_id = ? AND follows.deleted_at is NULL", userID).
		Limit(limit).Offset(offset).
		Find(&followers).Error
	return followers, err
}

// Get following list for a user
func GetFollowing(userID uint, limit, offset int, db *gorm.DB) ([]User, error) {
	var following []User
	err := db.Joins("JOIN follows ON users.id = follows.followed_id").
		Where("follows.follower_id = ? AND follows.deleted_at is NULL", userID).
		Limit(limit).Offset(offset).
		Find(&following).Error
	return following, err
}

// Update cached follow statistics for a user
func UpdateFollowStats(userID uint, db *gorm.DB) error {
	followersCount, err := GetFollowersCount(userID, db)
	if err != nil {
		return err
	}

	followingCount, err := GetFollowingCount(userID, db)
	if err != nil {
		return err
	}

	// Upsert follow stats
	var stats FollowStats
	err = db.Where("user_id = ?", userID).First(&stats).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			stats = FollowStats{
				UserID:         userID,
				FollowersCount: followersCount,
				FollowingCount: followingCount,
			}
			return db.Create(&stats).Error
		}
		return err
	}

	stats.FollowersCount = followersCount
	stats.FollowingCount = followingCount
	return db.Save(&stats).Error
}

