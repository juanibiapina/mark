package view

import "github.com/charmbracelet/lipgloss"

// Main is a container that splits the screen into two parts with a specific ratio
type Main struct {
	Left  Container
	Right Container
	Ratio float64
}

func (h Main) Render(width, height int) string {
	mainPanelWidth := int(float64(width) * h.Ratio)

	return lipgloss.JoinHorizontal(lipgloss.Top, h.Left.Render(width-mainPanelWidth, height), h.Right.Render(mainPanelWidth, height))
}
