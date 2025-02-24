package app

import (
	"mark/pkg/model"
)

type (
	threadEntriesMsg []model.ThreadEntry
	partialMessage   string
	replyMessage     string
	threadMsg        struct{ thread model.Thread }
	commitMsg        string
	errMsg           struct{ err error }
)
