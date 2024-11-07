package service

import (
	"example.com/chat_app/chat_service/structs"
)

func MapRoomEntityToDto(room *structs.ChatRoomEntity) *structs.RoomDto {
	members := make([]structs.UserDto, 0)
	for _, member := range room.Users {
		members = append(members, *MapUserPermissionsToDto(&member))
	}
	return &structs.RoomDto{
		Id:      room.Id,
		Members: members,
	}
}

func MapUserPermissionsToDto(user *structs.UserPermissions) *structs.UserDto {
	return &structs.UserDto{
		Id:   user.UserId,
		Role: user.Role.String(),
	}
}
