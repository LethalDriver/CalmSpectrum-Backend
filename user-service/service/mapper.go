package service

import (
	"example.com/chat_app/user_service/structs"
)

// MapUserDtoToEntity maps UserDto to UserEntity
func MapUserEntityToDto(user *structs.UserEntity) *structs.UserDto {
	return &structs.UserDto{
		Id:       user.Id,
		Username: user.Username,
		Email:    user.Email,
	}
}

func MapUserEntityToPublicDto(user *structs.UserEntity) *structs.PublicUserDto {
	return &structs.PublicUserDto{
		Id:       user.Id,
		Username: user.Username,
	}
}
