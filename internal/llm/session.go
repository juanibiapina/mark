package llm

type Session struct {
	Messages []Message `json:"messages"`

	streaming      bool
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

func (session *Session) IsStreaming() bool {
	return session.streaming
}

func (session *Session) StartStreaming() {
	session.streaming = true
}

func (session *Session) CancelStreaming() {
	if !session.streaming {
		return
	}
	session.streaming = false
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
