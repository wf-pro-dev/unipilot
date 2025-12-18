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

func GetAssignments() ([]map[string]string, error) {

	var response struct {
		Message     string              `json:"message"`
		Assignments []map[string]string `json:"assignments"`
		Error       string              `json:"error,omitempty"`
	}

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
			var serverError *errors.ServerError
			if err := json.Unmarshal(body, &serverError); err != nil {
				return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
			}
			return nil, serverError
		}

		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		if response.Error != "" {
			return nil, fmt.Errorf(response.Error)
		}

		if response.Assignments == nil {
			return make([]map[string]string, 0), nil
		}
	}

	return response.Assignments, nil

}

func CreateAssignment(a *assignment.Assignment) (map[string]interface{}, error) {

	assignmentData := a.ToMap()

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Post(fmt.Sprintf("%s/assignments", api_url))
	agent.JSON(assignmentData)

	if err := setAuthHeader(agent); err != nil {
		return nil, err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("server returned status %d: %s", statusCode, string(body))
	}

	var response struct {
		Message    string                 `json:"message"`
		Assignment map[string]interface{} `json:"assignment"`
		Error      string                 `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if statusCode != 200 {
		var serverError *errors.ServerError
		if err := json.Unmarshal(body, &serverError); err != nil {
			return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
		}
		return nil, serverError
	}

	if response.Assignment == nil {
		return nil, fmt.Errorf("no assignment data in response")
	}

	return response.Assignment, nil
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
