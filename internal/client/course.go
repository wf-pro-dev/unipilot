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
	"unipilot/internal/models/course"
	"unipilot/internal/network"
	"unipilot/internal/secrets"
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

		api_url, err := secrets.GetEnvVar("API_URL")
		if err != nil {
			return nil, fmt.Errorf("failed to get api url: %w", err)
		}

		resp, err := client.Get(fmt.Sprintf("%s/course/get", api_url))

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

	api_url, err := secrets.GetEnvVar("API_URL")
	if err != nil {
		return nil, fmt.Errorf("failed to get api url: %w", err)
	}

	courseData := c.ToMap()

	new_client, err := NewClientWithCookies()
	if err != nil {
		return nil, err
	}

	jsonData, _ := json.Marshal(courseData)

	fmt.Println("Creating course:", courseData["code"])

	resp, err := new_client.Post(
		fmt.Sprintf("%s/course", api_url),
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

	api_url, err := secrets.GetEnvVar("API_URL")
	if err != nil {
		return fmt.Errorf("failed to get api url: %w", err)
	}

	resp, err := new_client.Post(
		fmt.Sprintf("%s/course/update", api_url),
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

func RequestLinkCourse(courseCode string, usersID []uint) error {

	new_client, err := NewClientWithCookies()
	if err != nil {
		return err
	}

	linkData := map[string]interface{}{
		"course_code": courseCode,
		"users_id":    usersID,
	}

	jsonData, _ := json.Marshal(linkData)

	api_url, err := secrets.GetEnvVar("API_URL")
	if err != nil {
		return fmt.Errorf("failed to get api url: %w", err)
	}

	resp, err := new_client.Post(
		fmt.Sprintf("%s/course/link/request", api_url),
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

type AcceptLinkCourseResponse struct {
	Error       error                   `json:"error,omitempty"`
	Assignments []assignment.Assignment `json:"assignments"`
}

func AcceptLinkCourse(c *course.Course) (*AcceptLinkCourseResponse, error) {

	new_client, err := NewClientWithCookies()
	if err != nil {
		return nil, err
	}

	jsonData, _ := json.Marshal(c)

	api_url, err := secrets.GetEnvVar("API_URL")
	if err != nil {
		return nil, fmt.Errorf("failed to get api url: %w", err)
	}

	// Get the  course assignments and documents
	resp, err := new_client.Post(
		fmt.Sprintf("%s/course/link/accept", api_url),
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	fmt.Println("Response status code:", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	var response AcceptLinkCourseResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != nil {
		return nil, response.Error
	}

	if response.Assignments == nil {
		return nil, errors.New("no assignments in response")
	}
	log.Println("Assignments:", response.Assignments)

	return &response, nil
}
