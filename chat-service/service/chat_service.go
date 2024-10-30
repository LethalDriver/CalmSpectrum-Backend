package service

import (
	"context"
	"log"
	"time"

	"example.com/chat_app/chat_service/repository"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
)

type ChatService struct {
	roomRepo    repository.ChatRoomRepository
	roomManager RoomManager
}

func NewChatService(roomRepo *repository.MongoChatRoomRepository, roomManager RoomManager) *ChatService {
	return &ChatService{
		roomRepo:    roomRepo,
		roomManager: roomManager,
	}
}

func (s *ChatService) ConnectToRoom(ctx context.Context, roomId, userId, username string, ws *websocket.Conn) error {
	_, err := s.roomRepo.GetRoom(ctx, roomId)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return err
		}
	}
	userDetails := repository.UserDetails{
		Id:       userId,
		Username: username,
	}
	room := s.roomManager.ManageRoom(roomId)
	go room.Run(s)

	go handleConnection(ws, room, userDetails)

	log.Printf("Room: %s running", roomId)
	return nil
}

func (s *ChatService) pumpExistingMessages(conn *Connection, messages []repository.Message) {
	for _, message := range messages {
		conn.sendMessage <- message
	}
}

func (s *ChatService) processAndSaveMessage(ctx context.Context, roomId string, message *repository.Message) (repository.Message, error) {
	message.Id = uuid.New().String()
	message.SentAt = time.Now()
	message.ChatRoomId = roomId
	message.SeenBy = []repository.UserDetails{message.SentBy}
	return *message, s.roomRepo.AddMessageToRoom(ctx, roomId, message)
}
