package model

type Conversation struct {
	context          map[string]string
	messages         []Message
}

func MakeConversation() Conversation {
	return Conversation{
		context:  make(map[string]string),
	}
}

func (c *Conversation) AddMessage(m Message) {
	c.messages = append(c.messages, m)
}

func (c *Conversation) SetContext(k, v string) {
	c.context[k] = v
}

func (c *Conversation) Messages() []Message {
	return c.messages
}

func (c *Conversation) Context() map[string]string {
	return c.context
}
