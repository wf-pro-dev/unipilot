package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"unipilot/internal/models/assignment"
	"unipilot/internal/models/course"
	"unipilot/internal/secrets"

	"github.com/gofiber/fiber/v2"
)

func GetCourses() ([]map[string]string, error) {

	var response struct {
		Message string              `json:"message"`
		Courses []map[string]string `json:"courses"`
		Error   string              `json:"error,omitempty"`
	}

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Get(fmt.Sprintf("%s/courses", api_url))

	if err := setAuthHeader(agent); err != nil {
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

	if response.Courses == nil {
		return nil, errors.New("no assignment data in response")
	}

	return response.Courses, nil
}

func CreateCourse(c *course.Course) (map[string]interface{}, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	courseData := c.ToMap()

	fmt.Println("Creating course:", courseData["code"])

	agent := fiber.Post(fmt.Sprintf("%s/courses", api_url))
	agent.JSON(courseData)

	if err := setAuthHeader(agent); err != nil {
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
		Message string                 `json:"message"`
		Course  map[string]interface{} `json:"course"`
		Error   string                 `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
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

func UpdateCourse(id, column, value string) error {

	updateData := map[string]interface{}{
		"value":  value,
		"column": column,
	}

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Put(fmt.Sprintf("%s/courses/%s", api_url, id))
	agent.JSON(updateData)

	if err := setAuthHeader(agent); err != nil {
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

func RequestLinkCourse(c *course.LocalCourse, usersID []uint) error {

	linkData := map[string]interface{}{
		"course_code": c.Code,
		"users_id":    usersID,
	}

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Post(fmt.Sprintf("%s/courses/%d/link-request", api_url, c.RemoteID))
	agent.JSON(linkData)

	if err := setAuthHeader(agent); err != nil {
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

type AcceptLinkCourseResponse struct {
	Error       error                   `json:"error,omitempty"`
	Assignments []assignment.Assignment `json:"assignments"`
}

func AcceptLinkCourse(c *course.Course) (*AcceptLinkCourseResponse, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Post(fmt.Sprintf("%s/courses/%d/link-accept", api_url, c.ID))
	agent.JSON(c)

	if err := setAuthHeader(agent); err != nil {
		return nil, err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	fmt.Println("Response status code:", statusCode)

	if statusCode != 200 {
		return nil, fmt.Errorf("server returned status %d: %s", statusCode, string(body))
	}

	var response AcceptLinkCourseResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != nil {
		return nil, response.Error
	}

	if response.Assignments == nil {
		return nil, errors.New("no assignments in response")
	}
	return &response, nil
}

func DeleteCourse(id string) error {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Delete(fmt.Sprintf("%s/courses/%s", api_url, id))

	if err := setAuthHeader(agent); err != nil {
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
