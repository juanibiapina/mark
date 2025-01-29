package app

import tea "github.com/charmbracelet/bubbletea/v2"

type (
	replyMessage   string
	partialMessage string
)

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func message(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}
