package client

import (
	"encoding/json"
	"fmt"
	"unipilot/internal/models"
	"unipilot/internal/secrets"

	"unipilot/internal/errors"

	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
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

func CreateCourse(c *models.Course) error {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Post(fmt.Sprintf("%s/courses", api_url))
	agent.JSON(c)

	if err := SetAuthHeader(agent); err != nil {
		return err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return errs[0]
	}

	if statusCode != 200 {
		var serverError *errors.ServerError
		if err := json.Unmarshal(body, &serverError); err != nil {
			return errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
		}
		return serverError
	}

	var response struct {
		RemoteID uint `json:"remote_id"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
	}

	return nil
}

func UpdateCourse(id datatypes.UUID, column, value string) error {

	updateData := map[string]interface{}{
		"value":  value,
		"column": column,
	}

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Put(fmt.Sprintf("%s/courses/%s", api_url, id.String()))
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

func DeleteCourse(id datatypes.UUID) error {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Delete(fmt.Sprintf("%s/courses/%s", api_url, id.String()))

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

func CourseShare(c *models.LocalCourse, usersID []datatypes.UUID) error {

	linkData := map[string]interface{}{
		"users_id": usersID,
	}

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Post(fmt.Sprintf("%s/courses/%s/share", api_url, c.ID.String()))
	agent.JSON(linkData)

	if err := SetAuthHeader(agent); err != nil {
		return err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return errs[0]
	}

	if statusCode != fiber.StatusNoContent {
		serverError := errors.ParseServerError(body, statusCode)
		return serverError
	}

	return nil
}

type AcceptLinkCourseResponse struct {
	Assignments []models.Assignment `json:"assignments"`
}

func AcceptCourseInvitation(invitation *models.CourseInvitation) error {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Post(fmt.Sprintf("%s/courses/%d/accept", api_url, invitation.ID))

	if err := SetAuthHeader(agent); err != nil {
		return err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return errs[0]
	}

	if statusCode != fiber.StatusNoContent {
		serverError := errors.ParseServerError(body, statusCode)
		return serverError
	}

	return nil
}

func DeclineCourseInvitation(invitation *models.CourseInvitation) error {
	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Post(fmt.Sprintf("%s/courses/%d/decline", api_url, invitation.ID))

	if err := SetAuthHeader(agent); err != nil {
		return err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return errs[0]
	}
	if statusCode != fiber.StatusNoContent {
		serverError := errors.ParseServerError(body, statusCode)
		return serverError
	}

	return nil
}

func GetCoursesLinked() ([]models.Course, error) {

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
	var courses []models.Course
	if err := json.Unmarshal(body, &courses); err != nil {
		return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse response")
	}

	return courses, nil
}
