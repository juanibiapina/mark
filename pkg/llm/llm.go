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

type Conversation struct {
	messages []Message
}

func (c *Conversation) AddMessage(m Message) {
	c.messages = append(c.messages, m)
}

func (c *Conversation) Messages() []Message {
	return c.messages
}

func (c *Conversation) Reset() {
	c.messages = []Message{}
}

type Message struct {
	Role    Role
	Content string
}
