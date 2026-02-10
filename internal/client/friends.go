package client

import (
	"encoding/json"
	"fmt"
	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/secrets"

	"github.com/gofiber/fiber/v2"
)

type FriendStatusResponse struct {
	Status              *models.FriendshipStatus `json:"status"`                // Current friendship status (null if no relationship)
	IsPendingForYou     bool                     `json:"is_pending_for_you"`    // True if you need to respond to their request
	FriendsCount        int                      `json:"friends_count"`         // Number of friends the user has
	PendingRequestCount int                      `json:"pending_request_count"` // Number of pending requests for current user
	MutualFriendsCount  int                      `json:"mutual_friends_count"`  // Number of mutual friends
	CoursesCount        int                      `json:"courses_count"`         // Number of courses the user has
}

func GetFriends(userID string, cursor *models.Cursor, limit int) (*models.PageResponse[models.User], error) {

	fmt.Println("[Client] Getting friends for user:", userID)

	api_url := secrets.CONSTANTS["API_URL"]
	agent, err := GetAuthAgent(fiber.Get(fmt.Sprintf("%s/users/%s/friends?limit=%d", api_url, userID, limit)).JSON(cursor))
	if err != nil {
		return nil, err
	}

	statusCode, body, _ := agent.Bytes()
	if statusCode != fiber.StatusOK {
		serverError := errors.ParseServerError(body, statusCode)
		return nil, serverError
	}

	var response *models.PageResponse[models.User]
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
	}

	return response, nil
}

func GetFriendShipStatus(userID string) (*FriendStatusResponse, error) {
	api_url := secrets.CONSTANTS["API_URL"]

	agent, err := GetAuthAgent(fiber.Get(fmt.Sprintf("%s/users/%s/friend-status", api_url, userID)))
	if err != nil {
		return nil, err
	}

	statusCode, body, _ := agent.Bytes()
	if statusCode != fiber.StatusOK {
		var serverError *errors.ServerError
		serverError = errors.ParseServerError(body, statusCode)
		return nil, serverError
	}

	var response FriendStatusResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
	}

	return &response, nil

}

func SendFriendRequest(userID string) error {
	api_url := secrets.CONSTANTS["API_URL"]

	agent, err := GetAuthAgent(fiber.Post(fmt.Sprintf("%s/users/%s/friend-request", api_url, userID)))
	if err != nil {
		return err
	}

	statusCode, body, _ := agent.Bytes()
	if statusCode != fiber.StatusNoContent {
		serverError := errors.ParseServerError(body, statusCode)
		return serverError
	}

	return nil
}

func AcceptFriendRequest(userID string) error {
	api_url := secrets.CONSTANTS["API_URL"]

	agent, err := GetAuthAgent(fiber.Post(fmt.Sprintf("%s/users/%s/friend-request/accept", api_url, userID)))
	if err != nil {
		return err
	}

	statusCode, body, _ := agent.Bytes()
	if statusCode != fiber.StatusNoContent {
		serverError := errors.ParseServerError(body, statusCode)
		return serverError
	}

	return nil
}

func CancelFriendRequest(userID string) error {
	api_url := secrets.CONSTANTS["API_URL"]

	agent, err := GetAuthAgent(fiber.Delete(fmt.Sprintf("%s/users/%s/friend-request", api_url, userID)))
	if err != nil {
		return err
	}

	statusCode, body, _ := agent.Bytes()
	if statusCode != fiber.StatusNoContent {
		serverError := errors.ParseServerError(body, statusCode)
		return serverError
	}

	return nil
}

func RemoveFriend(userID string) error {
	api_url := secrets.CONSTANTS["API_URL"]

	agent, err := GetAuthAgent(fiber.Delete(fmt.Sprintf("%s/users/%s/friend", api_url, userID)))
	if err != nil {
		return err
	}

	statusCode, body, _ := agent.Bytes()
	if statusCode != fiber.StatusNoContent {
		serverError := errors.ParseServerError(body, statusCode)
		return serverError
	}

	return nil
}
