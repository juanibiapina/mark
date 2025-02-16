package app

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	ratio = 0.67
)

func (m *App) renderWindow(width, height int) string {
	main := NewPane(m.conversationView, m.borderConversation(), "Conversation")

	mainPanelWidth := int(float64(width) * ratio)

	return lipgloss.JoinHorizontal(lipgloss.Top, m.renderSidebar(width-mainPanelWidth, height), main.Render(*m, mainPanelWidth, height))
}
