package client

import (
	"encoding/json"
	"fmt"
	"log"
	"unipilot/internal/models"
	"unipilot/internal/secrets"

	"unipilot/internal/errors"

	"github.com/gofiber/fiber/v2"
)

func GetCourses() ([]models.Course, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Get(fmt.Sprintf("%s/courses", api_url))

	if err := SetAuthHeader(agent); err != nil {
		return nil, err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	if statusCode != 200 {
		serverError := errors.ParseServerError(body, statusCode)
		return nil, serverError
	}

	var response []models.Course
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
	}

	return response, nil

}

func CreateCourse(c *models.Course) (uint, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Post(fmt.Sprintf("%s/courses", api_url))
	agent.JSON(c)

	if err := SetAuthHeader(agent); err != nil {
		return 0, err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return 0, errs[0]
	}

	if statusCode != 200 {
		var serverError *errors.ServerError
		if err := json.Unmarshal(body, &serverError); err != nil {
			return 0, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
		}
		return 0, serverError
	}

	var response struct {
		RemoteID uint `json:"remote_id"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
	}

	return response.RemoteID, nil
}

func UpdateCourse(id, column, value string) error {

	updateData := map[string]interface{}{
		"value":  value,
		"column": column,
	}

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Put(fmt.Sprintf("%s/courses/%s", api_url, id))
	agent.JSON(updateData)

	if err := SetAuthHeader(agent); err != nil {
		return err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return errs[0]
	}

	if statusCode != 200 {
		serverError := errors.ParseServerError(body, statusCode)
		return serverError
	}

	return nil
}

func DeleteCourse(id string) error {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Delete(fmt.Sprintf("%s/courses/%s", api_url, id))

	if err := SetAuthHeader(agent); err != nil {
		return err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return errs[0]
	}

	if statusCode != 200 {
		serverError := errors.ParseServerError(body, statusCode)
		return serverError
	}

	return nil
}

func RequestLinkCourse(c *models.LocalCourse, usersID []uint) error {

	linkData := map[string]interface{}{
		"users_id": usersID,
	}

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Post(fmt.Sprintf("%s/courses/%d/link-request", api_url, c.RemoteID))
	agent.JSON(linkData)

	if err := SetAuthHeader(agent); err != nil {
		return err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return errs[0]
	}

	if statusCode != 200 {
		serverError := errors.ParseServerError(body, statusCode)
		return serverError
	}

	return nil
}

type AcceptLinkCourseResponse struct {
	Assignments []models.Assignment `json:"assignments"`
}

func AcceptLinkCourse(c *models.Course) ([]models.Assignment, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Post(fmt.Sprintf("%s/courses/link-accept", api_url))
	agent.JSON(c)

	if err := SetAuthHeader(agent); err != nil {
		return nil, err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	fmt.Println("Response status code:", statusCode)

	if statusCode != 200 {
		serverError := errors.ParseServerError(body, statusCode)
		return nil, serverError
	}

	var response []models.Assignment
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
	}

	log.Printf("Assignments: %+v", len(response))

	return response, nil
}

func GetCoursesLinked() (map[string]interface{}, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Get(fmt.Sprintf("%s/courses/linked", api_url))

	if err := SetAuthHeader(agent); err != nil {
		return nil, err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	if statusCode != 200 {
		serverError := errors.ParseServerError(body, statusCode)
		return nil, serverError
	}
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse response")
	}

	return response, nil
}
