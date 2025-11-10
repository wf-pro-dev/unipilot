package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"unipilot/internal/models/assignment"
	"unipilot/internal/network"
	"unipilot/internal/secrets"
)

func GetAssignments() ([]map[string]string, error) {

	var response struct {
		Message     string              `json:"message"`
		Assignments []map[string]string `json:"assignments"`
		Error       string              `json:"error,omitempty"`
	}

	isOnline := network.IsOnline()

	if isOnline {

		client, err := NewAuthClient()
		if err != nil {
			return nil, err
		}

		api_url := secrets.CONSTANTS["API_URL"]
		if err != nil {
			return nil, fmt.Errorf("failed to get api url: %w", err)
		}

		resp, err := client.Get(fmt.Sprintf("%s/assignments", api_url))

		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
		}

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
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

	new_client, err := NewAuthClient()
	if err != nil {
		return nil, err
	}

	jsonData, _ := json.Marshal(assignmentData)

	api_url := secrets.CONSTANTS["API_URL"]

	resp, err := new_client.Post(
		fmt.Sprintf("%s/assignment", api_url),
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
		Message    string                 `json:"message"`
		Assignment map[string]interface{} `json:"assignment"`
		Error      string                 `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != "" {
		return nil, errors.New(response.Error)
	}

	if response.Assignment == nil {
		return nil, fmt.Errorf("no assignment data in response")
	}

	return response.Assignment, nil
}

func SendAssignmentUpdate(id, column, value string) error {

	new_client, err := NewAuthClient()
	if err != nil {

		return err
	}

	updateData := map[string]interface{}{
		"id":     id,
		"value":  value,
		"column": column,
	}

	jsonData, _ := json.Marshal(updateData)

	api_url := secrets.CONSTANTS["API_URL"]

	resp, err := new_client.Post(
		fmt.Sprintf("%s/assignment/update", api_url),
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
