package app

import (
	"mark/pkg/model"
)

type (
	threadEntriesMsg          []model.ThreadEntry
	partialMessage            string
	replyMessage              string
	threadMsg                 struct{ thread model.Thread }
	pullRequestDescriptionMsg string
)

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }
