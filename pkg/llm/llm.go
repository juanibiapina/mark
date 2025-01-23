package llm

import "context"

type Llm interface {
	CompleteStreaming(ctx context.Context, messages []Message, pch chan string, ch chan string) error
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
