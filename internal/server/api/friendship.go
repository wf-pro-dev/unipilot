package server

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/server"
	"unipilot/internal/server/sse/grpc/messages"
)

// FriendStatusResponse represents the friendship status between two users
type FriendStatusResponse struct {
	Status              *models.FriendshipStatus `json:"status"`                // Current friendship status (null if no relationship)
	IsPendingForYou     bool                     `json:"is_pending_for_you"`    // True if you need to respond to their request
	FriendsCount        int                      `json:"friends_count"`         // Number of friends the user has
	PendingRequestCount int                      `json:"pending_request_count"` // Number of pending requests for current user
	MutualFriendsCount  int                      `json:"mutual_friends_count"`  // Number of mutual friends
}

// HandleSendFriendRequest sends a friend request to another user
// Creates a pending friendship that requires acceptance from the other user.
// Implements auto-accept if there's already a pending request from the other user.
//
// HTTP Method: POST
// Path: /api/users/:id/friend-request
// Content-Type: application/json
//
// Path Parameters:
//   - id: UUID of the user to send friend request to (required)
//
// Response (200 OK):
//   - success: Boolean indicating operation success
//   - message: Descriptive message
//
// Authentication: Required (AuthMiddleware)
//
// Special Cases:
//   - If other user already sent a request, auto-accepts and creates friendship
//   - Prevents self-friending
//   - Prevents duplicate requests
//   - Allows re-requesting after rejection
//
// Error Responses:
//   - 400 Bad Request: Invalid user_id, self-request, or already friends
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 500 Internal Server Error: Database operation failure
func HandleSendFriendRequest(c *fiber.Ctx) error {
	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}

	currentUser, err := server.GetUser(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "User not found in context", fiber.StatusInternalServerError)
	}

	// Extract target user ID from path parameter
	var addresseeID string
	if addresseeID = c.Params("id"); addresseeID == "" {
		return errors.WrapServer(
			fmt.Errorf("user ID required"),
			errors.ReqParamInvalid,
			"User ID required",
			fiber.StatusBadRequest,
		)
	}

	// Prevent self-friending
	if currentUser.ID == addresseeID {
		return errors.WrapServer(
			fmt.Errorf("cannot friend yourself"),
			errors.ReqParamInvalid,
			"Cannot send friend request to yourself",
			fiber.StatusBadRequest,
		)
	}

	// Send friend request
	if err := models.SendFriendRequest(currentUser.ID, addresseeID, db); err != nil {
		if errors.HasCode(err, errors.DBConstraintViolation) {
			return errors.WrapServer(err, errors.DBConstraintViolation, err.Error(), fiber.StatusBadRequest)
		}
		return errors.WrapServer(err, errors.DBQueryFailed, "Error sending friend request", fiber.StatusInternalServerError)
	}

	go func() {
		// Send message to addressee
		if GrpcClient != nil {
			_, err := (*GrpcClient).SendMessage(context.Background(),
				&messages.Message{
					ReceiverId: addresseeID,
					SenderId:   currentUser.ID,
					Title:      "Friend Request",
					Message:    fmt.Sprintf("%s sent you a friend request", currentUser.Username),
					Type:       string(models.MessageNoContent),
					Data:       []byte(""),
				},
			)
			if err != nil {
				server.LogWarn(context.Background(), errors.WrapServer(err, errors.GRPCFailed, "Failed to send message", fiber.StatusInternalServerError))
			}
		}

	}()

	return c.SendStatus(fiber.StatusNoContent)
}

