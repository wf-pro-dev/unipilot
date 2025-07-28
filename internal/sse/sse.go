package sse

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type SSE struct {
	events     chan Event
	errors     chan error
	mu         sync.Mutex
	ctx        context.Context
	cancelFunc context.CancelFunc
}

type Event struct {
	Type string
	Data json.RawMessage
}

func NewSSE() *SSE {
	ctx, cancel := context.WithCancel(context.Background())
	return &SSE{
		events:     make(chan Event, 1), // Buffered channel
		errors:     make(chan error, 1), // Buffered channel
		ctx:        ctx,
		cancelFunc: cancel,
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
		select {
		case <-c.ctx.Done():
			// If the context is cancelled, exit the connection loop.
			return
		default:
			log.Println("[SSEClient] Attempting to establish SSE connection...")
			err := c.establishAndStream(httpClient)
			if err != nil {
				// Don't push to error channel if it was a graceful shutdown.
				if c.ctx.Err() == nil {
					log.Printf("[SSEClient] Connection error: %v. Retrying in 5 seconds...", err)
					c.errors <- err
				}
			}

			// Wait before retrying, but exit immediately if cancelled.
			select {
			case <-time.After(5 * time.Second):
				// Continue to the next iteration of the loop.
			case <-c.ctx.Done():
				// Exit immediately.
				return
			}
		}
	}
}

func (c *SSE) establishAndStream(httpClient *http.Client) error {
	req, err := http.NewRequestWithContext(c.ctx, "GET", "https://newsroom.dedyn.io/acc-homework/events", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	log.Println("[SSEClient] Connection established. Streaming events...")
	reader := bufio.NewReader(resp.Body)
	for {
		// Check for context cancellation before each read.
		if c.ctx.Err() != nil {
			return c.ctx.Err()
		}

		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return errors.New("server closed connection (EOF)")
			}
			return fmt.Errorf("error reading from stream: %w", err)
		}

		if bytes.HasPrefix(line, []byte("data:")) {
			data := bytes.TrimSpace(bytes.TrimPrefix(line, []byte("data:")))
			if len(data) > 0 {
				c.events <- Event{Data: data}
			}
		}
	}
}

// Events returns the read-only channel for receiving events.
func (c *SSE) Events() <-chan Event {
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
