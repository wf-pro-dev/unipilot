package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"unipilot/internal/errors"
)

const (
	appName = "unipilot"
)

func getMainDir() (string, error) {

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", errors.Wrap(err, errors.FSDirFailed, "Failed to get user config directory")
	}

	fileDir := filepath.Join(configDir, appName)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(fileDir, 0755); err != nil {
		return "", errors.Wrap(err, errors.FSDirCreateFailed, "Failed to create main directory")
	}
	return fileDir, nil
}

func GetUserDir() (string, error) {

	credentials, err := GetUserFromFile()
	if err != nil {
		return "", errors.Wrap(err, errors.FSFileFailed, "Failed to get user from file")
	}

	mainDir, err := getMainDir()
	if err != nil {
		return "", errors.Wrap(err, errors.FSFileFailed, "Failed to get main directory")
	}
	userDir := filepath.Join(mainDir, fmt.Sprintf("user_%d", credentials.ID))

	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", errors.Wrap(err, errors.FSDirCreateFailed, "Failed to create user directory")
	}

	return userDir, nil
}

func getUserDirWithID(userID uint) (string, error) {

	mainDir, err := getMainDir()
	if err != nil {
		return "", errors.Wrap(err, errors.FSFileNotFound, "Failed to get main directory")
	}

	userDir := filepath.Join(mainDir, fmt.Sprintf("user_%d", userID))

	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", errors.Wrap(err, errors.FSDirCreateFailed, "Failed to create user directory")
	}

	return userDir, nil
}

func GetDocumentDir() (string, error) {

	userDir, err := GetUserDir()
	if err != nil {
		return "", err
	}

	documentDir := filepath.Join(userDir, "documents")
	// Create directory if it doesn't exist
	if err := os.MkdirAll(documentDir, 0755); err != nil {
		return "", errors.Wrap(err, errors.FSDirCreateFailed, "Failed to create document directory")
	}

	return documentDir, nil
}

func GetDBPath() (string, error) {

	userDir, err := GetUserDir()
	if err != nil {
		return "", errors.Wrap(err, errors.FSDirFailed, "Failed to get user directory")
	}

	dbPath := filepath.Join(userDir, "data.db")

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return "", errors.NewAppError(errors.FSFileNotFound, "Database file not found", nil)
	}

	return dbPath, nil
}

func GetDBPathWithID(userID uint) (string, error) {

	userDir, err := getUserDirWithID(userID)
	if err != nil {
		return "", errors.Wrap(err, errors.FSDirFailed, "Failed to get user directory")
	}

	dbPath := filepath.Join(userDir, "data.db")

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_, err := os.Create(dbPath)
		if err != nil {
			return "", errors.Wrap(err, errors.FSCreateFailed, "Failed to create database file")
		}
	}

	return dbPath, nil
}

// getCookieFilePath returns the canonical path for the cookie file.
func GetCookieFilePath() (string, error) {

	userDir, err := GetUserDir()
	if err != nil {
		return "", errors.Wrap(err, errors.FSDirFailed, "Failed to get user directory")
	}
	cookiePath := filepath.Join(userDir, "cookies.txt")

	return cookiePath, nil
}

func GetCredentialFile() (string, error) {

	mainDir, err := getMainDir()
	if err != nil {
		return "", errors.Wrap(err, errors.FSDirFailed, "Failed to get main directory")
	}
	credentialsPath := filepath.Join(mainDir, "credentials.json")
	if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
		_, err := os.Create(credentialsPath)
		if err != nil {
			return "", errors.Wrap(err, errors.FSCreateFailed, "Failed to create credential file")
		}
	}
	return credentialsPath, nil
}

func GetProfilePicturePath() (string, error) {

	userDir, err := GetUserDir()
	if err != nil {
		return "", errors.Wrap(err, errors.FSDirFailed, "Failed to get user directory")
	}
	profilePicturePath := filepath.Join(userDir, "profile_picture.png")
	if _, err := os.Stat(profilePicturePath); os.IsNotExist(err) {
		return "", errors.Wrap(err, errors.FSFileNotFound, "Profile picture file not found")
	}
	return profilePicturePath, nil
}
