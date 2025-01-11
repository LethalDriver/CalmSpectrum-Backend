package service

import (
	"context"
	"errors"
	"log"
	"time"

	"example.com/chat_app/chat_service/client"
	"example.com/chat_app/chat_service/structs"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
)

// ErrRoomNotFound is an error indicating that the chat room was not found.
var ErrRoomNotFound = errors.New("room not found")

// ChatService provides methods to manage chat rooms and handle connections.
type ChatService struct {
	roomRepo    ChatRoomRepository
	roomManager RoomManager
	ai          *client.AiAssistantClient
}

// NewChatService creates a new instance of ChatService.
func NewChatService(roomRepo ChatRoomRepository, roomManager RoomManager, ai *client.AiAssistantClient) *ChatService {
	return &ChatService{
		roomRepo:    roomRepo,
		roomManager: roomManager,
		ai:          ai,
	}
}

// ConnectToRoom connects a user to a chat room and starts handling the connection in a separate goroutine.
func (s *ChatService) ConnectToRoom(ctx context.Context, roomId, userId string, ws *websocket.Conn) {
	memoryRoom := s.roomManager.ManageRoom(roomId)
	go memoryRoom.Run(s)

	userDetails := structs.UserDetails{
		Id: userId,
	}
	go handleConnection(ws, memoryRoom, userDetails)

	log.Printf("Room: %s running", roomId)
}

// ValidateConnection validates if a user can connect to a chat room.
func (s *ChatService) ValidateConnection(ctx context.Context, roomId, userId string) error {
	dbRoom, err := s.roomRepo.GetRoom(ctx, roomId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrRoomNotFound
		}
		return err
	}
	if !checkIfUserBelongsToRoom(dbRoom, userId) {
		return ErrInsufficientPermissions
	}
	return nil
}

func (s *ChatService) GetMessagesSummary(ctx context.Context, roomId, userId string) (*structs.MessagesSummary, error) {
	room, err := s.roomRepo.GetRoom(ctx, roomId)
	if err != nil {
		return nil, err
	}
	if !checkIfUserBelongsToRoom(room, userId) {
		return nil, ErrInsufficientPermissions
	}

	unreadMessages, err := s.roomRepo.GetUnseenMessages(ctx, roomId, userId)
	if err != nil {
		return nil, err
	}

	var messageDtos []structs.MessageDto

	for _, message := range unreadMessages {
		messageDtos = append(messageDtos, s.mapMessageToMessageDto(message))
	}

	summary, err := s.ai.GetMessagesSummary(ctx, messageDtos)
	if err != nil {
		return nil, err
	}

	return summary, nil
}

// checkIfUserBelongsToRoom checks if a user belongs to a chat room.
func checkIfUserBelongsToRoom(room *structs.ChatRoomEntity, userId string) bool {
	for _, user := range room.Users {
		if user.UserId == userId {
			return true
		}
	}
	return false
}

// pumpExistingMessages sends persisted chat room messages to a new connection.
func (s *ChatService) pumpExistingMessages(conn *Connection, messages []structs.Message) {
	for _, message := range messages {
		conn.sendMessage <- message
	}
}

// processAndSaveMessage processes and saves a message to a chat room.
func (s *ChatService) processAndSaveMessage(ctx context.Context, roomId string, message *structs.Message) (structs.Message, error) {
	message.Id = uuid.New().String()
	message.SentAt = time.Now()
	message.ChatRoomId = roomId
	message.SeenBy = []string{message.SentBy}
	return *message, s.roomRepo.AddMessageToRoom(ctx, roomId, message)
}

func (s *ChatService) mapMessageToMessageDto(message structs.Message) structs.MessageDto {
	return structs.MessageDto{
		Content: message.Content,
		SentBy:  message.SentBy,
	}
}
