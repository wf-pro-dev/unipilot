package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"unipilot/internal/secrets"
	"unipilot/internal/services/utils"

	"github.com/golang-jwt/jwt/v5"
)

type TokenData struct {
	Token     string    `json:"token"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// SaveToken saves the JWT token to a secure file
func SaveToken(token string) error {
	// Parse token to get expiration
	claims := &jwt.RegisteredClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(token, claims)

	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	tokenData := TokenData{
		Token:     token,
		IssuedAt:  claims.IssuedAt.Time,
		ExpiresAt: claims.ExpiresAt.Time,
	}

	tokenFile, err := getTokenFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(tokenData, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal token: %w", err)
	}

	return os.WriteFile(tokenFile, data, 0600) // Secure file permissions
}

// SaveRefreshToken saves the refresh token to a secure file
func SaveRefreshToken(token string) error {
	// Parse token to get expiration
	claims := &jwt.RegisteredClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(token, claims)

	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	refreshTokenData := TokenData{
		Token:     token,
		IssuedAt:  claims.IssuedAt.Time,
		ExpiresAt: claims.ExpiresAt.Time,
	}

	refreshTokenFile, err := getRefreshTokenFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(refreshTokenData, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal refresh token: %w", err)
	}

	return os.WriteFile(refreshTokenFile, data, 0600) // Secure file permissions
}

// LoadToken loads the JWT token from file
func LoadToken() (string, error) {
	tokenFile, err := getTokenFilePath()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		return "", nil // No token file exists
	}

	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return "", fmt.Errorf("could not read token file: %w", err)
	}

	var tokenData TokenData
	if err := json.Unmarshal(data, &tokenData); err != nil {
		return "", fmt.Errorf("could not unmarshal token: %w", err)
	}

	// Check if token is expired
	if time.Now().After(tokenData.ExpiresAt) {
		// Token expired, remove file
		os.Remove(tokenFile)
		return "", nil
	}

	return tokenData.Token, nil
}

func LoadRefreshToken() (string, error) {
	refreshTokenFile, err := getRefreshTokenFilePath()

	if err != nil {
		return "", err
	}
	if _, err := os.Stat(refreshTokenFile); os.IsNotExist(err) {
		return "", nil // No refresh token file exists
	}
	data, err := os.ReadFile(refreshTokenFile)
	if err != nil {
		return "", fmt.Errorf("could not read refresh token file: %w", err)
	}
	var refreshTokenData TokenData
	if err := json.Unmarshal(data, &refreshTokenData); err != nil {
		return "", fmt.Errorf("could not unmarshal refresh token: %w", err)
	}

	return refreshTokenData.Token, nil
}

func RefreshToken(refreshToken string) (string, string, error) {
	api_url := secrets.CONSTANTS["API_URL"]
	// Set Authorization header
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/token/refresh", api_url), nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+refreshToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("failed to refresh token: %d", resp.StatusCode)
	}
	var response struct {
		Message      string `json:"message"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", "", fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Token, response.RefreshToken, nil
}

// ClearToken removes the token file
func ClearToken() error {
	tokenFile, err := getTokenFilePath()
	if err != nil {
		return err
	}
	return os.Remove(tokenFile)
}

// IsTokenValid checks if the current token is valid and not expired
func IsTokenValid() bool {
	token, err := LoadToken()
	if err != nil || token == "" {
		return false
	}

	// Parse token to check expiration
	claims := &jwt.RegisteredClaims{}
	_, _, err = jwt.NewParser().ParseUnverified(token, claims)

	if err != nil {
		return false
	}

	return claims.ExpiresAt != nil && claims.ExpiresAt.After(time.Now().Add(-30*time.Second))
}

// getTokenFilePath returns the path for the token file
func getTokenFilePath() (string, error) {
	fileDir, err := utils.GetUserDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(fileDir, "auth_token.json"), nil
}

func getRefreshTokenFilePath() (string, error) {
	fileDir, err := utils.GetUserDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(fileDir, "refresh_token.json"), nil
}
