package model

type Llm interface {
	CompleteStreaming(c *Conversation, s *StreamingMessage) error
	CompleteStructured(c *Conversation, rs ResponseSchema, v interface{}) error
}

type Role int

const (
	RoleUser Role = iota
	RoleAssistant
)

type Message struct {
	Role    Role
	Content string
}
