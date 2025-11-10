package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"unipilot/internal/secrets"
)

func SendUserUpdate(column, value string) error {

	new_client, err := NewAuthClient()
	if err != nil {

		return err
	}

	updateData := map[string]interface{}{
		"value":  value,
		"column": column,
	}

	jsonData, _ := json.Marshal(updateData)

	api_url := secrets.CONSTANTS["API_URL"]
	resp, err := new_client.Post(
		fmt.Sprintf("%s/user/update", api_url),
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
