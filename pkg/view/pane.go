package view

import "github.com/charmbracelet/lipgloss"

// Pane is a container with a border
type Pane struct {
	c           Container
	borderStyle lipgloss.Style
}

func NewPane(c Container, borderStyle lipgloss.Style) Pane {
	return Pane{
		c:           c,
		borderStyle: borderStyle,
	}
}

func (p Pane) Render(width, height int) string {
	return p.borderStyle.Render(p.c.Render(width-p.borderStyle.GetVerticalFrameSize(), height-p.borderStyle.GetHorizontalFrameSize()))
}
