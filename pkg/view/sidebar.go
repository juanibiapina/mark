package view

import "github.com/charmbracelet/lipgloss"

type Sidebar struct {
	Input Container
	Empty Container
}

func (s Sidebar) Render(width, height int) string {
	inputHeight := 5
	emptyHeight := height - inputHeight

	return lipgloss.JoinVertical(lipgloss.Left, s.Input.Render(width, inputHeight), s.Empty.Render(width, emptyHeight))
}
