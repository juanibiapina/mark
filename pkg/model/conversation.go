package model

type Conversation struct {
	prompt   Prompt
	messages []Message
}

func MakeConversation() Conversation {
	return Conversation{}
}

func (c *Conversation) AddMessage(m Message) {
	c.messages = append(c.messages, m)
}

func (c *Conversation) Messages() []Message {
	return c.messages
}

func (c *Conversation) SetPrompt(p Prompt) {
	c.prompt = p
}

func (c *Conversation) Prompt() Prompt {
	return c.prompt
}
