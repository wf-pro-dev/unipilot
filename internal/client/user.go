package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	resp, err := new_client.Post(
		"https://newsroom.dedyn.io/acc-homework/user/update",
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
