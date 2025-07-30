package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type LocalCredentials struct {
	IsAuthenticated bool `json:"is_authenticated"`
	User            struct {
		UserID   uint   `json:"user_id"`
		Username string `json:"username"`
	} `json:"user"`
}

var (
	credLock    sync.Mutex
	credentials *LocalCredentials
)

func GetCurrentUser() (*LocalCredentials, error) {
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

	var creds LocalCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}

	credentials = &creds
	return credentials, nil
}

func GetCurrentUserID() (uint, error) {
	creds, err := GetCurrentUser()
	if err != nil {
		return 0, err
	}
	return creds.User.UserID, nil
}

func getCredsPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "acc-homework", "credentials.json"), nil
}

func StoreCredentials(userID uint, username string) error {
	credLock.Lock()
	defer credLock.Unlock()

	creds := LocalCredentials{
		IsAuthenticated: true,
		User: struct {
			UserID   uint   `json:"user_id"`
			Username string `json:"username"`
		}{
			UserID:   userID,
			Username: username,
		},
	}

	path, err := getCredsPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := json.Marshal(creds)
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
