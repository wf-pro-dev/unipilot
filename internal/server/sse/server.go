package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"unipilot/internal/models"
	"unipilot/internal/models/notifications"
	"unipilot/internal/server"
)

type SSEClient struct {
	UserID     uint
	Messages   chan []byte
	Connected  bool
	LastActive time.Time
}

type SSEServer struct {
	clients map[uint]*SSEClient
	mu      sync.RWMutex
}

func NewSSEServer() *SSEServer {
	return &SSEServer{
		clients: make(map[uint]*SSEClient),
	}
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`))
}

func StartSSEServer() *SSEServer {
	sseServer := NewSSEServer()
	http.HandleFunc("/health", HealthHandler)
	http.HandleFunc("/unipilot/sse/v1", server.AuthMiddleware(sseServer.SSEHandler))
	log.Println("SSE server listening on :3000...")
	go func() {
		if err := http.ListenAndServe(":3000", nil); err != nil {
			log.Fatalf("SSE server error: %v", err)
		}
	}()

	return sseServer
}

func (s *SSEServer) AddClient(userID uint) *SSEClient {
	s.mu.Lock()
	defer s.mu.Unlock()

	client := &SSEClient{
		UserID:    userID,
		Messages:  make(chan []byte, 100),
		Connected: true,
	}

	server.LogDebug(context.Background(), "New SSE user id : ",
		"user_id", userID, "tags",
		[]string{"SSE", "NEW_USER"},
	)
	s.clients[userID] = client
	return client
}

func (s *SSEServer) RemoveClient(userID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if client, ok := s.clients[userID]; ok {
		close(client.Messages)
		delete(s.clients, userID)
	}
}

func (s *SSEServer) SendToUser(userID uint, message []byte) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.logActiveClients()
	if client, ok := s.clients[userID]; ok {
		select {
		case client.Messages <- message:
			return true
		default:
			// Channel full, client might be slow
			return false
		}
	}
	return false
}

func (s *SSEServer) Broadcast(message []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, client := range s.clients {
		select {
		case client.Messages <- message:
		default:
			// Skip if channel is full
		}
	}
}

func (s *SSEServer) logActiveClients() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	server.LogDebug(context.Background(), "Active Clients: ",
		"count", len(s.clients),
		"tags", []string{"SSE", "ACTIVE_CLIENTS"},
	)
}

type noTimeoutWriter struct {
	http.ResponseWriter
}

func (w *noTimeoutWriter) Write(p []byte) (int, error) {
	// Disable write timeout
	if conn, _, err := w.ResponseWriter.(http.Hijacker).Hijack(); err == nil {
		conn.SetWriteDeadline(time.Time{})
		conn.Close()
	}
	return w.ResponseWriter.Write(p)
}

func (s *SSEServer) SSEHandler(w http.ResponseWriter, r *http.Request) {

	// Get user from context (set by AuthMiddleware)
	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")

	// Create a flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		return
	}

	// Add client to server
	client := s.AddClient(userID)
	defer func() {
		server.LogDebug(r.Context(), "Removing client: ",
			"tags", []string{"SSE", "REMOVE_CLIENT"},
		)
		s.RemoveClient(userID)
	}()

	// Send initial connection message
	//fmt.Fprintf(w, "event: connected\ndata: %s\n\n", "SSE connection established")
	//flusher.Flush()

	// Keep connection alive and send messages
	heartbeatTicker := time.NewTicker(15 * time.Second)
	defer heartbeatTicker.Stop()

	for {
		select {
		case msg := <-client.Messages:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-heartbeatTicker.C:
			// Send heartbeat to keep connection alive
			client.LastActive = time.Now()
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			server.LogDebug(r.Context(), "Client disconnected: ",
				"tags", []string{"SSE", "DISCONNECTED"},
			)
			return
		}
	}
}
func (s *SSEServer) SendNotification(userID, senderID uint, entity models.Entity, entityID uint, nType notifications.NotificationType, title, message, action, data string) error {
	notification := notifications.LocalNotification{
		SenderID: senderID,
		Entity:   entity,
		EntityID: entityID,
		Type:     nType,
		Action:   action,
		Title:    title,
		Message:  message,
		Data:     data,
	}

	jsonData, err := json.Marshal(notification)
	if err != nil {
		log.Printf("[Error] error marshalling notification : %v ", err)
		return err
	}

	s.SendToUser(userID, jsonData)
	return nil
}
