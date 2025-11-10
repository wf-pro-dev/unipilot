package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"unipilot/internal/models/user"
	"unipilot/internal/secrets"
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

	new_client, err := NewAuthClient()
	if err != nil {
		return false, err
	}

	followData := map[string]uint{"followed_id": followedID}
	jsonData, _ := json.Marshal(followData)

	api_url := secrets.CONSTANTS["API_URL"]
	if err != nil {
		return false, fmt.Errorf("failed to get api url: %w", err)
	}

	resp, err := new_client.Post(fmt.Sprintf("%s/follow", api_url), "application/json", bytes.NewBuffer(jsonData))

	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	var response FollowResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
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

	new_client, err := NewAuthClient()
	if err != nil {
		return nil, 0, err
	}

	api_url := secrets.CONSTANTS["API_URL"]
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get api url: %w", err)
	}

	resp, err := new_client.Get(fmt.Sprintf("%s/followers?user_id=%d", api_url, userID))
	if err != nil {
		return nil, 0, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	var response FollowersResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, 0, err
	}

	return response.Followers, response.Total, nil
}

func GetFollowing(userID uint) ([]user.User, int, error) {

	new_client, err := NewAuthClient()
	if err != nil {
		return nil, 0, err
	}

	api_url := secrets.CONSTANTS["API_URL"]
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get api url: %w", err)
	}

	resp, err := new_client.Get(fmt.Sprintf("%s/following?user_id=%d", api_url, userID))
	if err != nil {
		return nil, 0, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	var response FollowingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, 0, err
	}

	return response.Following, response.Total, nil
}
