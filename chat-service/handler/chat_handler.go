package handler

import (
	"encoding/json"
	"net/http"

	"example.com/chat_app/chat_service/service"
)

type ChatHandler struct {
	chatService *service.ChatService
}

func NewChatHandler(cs *service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: cs}
}

func (ch *ChatHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	roomId := r.PathValue("roomId")
	userId := r.Header.Get("X-User-Id")

	messages, err := ch.chatService.ListMessages(r.Context(), roomId, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJsonResponse(w, messages)
}

// GetMessagesSummary handles requests to summarize chat messages.
// It expects a JSON array of strings in the request body and passes them
// along with roomId and userId to the service layer for processing.
func (ch *ChatHandler) GetMessagesSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	roomId := r.PathValue("roomId")
	userId := r.Header.Get("X-User-Id")

	var messages []string

	if err := json.NewDecoder(r.Body).Decode(&messages); err != nil {
		http.Error(w, "Invalid request body. Expected a JSON array of strings.", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	messagesSummary, err := ch.chatService.GetMessagesSummary(ctx, roomId, userId, messages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJsonResponse(w, messagesSummary)
}
