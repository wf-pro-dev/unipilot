package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"unipilot/internal/models/user"
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

	new_client, err := NewClientWithCookies()
	if err != nil {
		return false, err
	}

	followData := map[string]uint{"followed_id": followedID}
	jsonData, _ := json.Marshal(followData)

	resp, err := new_client.Post("https://newsroom.dedyn.io/acc-homework/follow", "application/json", bytes.NewBuffer(jsonData))

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

	new_client, err := NewClientWithCookies()
	if err != nil {
		return nil, 0, err
	}

	resp, err := new_client.Get(fmt.Sprintf("https://newsroom.dedyn.io/acc-homework/followers?user_id=%d", userID))

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

	new_client, err := NewClientWithCookies()
	if err != nil {
		return nil, 0, err
	}

	resp, err := new_client.Get(fmt.Sprintf("https://newsroom.dedyn.io/acc-homework/following?user_id=%d", userID))

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
