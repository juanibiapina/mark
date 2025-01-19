package model

type Role int

const (
	User Role = iota
	Assistant
)

type Conversation struct {
	Messages []Message
}

type Message struct {
	Role    Role
	Content string
}
