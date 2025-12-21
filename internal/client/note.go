package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"unipilot/internal/models/note"
	"unipilot/internal/secrets"

	"github.com/gofiber/fiber/v2"
)

func GetNotes() ([]map[string]string, error) {
	var response struct {
		Message string              `json:"message"`
		Notes   []map[string]string `json:"notes"`
		Error   string              `json:"error,omitempty"`
	}

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Get(fmt.Sprintf("%s/notes", api_url))

	if err := SetAuthHeader(agent); err != nil {
		return nil, err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("server returned status %d: %s", statusCode, string(body))
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != "" {
		return nil, errors.New(response.Error)
	}

	if response.Notes == nil {
		return make([]map[string]string, 0), nil
	}

	return response.Notes, nil
}

func CreateNote(n *note.LocalNote) (map[string]string, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Post(fmt.Sprintf("%s/notes", api_url))
	agent.JSON(n)

	if err := SetAuthHeader(agent); err != nil {
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
		Message string            `json:"message"`
		Note    map[string]string `json:"note"`
		Error   string            `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != "" {
		return nil, errors.New(response.Error)
	}

	if response.Note == nil || response.Note["content"] == "" {
		return nil, fmt.Errorf("Invalid note data in response")
	}

	return response.Note, nil
}

func UpdateNote(id, column, value string) error {

	updateData := map[string]interface{}{
		"value":  value,
		"column": column,
	}

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Put(fmt.Sprintf("%s/notes/%s", api_url, id))
	agent.JSON(updateData)

	if err := SetAuthHeader(agent); err != nil {
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

func DeleteNote(id string) error {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Delete(fmt.Sprintf("%s/notes/%s", api_url, id))

	if err := SetAuthHeader(agent); err != nil {
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
