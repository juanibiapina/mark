package llm

import "context"

type Llm interface {
	CompleteStreaming(ctx context.Context, c Conversation, pch chan string, ch chan string) error
}

type Conversation interface {
	Messages() []Message
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
