package notifications

import (
	"context"
	"log"

	"unipilot/internal/models"
	"unipilot/internal/models/notifications"
	"unipilot/internal/server"
	sse "unipilot/internal/server/sse"
)

type Server struct {
	UnimplementedNotificationsServiceServer
	SSE *sse.SSEServer
}

func (s *Server) SendNotification(ctx context.Context, notification *Notification) (*NotificationResponse, error) {
	log.Printf("[Notifications] Sending notification: %v", notification)
	err := s.SSE.SendNotification(
		uint(notification.UserId),
		uint(notification.SenderId),
		models.Entity(notification.Entity),
		uint(notification.EntityId),
		notifications.NotificationType(notification.Type),
		notification.Title,
		notification.Message,
		notification.Action,
		notification.Data,
	)
	server.LogDebug(ctx, "Notification sent: ",
		"title", notification.Title,
		"sender_id", notification.SenderId,
		"user_id", notification.UserId,
		"tags", []string{"SSE", "GRPC", "NOTIFICATION"},
	)
	return &NotificationResponse{Success: err == nil}, err
}

func (s *Server) Heartbeat(ctx context.Context, message *Message) (*Message, error) {
	server.LogDebug(ctx, "Heartbeat from client: ",
		"body", message.Body,
		"tags", []string{"GRPC", "HEARTBEAT"},
	)
	return &Message{Body: "heartbeat received"}, nil
}
