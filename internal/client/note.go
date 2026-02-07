package client

import (
	"encoding/json"
	"fmt"
	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/secrets"

	"github.com/gofiber/fiber/v2"
)

func GetNotes() ([]models.Note, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent, err := GetAuthAgent(fiber.Get(fmt.Sprintf("%s/notes", api_url)))
	if err != nil {
		return nil, err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("server returned status %d: %s", statusCode, string(body))
	}

	var notes []models.Note
	if err := json.Unmarshal(body, &notes); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return notes, nil
}

func CreateNote(n *models.Note) error {

	api_url := secrets.CONSTANTS["API_URL"]
	agent, err := GetAuthAgent(fiber.Post(fmt.Sprintf("%s/notes", api_url)).JSON(n))
	if err != nil {
		return err
	}

	statusCode, body, _ := agent.Bytes()
	if statusCode != fiber.StatusCreated {
		serverError := errors.ParseServerError(body, statusCode)
		return serverError
	}

	return nil
}

func UpdateNote(id string, column, value string) error {

	updateData := map[string]interface{}{
		"value":  value,
		"column": column,
	}

	api_url := secrets.CONSTANTS["API_URL"]
	agent, err := GetAuthAgent(fiber.Put(fmt.Sprintf("%s/notes/%s", api_url, id)).JSON(updateData))
	if err != nil {
		return err
	}

	statusCode, body, _ := agent.Bytes()
	if statusCode != fiber.StatusNoContent {
		serverError := errors.ParseServerError(body, statusCode)
		return serverError
	}

	return nil
}

func DeleteNote(id string) error {

	api_url := secrets.CONSTANTS["API_URL"]
	agent, err := GetAuthAgent(fiber.Delete(fmt.Sprintf("%s/notes/%s", api_url, id)))
	if err != nil {
		return err
	}

	statusCode, body, _ := agent.Bytes()
	if statusCode != fiber.StatusNoContent {
		serverError := errors.ParseServerError(body, statusCode)
		return serverError
	}

	return nil
}
