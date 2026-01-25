package client

import (
	"net/http"
	"sync"

	"unipilot/internal/errors"

	"github.com/gofiber/fiber/v2"
)

var (
	mu           sync.Mutex
	cond         *sync.Cond
	isRefreshing bool
)

func init() {
	cond = sync.NewCond(&mu)
}

func GetAuthAgent(agent *fiber.Agent) (*fiber.Agent, error) {

	if err := SetAuthHeader(agent); err != nil {
		return nil, err
	}

	return agent, nil
}

// SetAuthHeader sets the Authorization header on a fiber agent with automatic token refresh
func SetAuthHeader(agent *fiber.Agent) error {
	token, err := GetAuthToken()
	if err != nil {
		return err
	}

	// Add Authorization header if token exists
	if token != "" {
		agent.Set("Authorization", "Bearer "+token)
	}

	return nil
}

func GetAuthToken() (string, error) {
	token, err := LoadToken()
	if err != nil {
		return "", errors.Wrap(err, errors.FSFileNotFound, "Failed to load token")
	}

	if IsTokenValid() {
		return token, nil
	}

	mu.Lock()
	defer mu.Unlock()

	// Recheck after acquiring lock
	if IsTokenValid() {
		return LoadToken()
	}

	// Wait if refresh in progress
	for isRefreshing {
		cond.Wait()
	}

	// Recheck after waiting (another goroutine might have refreshed)
	if IsTokenValid() {
		return LoadToken()
	}

	// Perform refresh with cleanup guaranteed
	isRefreshing = true
	defer func() {
		isRefreshing = false
		cond.Broadcast()
	}()

	return doRefresh()
}

func doRefresh() (string, error) {
	refreshToken, err := LoadRefreshToken()
	if err != nil {
		return "", errors.Wrap(err, errors.FSFileNotFound, "Failed to load refresh token")
	}

	newToken, newRefreshToken, err := RefreshToken(refreshToken)
	if err != nil {
		return "", errors.Wrap(err, errors.ClientRequestFailed, "Failed to refresh token")
	}

	if err := SaveToken(newToken); err != nil {
		return "", errors.Wrap(err, errors.FSWriteFailed, "Failed to save token")
	}

	if err := SaveRefreshToken(newRefreshToken); err != nil {
		return "", errors.Wrap(err, errors.FSWriteFailed, "Failed to save refresh token")
	}

	return newToken, nil
}

// SetAuthHeaderRequest sets the Authorization header on an http.Request with automatic token refresh
func SetAuthHeaderRequest(req *http.Request) error {
	token, err := GetAuthToken()
	if err != nil {
		return err
	}

	// Add Authorization header if token exists
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return nil
}
