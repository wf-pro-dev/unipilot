package notifications

import (
	"context"
	"fmt"
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
	server.PrintLOG([]string{"SSE", "GRPC"},
		fmt.Sprintf("Notification %s sent from %v to %v", notification.Title, notification.SenderId, notification.UserId))
	return &NotificationResponse{Success: err == nil}, err
}

func (s *Server) Heartbeat(ctx context.Context, message *Message) (*Message, error) {
	server.PrintLOG([]string{"GRPC"}, fmt.Sprintf("Heartbeat from client: %s", message.Body))
	return &Message{Body: "heartbeat received"}, nil
}
