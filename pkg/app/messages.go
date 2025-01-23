package app

import tea "github.com/charmbracelet/bubbletea"

type focusInputMsg struct{}
type focusConversationMsg struct{}
type input struct{
	content string
}
type replyMessage string
type partialMessage string

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func message(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}
