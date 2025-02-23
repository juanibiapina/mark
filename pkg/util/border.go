package util

import (
	"image/color"

	"github.com/charmbracelet/lipgloss/v2"
)

func RenderBorderWithTitle(v string, borderStyle lipgloss.Style, title string, titleColor color.Color) string {
	width := lipgloss.Width(v)

	// render border
	result := borderStyle.Render(v)

	// place title
	if title != "" && len(title) < width-2 {
		style := lipgloss.NewStyle().Foreground(titleColor)
		result = PlaceOverlay(2, 0, style.Render(title), result)
	}

	return result
}
