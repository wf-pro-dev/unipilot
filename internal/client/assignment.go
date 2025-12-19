package client

import (
	"encoding/json"
	"fmt"
	"unipilot/internal/models/assignment"
	"unipilot/internal/network"
	"unipilot/internal/secrets"

	"unipilot/internal/errors"

	"github.com/gofiber/fiber/v2"
)

func GetAssignments() ([]assignment.Assignment, error) {

	isOnline := network.IsOnline()

	if isOnline {
		api_url := secrets.CONSTANTS["API_URL"]
		agent := fiber.Get(fmt.Sprintf("%s/assignments", api_url))

		if err := setAuthHeader(agent); err != nil {
			return nil, err
		}

		statusCode, body, errs := agent.Bytes()
		if len(errs) > 0 {
			return nil, errs[0]
		}

		if statusCode != 200 {
			serverError := errors.ParseServerError(body, statusCode)
			return nil, serverError
		}

		var response []assignment.Assignment
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
		}

		return response, nil
	}

	return nil, nil

}

func CreateAssignment(a *assignment.Assignment) (*assignment.Assignment, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Post(fmt.Sprintf("%s/assignments", api_url))
	agent.JSON(a)

	if err := setAuthHeader(agent); err != nil {
		return nil, err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	if statusCode != 200 {
		var serverError *errors.ServerError
		if err := json.Unmarshal(body, &serverError); err != nil {
			return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
		}
		return nil, serverError
	}

	var response *assignment.Assignment
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response, nil
}

func UpdateAssignment(id, column, value string) error {

	updateData := map[string]string{
		"value":  value,
		"column": column,
	}

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Put(fmt.Sprintf("%s/assignments/%s", api_url, id))
	agent.JSON(updateData)

	if err := setAuthHeader(agent); err != nil {
		return err
	}

	statusCode, _, errs := agent.Bytes()
	if len(errs) > 0 {
		return errs[0]
	}
	if statusCode != 200 {
		return fmt.Errorf("server returned status %d", statusCode)
	}

	return nil

}

func DeleteAssignment(id string) error {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Delete(fmt.Sprintf("%s/assignments/%s", api_url, id))

	if err := setAuthHeader(agent); err != nil {
		return err
	}

	statusCode, _, errs := agent.Bytes()
	if len(errs) > 0 {
		return errs[0]
	}
	if statusCode != 200 {
		return fmt.Errorf("server returned status %d", statusCode)
	}

	return nil
}
