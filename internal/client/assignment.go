package client

import (
	"encoding/json"
	"fmt"
	"unipilot/internal/models"
	"unipilot/internal/secrets"

	"unipilot/internal/errors"

	"github.com/gofiber/fiber/v2"
)

func GetAssignments() ([]models.Assignment, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent, err := GetAuthAgent(fiber.Get(fmt.Sprintf("%s/assignments", api_url)))
	if err != nil {
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

	var response []models.Assignment
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
	}

	return response, nil

}

func CreateAssignment(a *models.Assignment) (uint, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent, err := GetAuthAgent(fiber.Post(fmt.Sprintf("%s/assignments", api_url)).JSON(a))
	if err != nil {
		return 0, err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return 0, errs[0]
	}

	if statusCode != 200 {
		var serverError *errors.ServerError
		if err := json.Unmarshal(body, &serverError); err != nil {
			return 0, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
		}
		return 0, serverError
	}

	var response struct {
		RemoteID uint `json:"remote_id"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
	}

	return response.RemoteID, nil
}

func UpdateAssignment(id, column, value string) error {

	updateData := map[string]string{
		"value":  value,
		"column": column,
	}

	api_url := secrets.CONSTANTS["API_URL"]
	agent, err := GetAuthAgent(fiber.Put(fmt.Sprintf("%s/assignments/%s", api_url, id)).JSON(updateData))
	if err != nil {
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
	agent, err := GetAuthAgent(fiber.Delete(fmt.Sprintf("%s/assignments/%s", api_url, id)))
	if err != nil {
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
