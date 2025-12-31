package utils

import (
	"encoding/json"
	"log"
	"os"
	"unipilot/internal/errors"
	"unipilot/internal/models"
)

func GetUserFromFile() (*models.User, error) {

	credentialsFile, err := GetCredentialFile()
	if err != nil {
		log.Println("Failed to get credential file: ", err)
		return nil, errors.Wrap(err, errors.FSFileNotFound, "Failed to get credential file")
	}

	credentials, err := os.ReadFile(credentialsFile)
	if err != nil {
		log.Println("Failed to read credential file: ", err)
		return nil, errors.Wrap(err, errors.FSOpenFailed, "Failed to read credential file")
	}

	var user models.User
	err = json.Unmarshal(credentials, &user)
	if err != nil {
		return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to unmarshal credential file")
	}

	return &user, nil
}

func SetCredentials(user *models.User) error {

	credentialsFile, err := GetCredentialFile()
	if err != nil {
		return errors.Wrap(err, errors.FSFileNotFound, "Failed to found credential file")
	}

	credentials := user.ToMap()

	data, err := json.Marshal(credentials)
	if err != nil {
		return errors.Wrap(err, errors.ProcJSONMarshalFailed, "Failed to marshal credential file")
	}

	err = os.WriteFile(credentialsFile, data, 0600)
	if err != nil {
		return errors.Wrap(err, errors.FSWriteFailed, "Failed to write credential file")
	}
	return nil
}

func ClearCredentials() error {
	credentialsFile, err := GetCredentialFile()
	if err != nil {
		return errors.Wrap(err, errors.FSFileNotFound, "Failed to found credential file")
	}
	err = os.Remove(credentialsFile)
	if err != nil {
		return errors.Wrap(err, errors.FSDeleteFailed, "Failed to delete credential file")
	}
	return nil
}
