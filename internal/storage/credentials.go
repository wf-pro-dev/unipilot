package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"unipilot/internal/models/user"
)

var (
	credLock    sync.Mutex
	credentials *user.User
)

func GetCurrentUser() (*user.User, error) {
	credLock.Lock()
	defer credLock.Unlock()

	if credentials != nil {
		return credentials, nil
	}

	path, err := getCredsPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var user user.User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}

	return &user, nil
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

	data, err := json.Marshal(userData)
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

	credentials = nil
	return os.Remove(path)
}
