package app

import (
	"mark/pkg/model"

	tea "github.com/charmbracelet/bubbletea/v2"
)

func (m *App) saveConversation() tea.Cmd {
	return func() tea.Msg {
		err := m.db.SaveConversation(m.conversation)
		if err != nil {
			return errMsg{err}
		}

		return m.loadConversations()()
	}
}

func (m *App) loadConversations() tea.Cmd {
	return func() tea.Msg {
		conversations, err := m.db.ListConversations()
		if err != nil {
			return errMsg{err}
		}

		return conversationEntriesMsg(conversations)
	}
}

func (m *App) submitMessage() tea.Cmd {
	m.cancelStreaming()

	v := m.input.Value()
	if v != "" {
		m.conversation.AddMessage(model.Message{Role: model.RoleUser, Content: v})
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

func complete(m *App) tea.Cmd {
	return func() tea.Msg {
		err := m.ai.CompleteStreaming(&m.conversation, m.stream)
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
