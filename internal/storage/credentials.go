package storage

import (
	"encoding/json"

	"os"
	"path/filepath"
	"sync"
)

type LocalCredentials struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
}

var (
	credLock    sync.Mutex
	credentials *LocalCredentials
)

func GetCurrentUserID() (uint, error) {
	credLock.Lock()
	defer credLock.Unlock()

	if credentials != nil {
		return credentials.UserID, nil
	}

	path, err := getCredsPath()
	if err != nil {
		return 0, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	var creds LocalCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return 0, err
	}

	credentials = &creds
	return creds.UserID, nil
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
		UserID:   userID,
		Username: username,
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
