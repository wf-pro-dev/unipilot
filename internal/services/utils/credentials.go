package utils

import (
	"encoding/json"
	"log"
	"os"
	"unipilot/internal/models/user"
)

func GetUserFromFile() (*user.User, error) {

	credentialsFile, err := GetCredentialFile()
	if err != nil {
		return nil, err
	}

	credentials, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, err
	}

	var user user.User
	err = json.Unmarshal(credentials, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func SetCredentials(user *user.User) error {

	credentialsFile, err := GetCredentialFile()
	if err != nil {
		return err
	}

	credentials := user.ToMap()

	data, err := json.Marshal(credentials)
	if err != nil {
		return err
	}

	return os.WriteFile(credentialsFile, data, 0600)
}

func ClearCredentials() error {
	log.Println("Clearing credentials")
	credentialsFile, err := GetCredentialFile()
	if err != nil {
		return err
	}
	return os.Remove(credentialsFile)
}
