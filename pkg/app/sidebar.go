package app

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	inputHeight = 5
)

func (m *App) renderSidebar(width, height int) string {
	promptListHeight := height - inputHeight

	input := NewPane(m.input, m.borderInput(), "Message Assistant")
	prompts := NewPane(m.promptListView, m.borderPromptList(), "Prompts")

	return lipgloss.JoinVertical(lipgloss.Left, input.Render(width, inputHeight), prompts.Render(width, promptListHeight))
}