// HandleAcceptFriendRequest accepts a pending friend request
//
// HTTP Method: POST
// Path: /api/friend-requests/accept
// Content-Type: application/json
//
// Request Body:
//   - requester_id: UUID of the user who sent the request (string, required)
//
// Response (200 OK):
//   - success: Boolean indicating operation success
//   - message: "Friend request accepted"
//
// Authentication: Required (AuthMiddleware)
//
// Error Responses:
//   - 400 Bad Request: Missing requester_id or request not found
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 500 Internal Server Error: Database operation failure
func HandleAcceptFriendRequest(c *fiber.Ctx) error {
	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}
	currentUser, err := server.GetUser(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "User not found in context", fiber.StatusInternalServerError)
	}

	var requesterID string
	if requesterID = c.Params("id"); requesterID == "" {
		return errors.WrapServer(fmt.Errorf("request ID required"), errors.ReqParamMissing, "Request ID required", fiber.StatusBadRequest)
	}

	// Accept the friend request
	if err := models.AcceptFriendRequest(requesterID, currentUser.ID, db); err != nil {
		if errors.HasCode(err, errors.DBRecordNotFound) {
			return errors.WrapServer(err, errors.DBRecordNotFound, "Friend request not found", fiber.StatusBadRequest)
		}
		return errors.WrapServer(err, errors.DBQueryFailed, "Error accepting friend request", fiber.StatusInternalServerError)
	}

	go func() {
		// Send message to requester
		if GrpcClient != nil {
			_, err = (*GrpcClient).SendMessage(context.Background(),
				&messages.Message{
					ReceiverId: requesterID,
					SenderId:   currentUser.ID,
					Title:      "Friend Request Accepted",
					Message:    fmt.Sprintf("%s accepted your friend request", currentUser.Username),
					Type:       string(models.MessageNoContent),
					Data:       []byte(""),
				},
			)
		}
		if err != nil {
			server.LogWarn(context.Background(), errors.WrapServer(err, errors.GRPCFailed, "Failed to send message", fiber.StatusInternalServerError))
		}
	}()

	return c.SendStatus(fiber.StatusNoContent)
}

// HandleRejectFriendRequest rejects a pending friend request
//
// HTTP Method: POST
// Path: /api/friend-requests/reject
// Content-Type: application/json
//
// Request Body:
//   - requester_id: UUID of the user who sent the request (string, required)
//
// Response (200 OK):
//   - success: Boolean indicating operation success
//   - message: "Friend request rejected"
//
// Authentication: Required (AuthMiddleware)
func HandleRejectFriendRequest(c *fiber.Ctx) error {
	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}

	var friendshipID string
	if friendshipID = c.Params("id"); friendshipID != "" {
		return errors.WrapServer(fmt.Errorf("request ID required"), errors.ReqParamMissing, "Request ID required", fiber.StatusBadRequest)
	}

	// Reject the friend request
	if err := models.RejectFriendRequest(friendshipID, db); err != nil {
		if errors.HasCode(err, errors.DBRecordNotFound) {
			return errors.WrapServer(err, errors.DBRecordNotFound, "Friend request not found", fiber.StatusBadRequest)
		}
		return errors.WrapServer(err, errors.DBQueryFailed, "Error rejecting friend request", fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// HandleCancelFriendRequest cancels a friend request that the current user sent
//
// HTTP Method: DELETE
// Path: /api/users/:id/friend-request
//
// Path Parameters:
//   - id: UUID of the user to whom the request was sent
//
// Response (200 OK):
//   - success: Boolean indicating operation success
//   - message: "Friend request cancelled"
//
// Authentication: Required (AuthMiddleware)
func HandleCancelFriendRequest(c *fiber.Ctx) error {
	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}

	currentUser, err := server.GetUser(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "User not found in context", fiber.StatusInternalServerError)
	}

	var addresseeID string
	if addresseeID = c.Params("id"); addresseeID == "" {
		return errors.WrapServer(fmt.Errorf("user ID required"), errors.ReqParamMissing, "User ID required", fiber.StatusBadRequest)
	}

	// Cancel the friend request
	if err := models.CancelFriendRequest(currentUser.ID, addresseeID, db); err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error cancelling friend request", fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// HandleRemoveFriend removes a friendship (unfriend)
//
// HTTP Method: DELETE
// Path: /api/users/:id/friend
//
// Path Parameters:
//   - id: UUID of the friend to remove
//
// Response (200 OK):
//   - success: Boolean indicating operation success
//   - message: "Friend removed successfully"
//
// Authentication: Required (AuthMiddleware)
func HandleRemoveFriend(c *fiber.Ctx) error {
	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}

	currentUser, err := server.GetUser(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "User not found in context", fiber.StatusInternalServerError)
	}

	// Parse UUID
	var friendID string
	if friendID = c.Params("id"); friendID == "" {
		return errors.WrapServer(
			err,
			errors.ReqParamInvalid,
			"Invalid user ID format",
			fiber.StatusBadRequest,
		)
	}

	// Remove the friendship
	if err := models.RemoveFriend(currentUser.ID, friendID, db); err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error removing friend", fiber.StatusInternalServerError)
	}

	// Clear cache for both users
	clearCtx := context.Background()
	if err := CacheService.DeleteUserFriends(clearCtx, currentUser.ID); err != nil {
		server.LogWarn(clearCtx, errors.WrapServer(err, errors.CacheOperationFailed, "Failed to clear friends cache", fiber.StatusInternalServerError))
	}
	if err := CacheService.DeleteUserFriends(clearCtx, friendID); err != nil {
		server.LogWarn(clearCtx, errors.WrapServer(err, errors.CacheOperationFailed, "Failed to clear friends cache", fiber.StatusInternalServerError))
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// HandleGetFriends retrieves a paginated list of friends for a user
// Implements Redis caching with database fallback
//
// HTTP Method: GET
// Path: /api/users/:id/friends
//
// Path Parameters:
//   - id: UUID of the user whose friends to retrieve
//
// Query Parameters:
//   - limit: Maximum number of friends to return (default: 20)
//   - offset: Number of friends to skip for pagination (default: 0)
//
// Response (200 OK):
//   - friends: Array of user objects representing friends
//   - total: Total number of friends
//
// Authentication: Required (AuthMiddleware)
func HandleGetFriends(c *fiber.Ctx) error {
	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}

	// Parse UUID
	var userID string
	if userID = c.Params("id"); userID == "" {
		log.Println("User ID required:", userID)
		return errors.WrapServer(
			fmt.Errorf("user ID required"),
			errors.ReqParamMissing,
			"User ID required",
			fiber.StatusBadRequest,
		)
	}

	// Parse pagination parameters
	limit := 20

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		} else {
			return errors.WrapServer(fmt.Errorf("error limit is in"), errors.ReqParamInvalid, "Error converting limit to int", fiber.StatusBadRequest)
		}
	}

	var cursor *models.Cursor
	if err := c.BodyParser(&cursor); err != nil {
		return errors.WrapServer(err, errors.ReqParamInvalid, "Error unmarshalling cursor", fiber.StatusBadRequest)
	}

	results, err := models.GetFriends(userID, cursor, limit, db)
	if err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting friends from database", fiber.StatusInternalServerError)
	}

	return c.JSON(results)
}

