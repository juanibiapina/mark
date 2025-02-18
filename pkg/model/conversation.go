package model

import "time"

// ConversationEntry is an entry in the conversation list.
type ConversationEntry struct {
	ID string
}

type Conversation struct {
	ID       string    `json:"id"`
	Messages []Message `json:"messages"`
}

func MakeConversation() Conversation {
	return Conversation{
		ID: time.Now().Format(time.RFC3339Nano),
	}
}

func (c *Conversation) AddMessage(m Message) {
	c.Messages = append(c.Messages, m)
}
