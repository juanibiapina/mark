package domain

import "slices"

type ContextItem string

// implement list.Item interface
func (i ContextItem) FilterValue() string { return "" }

type Session struct {
	context []ContextItem
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
	session.context = append(session.context, ContextItem(content))
}

func (session *Session) DeleteContextItem(index int) {
	session.context = slices.Delete(session.context, index, index+1)
}

func (session *Session) Context() []ContextItem {
	return session.context
}

func (session *Session) ContextMessage() string {
	var message string

	for _, v := range session.context {
		message += string(v)
		message += "\n\n"
	}

	return message
}
