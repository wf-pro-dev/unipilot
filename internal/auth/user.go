package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"unipilot/internal/secrets"
)

func (a *Auth) GetUser() (map[string]interface{}, error) {

	api_url := secrets.CONSTANTS["API_URL"]

	resp, err := a.Client.Get(fmt.Sprintf("%s/user", api_url))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Message string                 `json:"message"`
		User    map[string]interface{} `json:"user"`
		Error   string                 `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != "" {
		return nil, errors.New(response.Error)
	}

	if response.User == nil {
		return nil, fmt.Errorf("no user data in response")
	}

	return response.User, nil
}
