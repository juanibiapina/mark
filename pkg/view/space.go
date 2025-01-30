package view

import "github.com/charmbracelet/lipgloss"

// Space is a container that renders a blank space
type Space struct {}

func (s Space) Render(width, height int) string {
	return lipgloss.NewStyle().Width(width).Height(height).Render("")
}
