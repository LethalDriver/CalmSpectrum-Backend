package structs

import "encoding/json"

type RoomDto struct {
	Id      string    `json:"id"`
	Name    string    `json:"name"`
	Members []UserDto `json:"members"`
}

type RoomCreateDto struct {
	Name string `json:"name"`
}

type UserDto struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type SeenMessage struct {
	MessageId string      `json:"messageId"`
	SeenBy    UserDetails `json:"seenBy"`
}

type DeleteMessage struct {
	MessageId string      `json:"messageId"`
	SentBy    UserDetails `json:"sentBy"`
}

type WsMessage struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data"`
}

type MessageDto struct {
	SentBy  string `json:"sentBy"`
	Content string `json:"content"`
}
