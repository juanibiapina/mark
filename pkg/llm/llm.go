package llm

type Llm interface {
	CompleteStreaming(c *Conversation) error
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
