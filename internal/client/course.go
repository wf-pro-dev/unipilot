package client

import (
	"encoding/json"
	"fmt"
	"unipilot/internal/models"
	"unipilot/internal/secrets"

	"unipilot/internal/errors"

	"github.com/gofiber/fiber/v2"
)

// FriendStatusResponse represents the friendship status between two users
type CourseStatusResponse struct {
	ID              string                   `json:"id"`                 // Course ID
	Status          *models.InvitationStatus `json:"status"`             // Current friendship status (null if no relationship)
	IsPendingForYou bool                     `json:"is_pending_for_you"` // True if you need to respond to their reque	st	MutualFriendsCount  int               `json:"mutual_friends_count"`  // Number of mutual friends
}

func GetCourses(userID string) ([]models.Course, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Get(fmt.Sprintf("%s/courses/%s", api_url, userID))

	if err := SetAuthHeader(agent); err != nil {
		return nil, err
	}

	statusCode, body, _ := agent.Bytes()
	if statusCode != fiber.StatusOK {
		serverError := errors.ParseServerError(body, statusCode)
		return nil, serverError
	}

	var courses []models.Course
	if err := json.Unmarshal(body, &courses); err != nil {
		return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
	}

	return courses, nil

}

func CreateCourse(c *models.Course) error {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Post(fmt.Sprintf("%s/courses", api_url))
	agent.JSON(c)

	if err := SetAuthHeader(agent); err != nil {
		return err
	}

	statusCode, body, _ := agent.Bytes()
	if statusCode != fiber.StatusCreated {
		serverError := errors.ParseServerError(body, statusCode)
		return serverError
	}

	return nil
}

func UpdateCourse(id string, column, value string) error {

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

	statusCode, body, _ := agent.Bytes()
	if statusCode != fiber.StatusNoContent {
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

	statusCode, body, _ := agent.Bytes()
	if statusCode != fiber.StatusNoContent {
		serverError := errors.ParseServerError(body, statusCode)
		return serverError
	}

	return nil
}

func CourseShare(c *models.LocalCourse, usersID []string) error {

	linkData := map[string]interface{}{
		"users_id": usersID,
	}

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Post(fmt.Sprintf("%s/courses/%s/share", api_url, c.ID))
	agent.JSON(linkData)

	if err := SetAuthHeader(agent); err != nil {
		return err
	}

	statusCode, body, _ := agent.Bytes()
	if statusCode != fiber.StatusNoContent {
		serverError := errors.ParseServerError(body, statusCode)
		return serverError
	}

	return nil
}

func GetClusterStatus(courseID string, userID string) (*CourseStatusResponse, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Get(fmt.Sprintf("%s/courses/%s/cluster-status/%s", api_url, courseID, userID))

	if err := SetAuthHeader(agent); err != nil {
		return nil, err
	}

	statusCode, body, _ := agent.Bytes()
	if statusCode != fiber.StatusOK {
		serverError := errors.ParseServerError(body, statusCode)
		return nil, serverError
	}

	var courseStatusResponse *CourseStatusResponse
	if err := json.Unmarshal(body, &courseStatusResponse); err != nil {
		return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse cluster status response")
	}

	return courseStatusResponse, nil
}

func SendClusterRequest(courseID string) error {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Post(fmt.Sprintf("%s/courses/%s/request", api_url, courseID))

	if err := SetAuthHeader(agent); err != nil {
		return err
	}

	statusCode, body, _ := agent.Bytes()
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
	agent := fiber.Post(fmt.Sprintf("%s/courses/%s/accept", api_url, invitation.ID))

	if err := SetAuthHeader(agent); err != nil {
		return err
	}

	statusCode, body, _ := agent.Bytes()
	if statusCode != fiber.StatusNoContent {
		serverError := errors.ParseServerError(body, statusCode)
		return serverError
	}

	return nil
}

func DeclineCourseInvitation(invitation *models.CourseInvitation) error {
	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Post(fmt.Sprintf("%s/courses/%s/decline", api_url, invitation.ID))

	if err := SetAuthHeader(agent); err != nil {
		return err
	}

	statusCode, body, _ := agent.Bytes()
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

func GetClusterUsers(courseID string, cursor *models.Cursor, limit int, search string) (*models.PageResponse[models.User], error) {
	api_url := secrets.CONSTANTS["API_URL"]
	query := fmt.Sprintf("%s/courses/%s/cluster-users?limit=%d", api_url, courseID, limit)
	if search != "" {
		query += fmt.Sprintf("&search=%s", search)
	}

	agent := fiber.Get(query).JSON(cursor)

	if err := SetAuthHeader(agent); err != nil {
		return nil, err
	}

	statusCode, body, _ := agent.Bytes()
	if statusCode != 200 {
		err := errors.ParseServerError(body, statusCode)
		return nil, err
	}

	var response *models.PageResponse[models.User]
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse response")
	}

	return response, nil
}
