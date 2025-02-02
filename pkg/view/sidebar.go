package view

import "github.com/charmbracelet/lipgloss"

type Sidebar struct {
	Input Container
	Prompts Container
}

func (s Sidebar) Render(width, height int) string {
	inputHeight := 5
	promptListHeight := height - inputHeight

	return lipgloss.JoinVertical(lipgloss.Left, s.Input.Render(width, inputHeight), s.Prompts.Render(width, promptListHeight))
}