// HandleGetPendingRequests retrieves pending friend requests for the current user
//
// HTTP Method: GET
// Path: /api/friend-requests/pending
//
// Query Parameters:
//   - limit: Maximum number of requests to return (default: 20)
//   - offset: Number of requests to skip for pagination (default: 0)
//
// Response (200 OK):
//   - requests: Array of user objects who sent friend requests
//   - total: Total number of pending requests
//
// Authentication: Required (AuthMiddleware)
func HandleGetPendingRequests(c *fiber.Ctx) error {
	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}

	currentUser, err := server.GetUser(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "User not found in context", fiber.StatusInternalServerError)
	}

	// Parse pagination parameters
	limit := 20
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get pending requests
	requests, err := models.GetPendingRequests(currentUser.ID, limit, offset, db)
	if err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting pending requests", fiber.StatusInternalServerError)
	}

	return c.JSON(requests)
}

// HandleGetSentRequests retrieves friend requests sent by the current user
//
// HTTP Method: GET
// Path: /api/friend-requests/sent
//
// Query Parameters:
//   - limit: Maximum number of requests to return (default: 20)
//   - offset: Number of requests to skip for pagination (default: 0)
//
// Response (200 OK):
//   - requests: Array of user objects to whom requests were sent
//   - total: Total number of sent requests
//
// Authentication: Required (AuthMiddleware)
func HandleGetSentRequests(c *fiber.Ctx) error {
	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}

	currentUser, err := server.GetUser(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "User not found in context", fiber.StatusInternalServerError)
	}

	// Parse pagination parameters
	limit := 20
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get sent requests
	requests, err := models.GetSentRequests(currentUser.ID, limit, offset, db)
	if err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting sent requests", fiber.StatusInternalServerError)
	}

	return c.JSON(requests)
}

