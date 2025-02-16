package app

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	ratio = 0.67
)

func (m *App) renderWindow() string {
	main := NewPane(m.conversationView, m.borderConversation(), "Conversation")

	mainPanelWidth := int(float64(m.width) * ratio)

	return lipgloss.JoinHorizontal(lipgloss.Top, m.renderSidebar(m.width-mainPanelWidth, m.height), main.Render(mainPanelWidth, m.height))
}
