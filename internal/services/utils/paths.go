package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	appName = "unipilot"
)

func getMainDir() (string, error) {

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	fileDir := filepath.Join(configDir, appName)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(fileDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create app data directory: %w", err)
	}
	return fileDir, nil
}

func getUserDir() (string, error) {

	credentials, err := GetUserFromFile()
	if err != nil {
		return "", fmt.Errorf("failed to get user from file: %w", err)
	}

	fileDir, err := getMainDir()
	if err != nil {
		return "", err
	}
	userDir := filepath.Join(fileDir, fmt.Sprintf("user_%d", credentials.ID))

	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create document directory: %w", err)
	}

	return userDir, nil
}

func getUserDirWithID(userID uint) (string, error) {

	fileDir, err := getMainDir()
	if err != nil {
		return "", err
	}

	userDir := filepath.Join(fileDir, fmt.Sprintf("user_%d", userID))

	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create document directory: %w", err)
	}

	return userDir, nil
}

func GetDocumentDir() (string, error) {

	fileDir, err := getUserDir()
	if err != nil {
		return "", err
	}

	documentDir := filepath.Join(fileDir, "documents")
	// Create directory if it doesn't exist
	if err := os.MkdirAll(documentDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create document directory: %w", err)
	}

	return documentDir, nil
}

func GetDBPath() (string, error) {

	fileDir, err := getUserDir()
	if err != nil {
		return "", err
	}

	dbPath := filepath.Join(fileDir, "data.db")

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_, err := os.Create(dbPath)
		if err != nil {
			return "", err
		}
	}

	return dbPath, nil
}

func GetDBPathWithID(userID uint) (string, error) {

	fileDir, err := getUserDirWithID(userID)
	if err != nil {
		return "", err
	}

	dbPath := filepath.Join(fileDir, "data.db")

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_, err := os.Create(dbPath)
		if err != nil {
			return "", err
		}
	}

	return dbPath, nil
}

// getCookieFilePath returns the canonical path for the cookie file.
func GetCookieFilePath() (string, error) {

	fileDir, err := getUserDir()
	if err != nil {
		return "", err
	}
	cookiePath := filepath.Join(fileDir, "cookies.txt")

	return cookiePath, nil
}

func GetCredentialFile() (string, error) {

	fileDir, err := getMainDir()
	if err != nil {
		return "", err
	}
	credentialsPath := filepath.Join(fileDir, "credentials.json")
	if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
		_, err := os.Create(credentialsPath)
		if err != nil {
			return "", err
		}
	}
	return credentialsPath, nil
}
