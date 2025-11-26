package server

import (
	"net/http"

	"github.com/google/uuid"
)

func GenerateChatIDHandler(w http.ResponseWriter, r *http.Request) {
	var chatID uuid.UUID
	chatID = uuid.New()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"chat_id": "` + chatID.String() + `"}`))
}
