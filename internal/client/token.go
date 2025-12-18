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

	Errors "unipilot/internal/errors"

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
		return Errors.Wrap(err, Errors.AuthTokenInvalid, "Failed to parse token")
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
		return Errors.Wrap(err, Errors.ProcJSONMarshalFailed, "Failed to marshal token")
	}

	return os.WriteFile(tokenFile, data, 0600) // Secure file permissions
}

// SaveRefreshToken saves the refresh token to a secure file
func SaveRefreshToken(token string) error {
	// Parse token to get expiration
	claims := &jwt.RegisteredClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(token, claims)

	if err != nil {
		return Errors.Wrap(err, Errors.AuthTokenInvalid, "Failed to parse token")
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
		return Errors.Wrap(err, Errors.ProcJSONMarshalFailed, "Failed to marshal refresh token")
	}

	return os.WriteFile(refreshTokenFile, data, 0600) // Secure file permissions
}

// LoadToken loads the JWT token from file
func LoadToken() (string, error) {
	tokenFile, err := getTokenFilePath()
	if err != nil {
		return "", Errors.Wrap(err, Errors.FSFileNotFound, "Failed to get token file")
	}

	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		return "", Errors.Wrap(err, Errors.FSFileNotFound, "Failed to get token file") // No token file exists
	}

	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return "", Errors.Wrap(err, Errors.FSOpenFailed, "Failed to read token file")
	}

	var tokenData TokenData
	if err := json.Unmarshal(data, &tokenData); err != nil {
		return "", Errors.Wrap(err, Errors.ProcJSONUnmarshalFailed, "Failed to unmarshal token")
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
		return "", Errors.Wrap(err, Errors.FSOpenFailed, "Failed to read refresh token file")
	}
	var refreshTokenData TokenData
	if err := json.Unmarshal(data, &refreshTokenData); err != nil {
		return "", Errors.Wrap(err, Errors.ProcJSONUnmarshalFailed, "Failed to unmarshal refresh token")
	}

	return refreshTokenData.Token, nil
}

func RefreshToken(refreshToken string) (string, string, error) {
	api_url := secrets.CONSTANTS["API_URL"]
	// Set Authorization header
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/auth/refresh-token", api_url), nil)
	if err != nil {
		return "", "", Errors.Wrap(err, Errors.ReqBodyInvalid, "Failed to create request")
	}

	req.Header.Set("Authorization", "Bearer "+refreshToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", Errors.Wrap(err, Errors.NetworkConnectionFailed, "Failed to refresh token")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", Errors.Wrap(err, Errors.ClientRequestFailed, "Failed to refresh token")
	}
	var response struct {
		Message      string `json:"message"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", "", Errors.Wrap(err, Errors.ProcJSONUnmarshalFailed, "Failed to decode response")
	}

	return response.Token, response.RefreshToken, nil
}

// ClearToken removes the token file
func ClearToken() error {
	tokenFile, err := getTokenFilePath()
	if err != nil {
		return Errors.Wrap(err, Errors.FSFileNotFound, "Failed to get token file")
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
		return "", Errors.Wrap(err, Errors.FSDirFailed, "Error getting user directory")
	}
	_, err = os.Stat(filepath.Join(fileDir, "auth_token.json"))
	if err != nil {
		credentialsFile, err := utils.GetCredentialFile()
		if err != nil {
			return "", Errors.Wrap(err, Errors.FSFileNotFound, "Credential file not found")
		}
		os.Remove(credentialsFile)
		return "", Errors.Wrap(err, Errors.FSFileNotFound, "Token file not found")
	}
	return filepath.Join(fileDir, "auth_token.json"), nil
}

func getRefreshTokenFilePath() (string, error) {
	fileDir, err := utils.GetUserDir()
	if err != nil {
		return "", Errors.Wrap(err, Errors.FSDirFailed, "Error getting user directory")
	}
	_, err = os.Stat(filepath.Join(fileDir, "refresh_token.json"))
	if err != nil {
		return "", Errors.Wrap(err, Errors.FSFileNotFound, "Refresh token file not found")
	}
	return filepath.Join(fileDir, "refresh_token.json"), nil
}
