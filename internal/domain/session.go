package domain

type Session struct {
	context *Context
	prompt  string
	reply   string
}

func MakeSession() Session {
	return Session{
		context: NewContext(),
	}
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

func (session *Session) SetReply(msg string) {
	session.reply = msg
}

func (session *Session) ClearReply() {
	session.reply = ""
}

func (session *Session) Reply() string {
	return session.reply
}

func (session *Session) Context() *Context {
	return session.context
}
