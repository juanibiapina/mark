package llm

type Session struct {
	Messages       []Message `json:"messages"`
	partialMessage string
}

func MakeSession() Session {
	return Session{}
}

func (session *Session) AddMessage(m Message) {
	session.Messages = append(session.Messages, m)
}

func (session *Session) AppendChunk(chunk string) {
	session.partialMessage += chunk
}

func (session *Session) FinishStreaming(msg string) {
	session.AddMessage(Message{
		Role:    RoleAssistant,
		Content: msg,
	})
	session.partialMessage = ""
}

func (session *Session) AcceptPartialMessage() {
	if session.partialMessage == "" {
		return
	}
	session.AddMessage(Message{
		Role:    RoleAssistant,
		Content: session.partialMessage,
	})
	session.partialMessage = ""
}

func (session *Session) PartialMessage() string {
	return session.partialMessage
}
