package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"unipilot/internal/models/course"
	"unipilot/internal/network"
	"unipilot/internal/storage"
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

func CreateCourse(courseData map[string]string) (map[string]string, error) {

	db, _, err := storage.GetLocalDB()
	if err != nil {
		return nil, err
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create local assignment
	c := course.LocalCourse{
		Code:       courseData["code"],
		Name:       courseData["name"],
		Duration:   courseData["duration"],
		RoomNumber: courseData["room_number"],
	}

	isOnline := network.IsOnline()

	if isOnline {

		new_client, err := NewClientWithCookies()
		if err != nil {
			return nil, err
		}

		jsonData, _ := json.Marshal(courseData)

		resp, err := new_client.Post(
			"https://newsroom.dedyn.io/acc-homework/course",
			"application/json",
			bytes.NewBuffer(jsonData),
		)

		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
		}

		var response struct {
			Message string            `json:"message"`
			Course  map[string]string `json:"course"`
			Error   string            `json:"error,omitempty"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		if response.Error != "" {
			return nil, errors.New(response.Error)
		}

		if response.Course == nil {
			return nil, errors.New("no course data in response")
		}

		c.NotionID = response.Course["notion_id"]
		c.SyncStatus = course.SyncStatusSynced

	} else {
		c.SyncStatus = course.SyncStatusPending
	}

	if err := tx.Create(&c).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("local create failed: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("commit failed: %w", err)
	}

	return c.ToMap(), nil
}
