package app

import (
	tea "github.com/charmbracelet/bubbletea/v2"
)

func cancelStreaming(agent *Agent) tea.Cmd {
	return func() tea.Msg {
		agent.Cancel()

		return nil
	}
}

func complete(m *App) tea.Cmd {
	return func() tea.Msg {
		err := m.agent.CompleteStreaming(m.session)
		if err != nil {
			return errMsg{err}
		}

		return nil
	}
}

func processEvents(events chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		v := <-events
		return eventMsg{v}
	}
}
