package app

import (
	"mark/pkg/model"
)

type (
	conversationEntriesMsg []model.ConversationEntry
	partialMessage         string
	replyMessage           string
	conversationMsg        struct{ conversation model.Conversation }
)

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }
