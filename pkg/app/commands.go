package app

import (
	tea "github.com/charmbracelet/bubbletea/v2"
)

func (m *App) loadConversations() tea.Cmd {
	return func() tea.Msg {
		conversations, err := m.db.ListConversations()
		if err != nil {
			return errMsg{err}
		}

		return conversationEntriesMsg(conversations)
	}
}
