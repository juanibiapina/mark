package util

import (
	"github.com/charmbracelet/lipgloss/v2"
)

func RenderBorderWithTitle(v string, borderStyle lipgloss.Style, title string, titleStyle lipgloss.Style) string {
	// get width of view
	width := lipgloss.Width(v)

	// render pane border
	result := borderStyle.Render(v)

	// place title
	if title != "" && len(title) < width-2 {
		result = PlaceOverlay(2, 0, titleStyle.Render(title), result)
	}

	return result
}
