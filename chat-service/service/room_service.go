package service

import (
	"context"
	"errors"
	"fmt"

	"example.com/chat_app/chat_service/repository"
	"example.com/chat_app/chat_service/structs"
	"go.mongodb.org/mongo-driver/mongo"
)

type ChatRoomRepository interface {
	CreateRoom(ctx context.Context, name string) (*structs.ChatRoomEntity, error)
	GetRoom(ctx context.Context, id string) (*structs.ChatRoomEntity, error)
	DeleteRoom(ctx context.Context, id string) error
	AddMessageToRoom(ctx context.Context, roomId string, message *structs.Message) error
	InsertSeenBy(ctx context.Context, roomId string, messageId string, userId string) error
	DeleteMessage(ctx context.Context, roomId string, messageId string) error
	InsertUserIntoRoom(ctx context.Context, roomId string, user structs.UserPermissions) error
	DeleteUserFromRoom(ctx context.Context, roomId string, userId string) error
	GetUsersPermissions(ctx context.Context, roomId string, userId string) (*structs.UserPermissions, error)
	ChangeUserRole(ctx context.Context, roomId string, userId string, role structs.Role) error
	GetUnseenMessages(ctx context.Context, roomId, userId string) ([]structs.Message, error)
	GetUsersRooms(ctx context.Context, userId string) ([]structs.ChatRoomEntity, error)
	GetMessageById(ctx context.Context, id string) (*structs.Message, error)
}

// ErrInsufficientPermissions is an error indicating that the user does not have sufficient permissions.
var ErrInsufficientPermissions = errors.New("insufficient permissions")

// RoomService provides methods to manage chat rooms and handle user permissions.
type RoomService struct {
	repo     ChatRoomRepository
	userRepo *repository.UserRepository
}

// NewRoomService creates a new instance of RoomService.
func NewRoomService(repo ChatRoomRepository, users *repository.UserRepository) *RoomService {
	return &RoomService{repo: repo,
		userRepo: users}
}

// GetRoomDto retrieves a chat room DTO if the user belongs to the room.
func (s *RoomService) GetRoomDto(ctx context.Context, roomId string, userId string) (*structs.RoomDto, error) {
	room, err := s.repo.GetRoom(ctx, roomId)
	if err != nil {
		return nil, err
	}
	if !checkIfUserBelongsToRoom(room, userId) {
		return nil, ErrInsufficientPermissions
	}
	roomDto := MapRoomEntityToDto(room)
	return roomDto, nil
}

// CreateRoom creates a new chat room and adds the creating user as an admin.
func (s *RoomService) CreateRoom(ctx context.Context, userId, name string) (*structs.ChatRoomEntity, error) {
	room, err := s.repo.CreateRoom(ctx, name)
	if err != nil {
		return nil, err
	}
	err = s.AddAdminToRoom(ctx, room.Id, userId)
	if err != nil {
		return nil, err
	}
	return room, nil
}

// DeleteRoom deletes a chat room if the user has admin privileges.
func (s *RoomService) DeleteRoom(ctx context.Context, roomId string, userId string) error {
	if err := s.validateAdminPrivileges(ctx, roomId, userId); err != nil {
		return err
	}
	return s.repo.DeleteRoom(ctx, roomId)
}

// AddUserToRoom adds a user to a chat room if the requesting user has admin privileges.
func (s *RoomService) AddUserToRoom(ctx context.Context, roomId string, newUserId string, addingUserId string) error {
	if err := s.validateAdminPrivileges(ctx, roomId, addingUserId); err != nil {
		return err
	}
	username, err := s.userRepo.GetUsernameById(ctx, newUserId)
	if err != nil {
		return err
	}
	userPermissions := structs.UserPermissions{
		UserId:   newUserId,
		Username: username,
		Role:     structs.Member,
	}
	return s.repo.InsertUserIntoRoom(ctx, roomId, userPermissions)
}

// AddUsersToRoom adds multiple users to a chat room if the requesting user has admin privileges.
func (s *RoomService) AddUsersToRoom(ctx context.Context, roomId string, newUsers []string, addingUserId string) error {
	if err := s.validateAdminPrivileges(ctx, roomId, addingUserId); err != nil {
		return ErrInsufficientPermissions
	}
	fmt.Printf("Adding users to room %s: %v\n", roomId, newUsers)
	for _, userId := range newUsers {
		username, err := s.userRepo.GetUsernameById(ctx, userId)
		if err != nil {
			return err
		}
		permission := structs.UserPermissions{
			UserId:   userId,
			Role:     structs.Member,
			Username: username,
		}
		fmt.Printf("Adding user %v", permission)
		err = s.repo.InsertUserIntoRoom(ctx, roomId, permission)
		if err != nil {
			return err
		}
	}
	return nil
}

// RemoveUserFromRoom removes a user from a chat room if the requesting user has admin privileges.
func (s *RoomService) RemoveUserFromRoom(ctx context.Context, roomId, requestingUserId, removedUserId string) error {
	if err := s.validateAdminPrivileges(ctx, roomId, requestingUserId); err != nil {
		return err
	}
	return s.repo.DeleteUserFromRoom(ctx, roomId, removedUserId)
}

// LeaveRoom allows a user to leave a chat room.
func (s *RoomService) LeaveRoom(ctx context.Context, roomId, userId string) error {
	return s.repo.DeleteUserFromRoom(ctx, roomId, userId)
}

// PromoteUser promotes a user to admin in a chat room if the requesting user has admin privileges.
func (s *RoomService) PromoteUser(ctx context.Context, roomId string, promotingUserId, promotedUserId string) error {
	if err := s.validateAdminPrivileges(ctx, roomId, promotingUserId); err != nil {
		return err
	}
	return s.repo.ChangeUserRole(ctx, roomId, promotedUserId, structs.Admin)
}

// DemoteUser demotes a user to member in a chat room if the requesting user has admin privileges.
func (s *RoomService) DemoteUser(ctx context.Context, roomId, demotingUserId, demotedUserId string) error {
	if err := s.validateAdminPrivileges(ctx, roomId, demotingUserId); err != nil {
		return err
	}
	return s.repo.ChangeUserRole(ctx, roomId, demotedUserId, structs.Member)
}

// AddAdminToRoom adds an admin to a chat room.
func (s *RoomService) AddAdminToRoom(ctx context.Context, roomId string, userId string) error {
	username, err := s.userRepo.GetUsernameById(ctx, userId)
	if err != nil {
		return err
	}
	userPermission := structs.UserPermissions{
		UserId:   userId,
		Username: username,
		Role:     structs.Admin,
	}
	return s.repo.InsertUserIntoRoom(ctx, roomId, userPermission)
}

// validateAdminPrivileges checks if a user has admin privileges in a chat room.
func (s *RoomService) validateAdminPrivileges(ctx context.Context, roomId, userId string) error {
	userPermissions, err := s.repo.GetUsersPermissions(ctx, roomId, userId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrInsufficientPermissions
		}
		return err
	}
	if userPermissions.Role != structs.Admin {
		return ErrInsufficientPermissions
	}
	return nil
}

func (s *RoomService) ListRoomsForUser(ctx context.Context, userId string) ([]structs.RoomDto, error) {
	rooms, err := s.repo.GetUsersRooms(ctx, userId)
	if err != nil {
		return nil, err
	}

	roomDtos := make([]structs.RoomDto, 0, len(rooms))
	for _, room := range rooms {
		roomDto := MapRoomEntityToDto(&room)
		roomDtos = append(roomDtos, *roomDto)
	}

	return roomDtos, nil
}
