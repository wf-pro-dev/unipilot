package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
	"unipilot/internal/models/user"
)

var (
	credLock sync.Mutex
)

func GetCurrentUser() (*user.User, error) {
	credLock.Lock()
	defer credLock.Unlock()

	path, err := getCredsPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var response struct {
		ID         uint   `json:"id"`
		Username   string `json:"username"`
		Email      string `json:"email"`
		Avatar     string `json:"avatar"`
		University string `json:"university"`
		Semester   string `json:"semester"`
		Year       string `json:"year"`
		Language   string `json:"language"`
		CreatedAt  string `json:"created_at"`
		UpdatedAt  string `json:"updated_at"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	currentUser := &user.User{
		Username:   response.Username,
		Email:      response.Email,
		Avatar:     response.Avatar,
		University: response.University,
		Semester:   response.Semester,
		Year:       response.Year,
		Language:   response.Language,
	}

	createdAt, err := time.Parse(time.RFC3339, response.CreatedAt)
	if err != nil {
		return nil, err
	}
	updatedAt, err := time.Parse(time.RFC3339, response.UpdatedAt)
	if err != nil {
		return nil, err
	}

	currentUser.CreatedAt = createdAt
	currentUser.UpdatedAt = updatedAt
	currentUser.ID = response.ID

	return currentUser, nil
}

func GetCurrentUserID() (uint, error) {
	user, err := GetCurrentUser()
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

func getCredsPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "acc-homework", "credentials.json"), nil
}

func StoreCredentials(user user.User) error {
	credLock.Lock()
	defer credLock.Unlock()

	userData := user

	path, err := getCredsPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := json.Marshal(userData.ToMap())
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func ClearCredentials() error {
	credLock.Lock()
	defer credLock.Unlock()

	path, err := getCredsPath()
	if err != nil {
		return err
	}

	return os.Remove(path)
}
