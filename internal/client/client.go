package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/net/publicsuffix"
)

const (
	appName = "acc-homework" // Change this to your application name
)

type CookieJar struct {
	cookies []*http.Cookie
}

func (j *CookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.cookies = cookies
}

func (j *CookieJar) Cookies(u *url.URL) []*http.Cookie {
	return j.cookies
}

func NewClient() (*http.Client, error) {
	cookies, err := LoadCookies()
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Jar: &CookieJar{cookies: cookies},
	}, nil
}

// getCookieFilePath returns the canonical path for the cookie file.
func getCookieFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	// Create app directory and cookies subdirectory
	appDir := filepath.Join(configDir, appName)
	cookiesDir := filepath.Join(appDir, "cookies")

	// Create directories with 0755 permissions (rwx for owner, rx for others)
	if err := os.MkdirAll(cookiesDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(cookiesDir, "cookies.txt"), nil
}

// SaveCookies serializes the cookies from the client's jar and saves them to a file.
func SaveCookies(client *http.Client) error {

	// Check if client or client.Jar is nil
	if client == nil {
		return fmt.Errorf("client is nil")
	}

	if client.Jar == nil {
		return fmt.Errorf("client jar is nil")
	}

	cookieFile, err := getCookieFilePath()
	if err != nil {
		return err
	}

	// Use proper URL instead of nil to avoid nil pointer dereference
	targetURL, _ := url.Parse("https://newsroom.dedyn.io")
	cookies := client.Jar.Cookies(targetURL)

	data, err := json.MarshalIndent(cookies, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal cookies: %w", err)
	}

	return os.WriteFile(cookieFile, data, 0600)
}

// NewClientWithCookies creates a new http.Client and loads cookies from the file.
func NewClientWithCookies() (*http.Client, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}

	// Configure transport to handle connection pooling properly
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	}

	cookieFile, err := getCookieFilePath()
	if err != nil {
		return nil, err
	}
	// It's okay if the cookie file doesn't exist yet.
	if _, err := os.Stat(cookieFile); os.IsNotExist(err) {
		log.Println("No cookie file found, creating new client.")
		return &http.Client{
			Jar:       jar,
			Transport: transport,
		}, nil
	}

	data, err := os.ReadFile(cookieFile)
	if err != nil {
		return nil, fmt.Errorf("could not read cookie file: %w", err)
	}

	var cookies []*http.Cookie
	if err := json.Unmarshal(data, &cookies); err != nil {
		return nil, fmt.Errorf("could not unmarshal cookies: %w", err)
	}
	// Assuming the cookies are for the correct domain.
	if len(cookies) > 0 {
		// We need a URL to set cookies in the jar.
		// This should be the base URL of your service.
		url, _ := url.Parse("https://newsroom.dedyn.io")
		jar.SetCookies(url, cookies)
	}

	return &http.Client{
		Jar:       jar,
		Transport: transport,
	}, nil
}

func LoadCookies() ([]*http.Cookie, error) {
	path, err := getCookieFilePath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No cookies yet
		}
		return nil, err
	}
	defer file.Close()

	var cookies []*http.Cookie
	err = json.NewDecoder(file).Decode(&cookies)
	return cookies, err
}

// ClearCookies removes the cookie file from disk.
func ClearCookies() error {
	cookieFile, err := getCookieFilePath()
	if err != nil {
		return err
	}
	// It's not an error if the file doesn't exist.
	if err := os.Remove(cookieFile); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// NewSSEClient creates an HTTP client specifically configured for SSE connections
func NewSSEClient() (*http.Client, error) {
	// Configure transport specifically for SSE connections
	transport := &http.Transport{
		MaxIdleConns:          50,
		MaxIdleConnsPerHost:   2, // Limit idle connections per host for SSE
		IdleConnTimeout:       30 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		DisableKeepAlives:     false,
		// Disable HTTP/2 for SSE to avoid protocol issues
		ForceAttemptHTTP2: false,
	}

	// Include cookie jar for authentication
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}

	// Load cookies if they exist
	cookieFile, err := getCookieFilePath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(cookieFile); !os.IsNotExist(err) {
		data, err := os.ReadFile(cookieFile)
		if err == nil {
			var cookies []*http.Cookie
			if err := json.Unmarshal(data, &cookies); err == nil && len(cookies) > 0 {
				url, _ := url.Parse("https://newsroom.dedyn.io")
				jar.SetCookies(url, cookies)
			}
		}
	}

	return &http.Client{
		Jar:       jar,
		Transport: transport,
		Timeout:   0, // No timeout for SSE connections
	}, nil
}
