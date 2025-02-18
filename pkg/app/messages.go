package app

import (
	"mark/pkg/model"
)

type (
	replyMessage           string
	partialMessage         string
	conversationEntriesMsg []model.ConversationEntry
)

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }
