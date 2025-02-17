package model

import "time"

type Conversation struct {
	ID       string    `json:"id"`
	Prompt   Prompt    `json:"prompt"`
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
