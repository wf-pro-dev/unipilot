package client

import (
	"encoding/json"
	"fmt"
	"unipilot/internal/errors"
	"unipilot/internal/models/user"
	"unipilot/internal/secrets"

	"github.com/gofiber/fiber/v2"
)

func GetRemoteUsers() ([]user.User, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Get(fmt.Sprintf("%s/users", api_url))

	if err := SetAuthHeader(agent); err != nil {
		return nil, err
	}

	statusCode, body, errs := agent.Bytes()
	if errs != nil && len(errs) > 0 {
		return nil, errs[0]
	}

	if statusCode != 200 {
		var serverError *errors.AppError
		if err := json.Unmarshal(body, &serverError); err != nil {
			return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
		}
		return nil, serverError.ToServerError(statusCode)
	}

	var response []user.User
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse response")
	}

	return response, nil

}