// HandleGetFriendStatus retrieves the friendship status between current user and another user
//
// HTTP Method: GET
// Path: /api/users/:id/friend-status
//
// Path Parameters:
//   - id: UUID of the other user
//
// Response (200 OK):
//   - status: Current friendship status (null if no relationship)
//   - is_pending_for_you: Whether you need to respond to their request
//   - friends_count: Number of friends the other user has
//   - pending_request_count: Number of pending requests for you
//   - mutual_friends_count: Number of mutual friends
//
// Authentication: Required (AuthMiddleware)
func HandleGetFriendStatus(c *fiber.Ctx) error {
	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}

	currentUser, err := server.GetUser(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "User not found in context", fiber.StatusInternalServerError)
	}

	// Parse UUID
	var otherUserID string
	if otherUserID = c.Params("id"); otherUserID == "" {
		return errors.WrapServer(
			err,
			errors.ReqParamInvalid,
			"Invalid user ID format",
			fiber.StatusBadRequest,
		)
	}

	// Get friendship status
	status, isPending, err := models.GetFriendshipStatus(currentUser.ID, otherUserID, db)
	if err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting friendship status", fiber.StatusInternalServerError)
	}

	// Get friends count for the other user
	friendsCount, err := models.GetFriendsCount(otherUserID, db)
	if err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting friends count", fiber.StatusInternalServerError)
	}

	// Get pending requests count for current user
	pendingCount, err := models.GetPendingRequestsCount(currentUser.ID, db)
	if err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting pending requests count", fiber.StatusInternalServerError)
	}

	// Get mutual friends count
	mutualFriends, err := models.GetMutualFriends(currentUser.ID, otherUserID, db)
	if err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error getting mutual friends", fiber.StatusInternalServerError)
	}

	response := FriendStatusResponse{
		Status:              status,
		IsPendingForYou:     isPending,
		FriendsCount:        friendsCount,
		PendingRequestCount: pendingCount,
		MutualFriendsCount:  len(mutualFriends),
	}

	return c.JSON(response)
}

// HandleBlockUser blocks another user
//
// HTTP Method: POST
// Path: /api/users/:id/block
//
// Path Parameters:
//   - id: UUID of the user to block
//
// Response (200 OK):
//   - success: Boolean indicating operation success
//   - message: "User blocked successfully"
//
// Authentication: Required (AuthMiddleware)
func HandleBlockUser(c *fiber.Ctx) error {
	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}

	currentUser, err := server.GetUser(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "User not found in context", fiber.StatusInternalServerError)
	}
	// Parse UUID
	var blockedID string
	if blockedID = c.Params("id"); blockedID != "" {
		return errors.WrapServer(
			err,
			errors.ReqParamInvalid,
			"Invalid user ID format",
			fiber.StatusBadRequest,
		)
	}

	if currentUser.ID == blockedID {
		return errors.WrapServer(
			fmt.Errorf("cannot block yourself"),
			errors.ReqParamInvalid,
			"Cannot block yourself",
			fiber.StatusBadRequest,
		)
	}

	// Block the user
	if err := models.BlockUser(currentUser.ID, blockedID, db); err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error blocking user", fiber.StatusInternalServerError)
	}

	// Clear cache
	clearCtx := context.Background()
	if err := CacheService.DeleteUserFriends(clearCtx, currentUser.ID); err != nil {
		server.LogWarn(clearCtx, errors.WrapServer(err, errors.CacheOperationFailed, "Failed to clear friends cache", fiber.StatusInternalServerError))
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// HandleUnblockUser unblocks a user
//
// HTTP Method: DELETE
// Path: /api/users/:id/block
//
// Path Parameters:
//   - id: UUID of the user to unblock
//
// Response (200 OK):
//   - success: Boolean indicating operation success
//   - message: "User unblocked successfully"
//
// Authentication: Required (AuthMiddleware)
func HandleUnblockUser(c *fiber.Ctx) error {
	ctx := c.UserContext()
	db, err := server.GetDB(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "DB not found in context", fiber.StatusInternalServerError)
	}

	currentUser, err := server.GetUser(ctx)
	if err != nil {
		return errors.WrapServer(err, errors.InternalError, "User not found in context", fiber.StatusInternalServerError)
	}

	// Parse UUID
	var blockedID string
	if blockedID = c.Params("id"); blockedID != "" {
		return errors.WrapServer(
			err,
			errors.ReqParamInvalid,
			"Invalid user ID format",
			fiber.StatusBadRequest,
		)
	}

	// Unblock the user
	if err := models.UnblockUser(currentUser.ID, blockedID, db); err != nil {
		return errors.WrapServer(err, errors.DBQueryFailed, "Error unblocking user", fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
