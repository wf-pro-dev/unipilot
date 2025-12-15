package client

import (
	"encoding/json"
	"fmt"
	"time"
	"unipilot/internal/models/user"
	"unipilot/internal/secrets"

	"github.com/gofiber/fiber/v2"
)

func GetRemoteUsers() ([]user.User, error) {

	var response struct {
		Message string                   `json:"message"`
		Users   []map[string]interface{} `json:"users"`
		Error   string                   `json:"error,omitempty"`
	}

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Get(fmt.Sprintf("%s/users", api_url))

	if err := setAuthHeader(agent); err != nil {
		return nil, err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("server returned status %d: %s", statusCode, string(body))
	}

	if err := json.Unmarshal(body, &response); err != nil {
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
