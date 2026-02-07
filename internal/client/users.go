package client

import (
	"encoding/json"
	"fmt"
	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/secrets"

	"github.com/gofiber/fiber/v2"
)

func GetRemoteUsers(cursor *models.Cursor, limit int) (*models.PageResponse[models.User], error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Get(fmt.Sprintf("%s/users?limit=%d", api_url, limit)).JSON(cursor)

	if err := SetAuthHeader(agent); err != nil {
		return nil, err
	}

	statusCode, body, _ := agent.Bytes()
	if statusCode != 200 {
		err := errors.ParseServerError(body, statusCode)
		return nil, err
	}

	var response *models.PageResponse[models.User]
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse response")
	}

	return response, nil

}
