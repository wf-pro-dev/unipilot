package client

import (
	"encoding/json"
	"fmt"
	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/secrets"

	"github.com/gofiber/fiber/v2"
)

func GetRemoteUsers() ([]models.User, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Get(fmt.Sprintf("%s/users", api_url))

	if err := SetAuthHeader(agent); err != nil {
		return nil, err
	}

	statusCode, body, _ := agent.Bytes()
	if statusCode != 200 {
		var serverError *errors.AppError
		if err := json.Unmarshal(body, &serverError); err != nil {
			return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
		}
		return nil, serverError.ToServerError(statusCode)
	}

	var response []models.User
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse response")
	}

	return response, nil

}
