package app

import (
	"mark/pkg/util"

	"github.com/charmbracelet/lipgloss"
)

func (m *App) windowView() string {
	return lipgloss.JoinHorizontal(lipgloss.Top, m.sidebarView(), m.mainView())
}

func (m *App) sidebarView() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.inputView(), m.promptListView())
}

func (m *App) mainView() string {
	return util.RenderBorderWithTitle(
		m.conversationView(),
		m.borderConversation(),
		"Conversation",
	)
}

func (m *App) inputView() string {
	return util.RenderBorderWithTitle(
		m.input.View(),
		m.borderInput(),
		"Message Assistant",
	)
}

func (m *App) promptListView() string {
	return util.RenderBorderWithTitle(
		m.promptList.View(),
		m.borderPromptList(),
		"Prompts",
	)
}

func (m *App) conversationView() string {
	return m.conversationViewport.View()
}
