package util

import (
	"github.com/charmbracelet/lipgloss"
)

func RenderBorderWithTitle(v string, borderStyle lipgloss.Style, title string) string {
	width := lipgloss.Width(v)

	r := borderStyle.Render(v)
	if title != "" && len(title) < width-2 {
		r = PlaceOverlay(2, 0, title, r)
	}
	return r
}
