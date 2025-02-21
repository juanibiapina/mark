package app

import (
	"mark/pkg/model"

	tea "github.com/charmbracelet/bubbletea/v2"
)

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

func (m *App) submitMessage() tea.Cmd {
	m.cancelStreaming()

	v := m.input.Value()
	if v != "" {
		m.thread.AddMessage(model.Message{Role: model.RoleUser, Content: v})
		m.input.Reset()
	}

	// maybe update the prompt here

	m.startStreaming()

	cmds := []tea.Cmd{
		complete(m),      // call completions API
		processStream(m), // start receiving partial message
	}
	return tea.Batch(cmds...)
}

func (m *App) deleteSelectedThread() tea.Cmd {
	if len(m.threadListEntries) == 0 {
		return nil
	}

	selectedEntryID := m.threadListEntries[m.threadListCursor].ID

	return func() tea.Msg {
		err := m.db.DeleteThread(selectedEntryID)
		if err != nil {
			return errMsg{err}
		}

		return removeThreadMsg(selectedEntryID)
	}
}

func complete(m *App) tea.Cmd {
	return func() tea.Msg {
		err := m.ai.CompleteStreaming(&m.thread, m.stream)
		if err != nil {
			return errMsg{err}
		}

		return nil
	}
}

func processStream(m *App) tea.Cmd {
	return func() tea.Msg {
		select {
		case v := <-m.stream.Reply:
			return replyMessage(v)
		case v := <-m.stream.Chunks:
			return partialMessage(v)
		}
	}
}
