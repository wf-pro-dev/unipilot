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
)

func CreateNote(n *note.Note) (map[string]string, error) {

	noteData := n.ToMap()

	new_client, err := NewClientWithCookies()
	if err != nil {
		return nil, err
	}

	jsonData, _ := json.Marshal(noteData)

	resp, err := new_client.Post(
		"https://newsroom.dedyn.io/acc-homework/note",
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

	resp, err := new_client.Post(
		"https://newsroom.dedyn.io/acc-homework/note/update",
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
