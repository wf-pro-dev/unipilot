package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"unipilot/internal/models/note"
	"unipilot/internal/secrets"
)

func CreateNote(n *note.Note) (map[string]string, error) {

	noteData := n.ToMap()

	new_client, err := NewClientWithCookies()
	if err != nil {
		return nil, err
	}

	jsonData, _ := json.Marshal(noteData)

	api_url, err := secrets.GetEnvVar("API_URL")
	if err != nil {
		return nil, fmt.Errorf("failed to get api url: %w", err)
	}

	resp, err := new_client.Post(
		fmt.Sprintf("%s/note", api_url),
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	log.Printf("Response status code: %d\n", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Message string            `json:"message"`
		Note    map[string]string `json:"note"`
		Error   string            `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != "" {
		return nil, errors.New(response.Error)
	}

	if response.Note == nil {
		return nil, fmt.Errorf("no note data in response")
	}

	return response.Note, nil
}

func SendNoteUpdate(id, column, value string) error {

	new_client, err := NewClientWithCookies()
	if err != nil {

		return err
	}

	updateData := map[string]interface{}{
		"id":     id,
		"value":  value,
		"column": column,
	}

	jsonData, _ := json.Marshal(updateData)

	api_url, err := secrets.GetEnvVar("API_URL")
	if err != nil {
		return fmt.Errorf("failed to get api url: %w", err)
	}

	resp, err := new_client.Post(
		fmt.Sprintf("%s/note/update", api_url),
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
