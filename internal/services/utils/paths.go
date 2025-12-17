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
		return "", errors.Wrap(err, errors.FSPathNotFound, "Failed to get user config directory")
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
		return "", errors.Wrap(err, errors.FSFileNotFound, "Failed to get user from file")
	}

	fileDir, err := getMainDir()
	if err != nil {
		return "", errors.Wrap(err, errors.FSPathNotFound, "Failed to get main directory")
	}
	userDir := filepath.Join(fileDir, fmt.Sprintf("user_%d", credentials.ID))

	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", errors.Wrap(err, errors.FSDirCreateFailed, "Failed to create user directory")
	}

	return userDir, nil
}

func getUserDirWithID(userID uint) (string, error) {

	fileDir, err := getMainDir()
	if err != nil {
		return "", errors.Wrap(err, errors.FSPathNotFound, "Failed to get main directory")
	}

	userDir := filepath.Join(fileDir, fmt.Sprintf("user_%d", userID))

	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", errors.Wrap(err, errors.FSDirCreateFailed, "Failed to create user directory")
	}

	return userDir, nil
}

func GetDocumentDir() (string, error) {

	fileDir, err := GetUserDir()
	if err != nil {
		return "", err
	}

	documentDir := filepath.Join(fileDir, "documents")
	// Create directory if it doesn't exist
	if err := os.MkdirAll(documentDir, 0755); err != nil {
		return "", errors.Wrap(err, errors.FSDirCreateFailed, "Failed to create document directory")
	}

	return documentDir, nil
}

func GetDBPath() (string, error) {

	fileDir, err := GetUserDir()
	if err != nil {
		return "", errors.Wrap(err, errors.FSPathNotFound, "Failed to get user directory")
	}

	dbPath := filepath.Join(fileDir, "data.db")

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		_, err := os.Create(dbPath)
		if err != nil {
			return "", errors.Wrap(err, errors.FSCreateFailed, "Failed to create database file")
		}
	}

	return dbPath, nil
}

func GetDBPathWithID(userID uint) (string, error) {

	fileDir, err := getUserDirWithID(userID)
	if err != nil {
		return "", errors.Wrap(err, errors.FSPathNotFound, "Failed to get user directory")
	}

	dbPath := filepath.Join(fileDir, "data.db")

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

	fileDir, err := GetUserDir()
	if err != nil {
		return "", errors.Wrap(err, errors.FSPathNotFound, "Failed to get user directory")
	}
	cookiePath := filepath.Join(fileDir, "cookies.txt")

	return cookiePath, nil
}

func GetCredentialFile() (string, error) {

	fileDir, err := getMainDir()
	if err != nil {
		return "", errors.Wrap(err, errors.FSPathNotFound, "Failed to get main directory")
	}
	credentialsPath := filepath.Join(fileDir, "credentials.json")
	if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
		_, err := os.Create(credentialsPath)
		if err != nil {
			return "", errors.Wrap(err, errors.FSCreateFailed, "Failed to create credential file")
		}
	}
	return credentialsPath, nil
}

func GetProfilePicturePath() (string, error) {

	fileDir, err := GetUserDir()
	if err != nil {
		return "", errors.Wrap(err, errors.FSPathNotFound, "Failed to get user directory")
	}
	profilePicturePath := filepath.Join(fileDir, "profile_picture.png")
	if _, err := os.Stat(profilePicturePath); os.IsNotExist(err) {
		return "", errors.Wrap(err, errors.FSFileNotFound, "Profile picture not found")
	}
	return profilePicturePath, nil
}
