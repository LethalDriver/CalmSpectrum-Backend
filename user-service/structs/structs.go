package structs

type UserDto struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type PublicUserDto struct {
	Id       string `json:"id"`
	Username string `json:"username"`
}
