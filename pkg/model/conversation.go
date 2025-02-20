package model

import "time"

// ConversationEntry is an entry in the conversation list.
type ConversationEntry struct {
	ID string
}

type Conversation struct {
	ID        string    `json:"id"`
	Messages  []Message `json:"messages"`
	CreatedAt time.Time `json:"created_at"`
}

func MakeConversation() Conversation {
	t := time.Now()

	return Conversation{
		ID:        t.Format(time.RFC3339Nano),
		CreatedAt: t,
	}
}

func (c *Conversation) AddMessage(m Message) {
	c.Messages = append(c.Messages, m)
}
