package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"unipilot/internal/models/user"
	"unipilot/internal/secrets"
)

func GetRemoteUsers() ([]user.User, error) {

	var response struct {
		Message string                   `json:"message"`
		Users   []map[string]interface{} `json:"users"`
		Error   string                   `json:"error,omitempty"`
	}

	new_client, err := NewAuthClient()
	if err != nil {
		return nil, err
	}

	api_url, err := secrets.GetEnvVar("API_URL")
	if err != nil {
		return nil, fmt.Errorf("failed to get api url: %w", err)
	}

	resp, err := new_client.Get(fmt.Sprintf("%s/users", api_url))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	var users []user.User

	for _, u := range response.Users {
		user_courses := u["courses_codes"].([]interface{})
		user_courses_strings := make([]string, len(user_courses))
		for i, course := range user_courses {
			user_courses_strings[i] = course.(string)
		}

		user := user.User{
			Username:    u["username"].(string),
			Email:       u["email"].(string),
			Avatar:      u["avatar"].(string),
			University:  u["university"].(string),
			CoursesCode: user_courses_strings,
		}

		user.ID = uint(u["id"].(float64))
		user.CreatedAt, _ = time.Parse(time.RFC3339, u["created_at"].(string))
		users = append(users, user)
	}

	return users, nil

}
