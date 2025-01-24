package llm

type Llm interface {
	CompleteStreaming(c *Conversation) error
}

type Role int

const (
	RoleUser Role = iota
	RoleAssistant
)

type Conversation struct {
	messages         []Message
	StreamingMessage *StreamingMessage
}

func (c *Conversation) CancelStreaming() {
	if c.StreamingMessage == nil {
		return
	}

	c.StreamingMessage.Cancel()

	// Add the partial message to the chat history
	c.AddMessage(Message{Role: RoleAssistant, Content: c.StreamingMessage.Content})

	c.StreamingMessage = nil
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
