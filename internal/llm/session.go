package llm

import "strings"

type Session struct {
	context []string
	prompt  string
	reply   string
}

func MakeSession() Session {
	return Session{}
}

func (session *Session) Prompt() string {
	return session.prompt
}

func (session *Session) SetPrompt(content string) {
	session.prompt = content
}

func (session *Session) AppendChunk(chunk string) {
	session.reply += chunk
}

func (session *Session) FinishStreaming(msg string) {
	session.reply = msg
}

func (session *Session) ClearReply() {
	session.reply = ""
}

func (session *Session) Reply() string {
	return session.reply
}

func (session *Session) AddContext(content string) {
	session.context = append(session.context, content)
}

func (session *Session) ContextMessage() string {
	return strings.Join(session.context, "\n\n")
}
