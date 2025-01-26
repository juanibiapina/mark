package llm

type Conversation struct {
	context          []string
	variables        map[string]string
	messages         []Message
}

func MakeConversation() Conversation {
	return Conversation{
		variables: make(map[string]string),
	}
}

func (c *Conversation) AddMessage(m Message) {
	c.messages = append(c.messages, m)
}

func (c *Conversation) AddContext(context string) {
	c.context = append(c.context, context)
}

func (c *Conversation) AddVariable(key, value string) {
	c.variables[key] = value
}

func (c *Conversation) Messages() []Message {
	return c.messages
}
