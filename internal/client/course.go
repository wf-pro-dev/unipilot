package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"unipilot/internal/models/course"
	"unipilot/internal/network"
)

func GetCourses() ([]map[string]string, error) {

	var response struct {
		Message string              `json:"message"`
		Courses []map[string]string `json:"courses"`
		Error   string              `json:"error,omitempty"`
	}

	isOnline := network.IsOnline()

	if isOnline {

		client, err := NewClientWithCookies()
		if err != nil {
			return nil, err
		}

		resp, err := client.Get("https://newsroom.dedyn.io/acc-homework/course/get")

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

			return nil, errors.New(response.Error)

		}

		if response.Courses == nil {
			return nil, errors.New("no assignment data in response")
		}

	}

	return response.Courses, nil
}

func CreateCourse(c *course.Course) (map[string]interface{}, error) {

	courseData := c.ToMap()

	new_client, err := NewClientWithCookies()
	if err != nil {
		return nil, err
	}

	jsonData, _ := json.Marshal(courseData)

	fmt.Println("Creating course:", courseData["code"])

	resp, err := new_client.Post(
		"https://newsroom.dedyn.io/acc-homework/course",
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
		Message string                 `json:"message"`
		Course  map[string]interface{} `json:"course"`
		Error   string                 `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != "" {
		return nil, errors.New(response.Error)
	}

	if response.Course == nil {
		return nil, fmt.Errorf("no course data in response")
	}

	return response.Course, nil
}

func SendCourseUpdate(id, column, value string) error {

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
		"https://newsroom.dedyn.io/acc-homework/course/update",
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
