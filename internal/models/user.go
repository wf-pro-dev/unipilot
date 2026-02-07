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
	Base
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

	SentFriendRequests     []*Friendship `gorm:"foreignKey:RequesterID;references:ID"`
	ReceivedFriendRequests []*Friendship `gorm:"foreignKey:AddresseeID;references:ID"`
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

func GetUser(id string, db *gorm.DB) (*User, error) {
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
	UserID string `gorm:"column:user_id"`
	Code   string `gorm:"column:code"`
}

func GetUsersCourseCodes(userIDs []string, db *gorm.DB) ([]UserCourseCodes, error) {
	var courseCodes []UserCourseCodes
	if err := db.Model(&Course{}).
		Select("user_id, code").
		Where("user_id IN ? AND deleted_at IS NULL", userIDs).
		Find(&courseCodes).Error; err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return courseCodes, nil
}

func GetCourseUsers(courseID string, db *gorm.DB) ([]string, error) {
	var userIDs []string

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

func GetUserClusterIDs(userID string, db *gorm.DB) ([]string, error) {
	var clusterIDs []string

	// The CourseID in the invitation is ALWAYS the Root IDs
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

// FriendshipStatus represents the state of a friendship
type FriendshipStatus string

const (
	FriendshipPending  FriendshipStatus = "pending"
	FriendshipAccepted FriendshipStatus = "accepted"
	FriendshipRejected FriendshipStatus = "rejected"
	FriendshipBlocked  FriendshipStatus = "blocked"
)

// Friendship represents a bidirectional friendship between users
// Uses a request/accept model where one user initiates and another responds
type Friendship struct {
	Base
	RequesterID string           `gorm:"not null;index:idx_friendship_users"` // User who sent the friend request
	AddresseeID string           `gorm:"not null;index:idx_friendship_users"` // User who received the friend request
	Status      FriendshipStatus `gorm:"type:varchar(20);not null;default:'pending';index:idx_friendship_status"`

	// Foreign key relationships with cascade delete
	Requester User `gorm:"foreignKey:RequesterID;references:ID;constraint:OnDelete:CASCADE"`
	Addressee User `gorm:"foreignKey:AddresseeID;references:ID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for the Friendship model
func (Friendship) TableName() string {
	return "friendships"
}

// BeforeCreate hook to ensure requester_id < addressee_id for consistency
// This prevents duplicate friendships (A->B and B->A)
func (f *Friendship) BeforeCreate(tx *gorm.DB) error {

	if err := f.Base.BeforeCreate(tx); err != nil {
		return err
	}
	// Ensure requesterID != addresseeID
	if f.RequesterID == f.AddresseeID {
		return errors.NewAppError(errors.ValidationInvalid, "Cannot befriend yourself", nil)
	}
	return nil
}

func GetFriendshipByID(friendshipID string, db *gorm.DB) (*Friendship, error) {
	var friendship Friendship
	err := db.Where("id = ?", friendshipID).First(&friendship).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return &friendship, nil
}

// GetFriendship retrieves a friendship between two users (regardless of who initiated)
func GetFriendship(userID1, userID2 string, db *gorm.DB) (*Friendship, error) {
	var friendship Friendship
	err := db.Where(
		"(requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?)",
		userID1, userID2, userID2, userID1,
	).First(&friendship).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.HandleDBReadError(err)
	}
	return &friendship, nil
}

// AreFriends checks if two users are friends (accepted status)
func AreFriends(userID1, userID2 string, db *gorm.DB) (bool, error) {
	var count int64
	err := db.Model(&Friendship{}).Where(
		"((requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?)) AND status = ?",
		userID1, userID2, userID2, userID1, FriendshipAccepted,
	).Count(&count).Error

	if err != nil {
		return false, errors.HandleDBReadError(err)
	}
	return count > 0, nil
}

// GetFriendshipStatus returns the status between two users
// Returns: status, isPending (waiting for current user to respond), error
func GetFriendshipStatus(currentUserID, otherUserID string, db *gorm.DB) (*FriendshipStatus, bool, error) {
	friendship, err := GetFriendship(currentUserID, otherUserID, db)
	if err != nil {
		return nil, false, err
	}
	if friendship == nil {
		return nil, false, nil
	}

	// Check if current user needs to respond to a pending request
	isPendingForCurrentUser := friendship.Status == FriendshipPending && friendship.AddresseeID == currentUserID

	return &friendship.Status, isPendingForCurrentUser, nil
}

// GetFriendsCount returns the number of accepted friends for a user
func GetFriendsCount(userID string, db *gorm.DB) (int, error) {
	var count int64
	err := db.Model(&Friendship{}).Where(
		"(requester_id = ? OR addressee_id = ?) AND status = ?",
		userID, userID, FriendshipAccepted,
	).Count(&count).Error

	if err != nil {
		return 0, errors.HandleDBReadError(err)
	}
	return int(count), nil
}

// GetPendingRequestsCount returns the number of pending friend requests for a user
func GetPendingRequestsCount(userID string, db *gorm.DB) (int, error) {
	var count int64
	err := db.Model(&Friendship{}).Where(
		"addressee_id = ? AND status = ?",
		userID, FriendshipPending,
	).Count(&count).Error

	if err != nil {
		return 0, errors.HandleDBReadError(err)
	}
	return int(count), nil
}

// SendFriendRequest creates a new friend request
func SendFriendRequest(requesterID, addresseeID string, db *gorm.DB) error {
	// Check if a friendship already exists
	existing, err := GetFriendship(requesterID, addresseeID, db)
	if err != nil {
		return err
	}

	if existing != nil {
		switch existing.Status {
		case FriendshipAccepted:
			return errors.NewAppError(errors.DBConstraintViolation, "Already friends", nil)
		case FriendshipPending:
			if existing.RequesterID == requesterID {
				return errors.NewAppError(errors.DBConstraintViolation, "Friend request already sent", nil)
			}
			// If there's a pending request from the other user, auto-accept it
			existing.Status = FriendshipAccepted
			if err := db.Save(existing).Error; err != nil {
				return errors.HandleDBWriteError(err)
			}
			return nil
		case FriendshipRejected:
			// Allow re-requesting after rejection (update existing record)
			existing.RequesterID = requesterID
			existing.AddresseeID = addresseeID
			existing.Status = FriendshipPending
			if err := db.Save(existing).Error; err != nil {
				return errors.HandleDBWriteError(err)
			}
			return nil
		case FriendshipBlocked:
			return errors.NewAppError(errors.DBConstraintViolation, "Cannot send friend request", nil)
		}
	}

	// Create new friend request
	friendship := &Friendship{
		RequesterID: requesterID,
		AddresseeID: addresseeID,
		Status:      FriendshipPending,
	}

	if err := db.Create(friendship).Error; err != nil {
		return errors.HandleDBCreateError(err)
	}

	return nil
}

// AcceptFriendRequest accepts a pending friend request
func AcceptFriendRequest(requesterID, addresseeID string, db *gorm.DB) error {

	if err := db.Model(&Friendship{}).
		Where("requester_id = ? AND addressee_id = ?", requesterID, addresseeID).
		Update("status", FriendshipAccepted).Error; err != nil {
		return errors.HandleDBWriteError(err)
	}

	return nil
}

// RejectFriendRequest rejects a pending friend request
func RejectFriendRequest(friendshipID string, db *gorm.DB) error {
	var friendship Friendship
	err := db.Where("id = ?", friendshipID).First(&friendship).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewAppError(errors.DBRecordNotFound, "Friend request not found", nil)
		}
		return errors.HandleDBReadError(err)
	}

	friendship.Status = FriendshipRejected
	if err := db.Save(&friendship).Error; err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}

// RemoveFriend removes a friendship (unfriend)
func RemoveFriend(userID1, userID2 string, db *gorm.DB) error {
	err := db.Where(
		"((requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?)) AND status = ?",
		userID1, userID2, userID2, userID1, FriendshipAccepted,
	).Delete(&Friendship{}).Error

	if err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}

// CancelFriendRequest cancels a pending friend request that the current user sent
func CancelFriendRequest(requesterID, addresseeID string, db *gorm.DB) error {
	err := db.Where("requester_id = ? AND addressee_id = ? AND status = ?",
		requesterID, addresseeID, FriendshipPending).Delete(&Friendship{}).Error

	if err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}

// BlockUser blocks another user
func BlockUser(blockerID, blockedID string, db *gorm.DB) error {
	// Remove any existing friendship
	existing, err := GetFriendship(blockerID, blockedID, db)
	if err != nil {
		return err
	}

	if existing != nil {
		existing.RequesterID = blockerID
		existing.AddresseeID = blockedID
		existing.Status = FriendshipBlocked
		if err := db.FirstOrCreate(existing).Error; err != nil {
			return errors.HandleDBWriteError(err)
		}
	} else {
		friendship := &Friendship{
			RequesterID: blockerID,
			AddresseeID: blockedID,
			Status:      FriendshipBlocked,
		}
		if err := db.Create(friendship).Error; err != nil {
			return errors.HandleDBCreateError(err)
		}
	}
	return nil
}

// UnblockUser unblocks a user
func UnblockUser(blockerID, blockedID string, db *gorm.DB) error {
	err := db.Where("requester_id = ? AND addressee_id = ? AND status = ?",
		blockerID, blockedID, FriendshipBlocked).Delete(&Friendship{}).Error

	if err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}

// GetFriends retrieves the list of accepted friends for a user
func GetFriends(userID string, cursor *Cursor, limit int, db *gorm.DB) (*PageResponse[User], error) {
	var friends []User

	// Get friendships where user is either requester or addressee with accepted status
	query := db.Table("users").
		Select("users.*").
		Joins("JOIN friendships ON (friendships.requester_id = users.id OR friendships.addressee_id = users.id)").
		Where("(friendships.requester_id = ? OR friendships.addressee_id = ?) AND friendships.status = ? AND users.id != ? AND friendships.deleted_at IS NULL",
			userID, userID, FriendshipAccepted, userID).
		Order("users.username ASC").
		Limit(limit + 1)

	if cursor != nil {
		query = query.Where(
			"(friendships.created_at > ?) OR (friendships.created_at = ? AND friendships.id > ?)",
			cursor.CreatedAt, cursor.CreatedAt, cursor.ID,
		)
	}

	if err := query.Scan(&friends).Error; err != nil {
		return nil, err
	}

	hasMore := len(friends) > limit
	if hasMore {
		friends = friends[:limit]
	}

	var nextCursor *Cursor
	if hasMore && len(friends) > 0 {
		last := friends[len(friends)-1]
		nextCursor = &Cursor{
			CreatedAt: last.UpdatedAt,
			ID:        last.ID,
		}
	}

	return &PageResponse[User]{
		Data:    friends,
		Cursor:  nextCursor,
		HasMore: hasMore,
	}, nil
}

// GetPendingRequests retrieves pending friend requests received by the user
func GetPendingRequests(userID string, limit, offset int, db *gorm.DB) ([]User, error) {
	var users []User

	err := db.Table("users").
		Select("users.*").
		Joins("JOIN friendships ON friendships.requester_id = users.id").
		Where("friendships.addressee_id = ? AND friendships.status = ? AND friendships.deleted_at IS NULL",
			userID, FriendshipPending).
		Order("friendships.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error

	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return users, nil
}

// GetSentRequests retrieves pending friend requests sent by the user
func GetSentRequests(userID string, limit, offset int, db *gorm.DB) ([]User, error) {
	var users []User

	err := db.Table("users").
		Select("users.*").
		Joins("JOIN friendships ON friendships.addressee_id = users.id").
		Where("friendships.requester_id = ? AND friendships.status = ? AND friendships.deleted_at IS NULL",
			userID, FriendshipPending).
		Order("friendships.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error

	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return users, nil
}

// GetMutualFriends retrieves mutual friends between two users
func GetMutualFriends(userID1, userID2 string, db *gorm.DB) ([]User, error) {
	var mutualFriends []User

	// Subquery to get user1's friends
	user1FriendsQuery := db.Table("friendships").
		Select("CASE WHEN requester_id = ? THEN addressee_id ELSE requester_id END as friend_id", userID1).
		Where("(requester_id = ? OR addressee_id = ?) AND status = ?", userID1, userID1, FriendshipAccepted)

	// Subquery to get user2's friends
	user2FriendsQuery := db.Table("friendships").
		Select("CASE WHEN requester_id = ? THEN addressee_id ELSE requester_id END as friend_id", userID2).
		Where("(requester_id = ? OR addressee_id = ?) AND status = ?", userID2, userID2, FriendshipAccepted)

	// Find intersection
	err := db.Table("users").
		Select("users.*").
		Joins("JOIN (?) AS user1_friends ON users.id = user1_friends.friend_id", user1FriendsQuery).
		Joins("JOIN (?) AS user2_friends ON users.id = user2_friends.friend_id", user2FriendsQuery).
		Order("users.username ASC").
		Find(&mutualFriends).Error

	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return mutualFriends, nil
}

func GetUserCourseInvitations(userID string, db *gorm.DB) ([]CourseInvitation, error) {
	var invitations []CourseInvitation
	err := db.Preload("Course").Where("receiver_id = ? OR owner_id = ?", userID, userID).Find(&invitations).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return invitations, nil
}
