package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/secrets"

	"github.com/gofiber/fiber/v2"
)

func GetRemoteUsers(cursor *models.Cursor, limit int, search string, filters models.Filter) (*models.PageResponse[models.User], error) {

	api_url := secrets.CONSTANTS["API_URL"]
	query := fmt.Sprintf("%s/users?limit=%d", api_url, limit)
	if search != "" {
		query += fmt.Sprintf("&search=%s", search)
	}
	if len(filters) > 0 {
		for key, value := range filters {
			value = url.QueryEscape(value)
			query += fmt.Sprintf("&%s=%s", key, value)
		}
	}
	agent := fiber.Get(query).JSON(cursor)

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
