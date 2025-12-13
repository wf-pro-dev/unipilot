package client

import (
	"encoding/json"
	"fmt"
	"unipilot/internal/models/user"
	"unipilot/internal/secrets"

	"github.com/gofiber/fiber/v2"
)

// FollowRequest represents a follow request
type FollowRequest struct {
	FollowedID uint `json:"followed_id"`
}

// FollowResponse represents a follow response
type FollowResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func Follow(followedID uint) (bool, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Post(fmt.Sprintf("%s/users/%d/follow", api_url, followedID))

	if err := setAuthHeader(agent); err != nil {
		return false, err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return false, errs[0]
	}

	if statusCode != 200 {
		return false, fmt.Errorf("server returned status %d: %s", statusCode, string(body))
	}

	var response FollowResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return false, err
	}

	return response.Success, nil

}

// FollowersResponse represents a followers list response
type FollowersResponse struct {
	Followers []user.User `json:"followers"`
	Total     int         `json:"total"`
}

// FollowingResponse represents a following list response
type FollowingResponse struct {
	Following []user.User `json:"following"`
	Total     int         `json:"total"`
}

func GetFollowers(userID uint) ([]user.User, int, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Get(fmt.Sprintf("%s/users/%d/followers", api_url, userID))

	if err := setAuthHeader(agent); err != nil {
		return nil, 0, err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, 0, errs[0]
	}

	if statusCode != 200 {
		return nil, 0, fmt.Errorf("server returned status %d: %s", statusCode, string(body))
	}

	var response FollowersResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, 0, err
	}

	return response.Followers, response.Total, nil
}

func GetFollowing(userID uint) ([]user.User, int, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Get(fmt.Sprintf("%s/users/%d/following", api_url, userID))

	if err := setAuthHeader(agent); err != nil {
		return nil, 0, err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, 0, errs[0]
	}

	if statusCode != 200 {
		return nil, 0, fmt.Errorf("server returned status %d: %s", statusCode, string(body))
	}

	var response FollowingResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, 0, err
	}

	return response.Following, response.Total, nil
}
