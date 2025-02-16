package app

import (
	"mark/pkg/view"

	"github.com/charmbracelet/lipgloss"
)

const (
	ratio = 0.67
)

func (m *App) renderWindow() string {
	sidebar := view.Sidebar{
		Input:   view.NewPane(m.input, m.borderInput(), "Message Assistant"),
		Prompts: view.NewPane(m.promptListView, m.borderPromptList(), "Prompts"),
	}
	main := view.NewPane(m.conversationView, m.borderConversation(), "Conversation")

	mainPanelWidth := int(float64(m.width) * ratio)

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar.Render(m.width-mainPanelWidth, m.height), main.Render(mainPanelWidth, m.height))
}
