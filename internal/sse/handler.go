package sse

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
	"unipilot/internal/client"
	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/secrets"
)

type SSE struct {
	events     chan models.Message
	errors     chan error
	mu         sync.Mutex
	ctx        context.Context
	cancelFunc context.CancelFunc

	// retries
	retryCount int
	baseDelay  time.Duration
	maxDelay   time.Duration
	maxRetries int
}

func NewSSE() *SSE {
	ctx, cancel := context.WithCancel(context.Background())
	return &SSE{
		events:     make(chan models.Message, 1), // Buffered channel
		errors:     make(chan error, 1),          // Buffered channel
		ctx:        ctx,
		cancelFunc: cancel,

		retryCount: 0,
		baseDelay:  1 * time.Second,
		maxDelay:   60 * time.Second,
		maxRetries: -1,
	}
}

// Connect now accepts a context to handle cancellation.
func (c *SSE) Connect(httpClient *http.Client) {
	defer func() {
		close(c.events)
		close(c.errors)
		log.Println("[SSEClient] Connection loop terminated and channels closed.")
	}()

	for {
		log.Printf("[DEBUG] Loop iteration - retryCount: %d", c.retryCount)
		select {
		case <-c.ctx.Done():
			return
		default:
			err := c.establishAndStream(httpClient)

			if err != nil {
				if c.ctx.Err() == nil {
					// Increment retry count on failure
					c.incrementRetry()

					// Calculate backoff delay
					delay := c.calculateBackoff()

					log.Printf("[SSEClient] Connection error: %v. Retry #%d in %v...",
						err, c.retryCount, delay)

					c.errors <- err

					// Wait before retrying
					select {
					case <-time.After(delay):
					case <-c.ctx.Done():
						return
					}
				} else {
					return
				}

			}
		}
	}
}

func (c *SSE) establishAndStream(httpClient *http.Client) error {

	base_url := secrets.CONSTANTS["BASE_URL"]

	url := fmt.Sprintf("%s/unipilot/sse/v1", base_url)

	req, err := http.NewRequestWithContext(c.ctx, "GET", url, nil)
	if err != nil {
		return errors.Wrap(err, errors.ClientRequestFailed, "Failed to create SSE request")
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	// Add headers to prevent connection reuse issues
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Pragma", "no-cache")

	// Set auth header with token refresh
	if err := client.SetAuthHeaderRequest(req); err != nil {
		return errors.Wrap(err, errors.ClientRequestFailed, "Failed to set auth header")
	}

	// Close connection on completion to prevent reuse
	req.Close = true

	resp, err := httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, errors.NetworkConnectionFailed, "HTTP request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return errors.NewAppError(errors.AuthUnauthorized, "Unauthorized", nil)
	}

	if resp.StatusCode != http.StatusOK {
		return errors.NewAppError(errors.ClientResponseInvalid, "Received non-200 status code", nil)
	}

	// Reset retry count when successful connection is established
	c.resetRetry()

	reader := bufio.NewReader(resp.Body)
	for {
		// Check for context cancellation before each read.
		if c.ctx.Err() != nil {
			return c.ctx.Err()
		}

		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return errors.NewAppError(errors.NetworkConnectionFailed, "Server closed connection (EOF)", nil)
			}
			log.Printf("[SSEClient] Error reading from SSE stream: %v %v", line, err)
			return errors.Wrap(err, errors.FSStreamFailed, "Error reading from SSE stream")
		}

		if bytes.HasPrefix(line, []byte("data:")) {
			data := bytes.TrimSpace(bytes.TrimPrefix(line, []byte("data:")))
			if len(data) > 0 {
				var message models.Message
				if err := json.Unmarshal(data, &message); err != nil {
					return errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to unmarshal message")
				}
				c.events <- message
			}
		}
	}
}

// Events returns the read-only channel for receiving events.
func (c *SSE) Events() <-chan models.Message {
	return c.events
}

// Errors returns the read-only channel for receiving errors.
func (c *SSE) Errors() <-chan error {
	return c.errors
}

// StopListener signals the SSE connection to close.
func (c *SSE) StopListener() {
	if c.cancelFunc != nil {
		log.Println("Signaling SSE client to disconnect...")
		c.cancelFunc() // This cancels the context passed to sseClient.Connect
	}
}

func (c *SSE) calculateBackoff() time.Duration {

	delay := c.baseDelay * (1 << c.retryCount)

	// Cap at maximum
	if delay > c.maxDelay {
		delay = c.maxDelay
	}

	// Add random jitter: ±25% of delay
	jitter := time.Duration(rand.Int63n(int64(delay / 2))) // 0 to 50% of delay
	delay = delay - (delay / 4) + jitter                   // delay ±25%

	return delay
}

func (c *SSE) incrementRetry() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.retryCount++
}

func (c *SSE) resetRetry() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.retryCount = 0
}
