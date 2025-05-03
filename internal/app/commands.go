package app

import (
	"slices"

	"mark/internal/llm"
	"mark/internal/util"

	tea "github.com/charmbracelet/bubbletea/v2"
)

func cancelStreaming(agent *llm.Agent) tea.Cmd {
	return func() tea.Msg {
		agent.Cancel()

		return nil
	}
}

func (m *App) saveThread() tea.Cmd {
	return func() tea.Msg {
		err := m.db.SaveThread(m.thread)
		if err != nil {
			return errMsg{err}
		}

		return m.loadThreads()()
	}
}

func (m *App) loadSelectedThread() tea.Cmd {
	if len(m.threadListEntries) == 0 {
		return nil
	}

	selectedEntry := m.threadListEntries[m.threadListCursor]

	return func() tea.Msg {
		thread, err := m.db.LoadThread(selectedEntry.ID)
		if err != nil {
			return errMsg{err}
		}

		return threadMsg{thread}
	}
}

func (m *App) loadThreads() tea.Cmd {
	return func() tea.Msg {
		threads, err := m.db.ListThreads()
		if err != nil {
			return errMsg{err}
		}

		return threadEntriesMsg(threads)
	}
}

func (m *App) deleteSelectedThread() tea.Cmd {
	if len(m.threadListEntries) == 0 {
		return nil
	}

	selectedEntryID := m.threadListEntries[m.threadListCursor].ID

	// Remove the thread from the list of entries
	for i, entry := range m.threadListEntries {
		if entry.ID == selectedEntryID {
			m.threadListEntries = slices.Delete(m.threadListEntries, i, i+1)
			break
		}
	}

	// Ensure the cursor is in a valid position
	if len(m.threadListEntries) == 0 {
		m.threadListCursor = 0
	} else {
		m.threadListCursor = util.Clamp(m.threadListCursor, 0, len(m.threadListEntries)-1)
	}

	m.renderActiveThread()
	m.renderThreadList()

	return func() tea.Msg {
		err := m.db.DeleteThread(selectedEntryID)
		if err != nil {
			return errMsg{err}
		}

		return nil
	}
}

func complete(m *App) tea.Cmd {
	return func() tea.Msg {
		m.agent.Cancel()

		err := m.agent.CompleteStreaming(m.thread)
		if err != nil {
			return errMsg{err}
		}

		return nil
	}
}

func processStream(m *App) tea.Cmd {
	return func() tea.Msg {
		select {
		case v := <-m.agent.Stream.Reply:
			return replyMessage(v)
		case v := <-m.agent.Stream.Chunks:
			return partialMessage(v)
		}
	}
}
