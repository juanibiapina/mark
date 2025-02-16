package app

import (
	"mark/pkg/util"

	"github.com/charmbracelet/lipgloss"
)

// Pane is a container with a border
type Pane struct {
	c           Container
	borderStyle lipgloss.Style
	title       string
}

func NewPane(c Container, borderStyle lipgloss.Style, title string) Pane {
	return Pane{
		c:           c,
		borderStyle: borderStyle,
		title:       title,
	}
}

func (p Pane) Render(m App, width, height int) string {
	body := p.c.Render(m, width-p.borderStyle.GetVerticalFrameSize(), height-p.borderStyle.GetHorizontalFrameSize())
	r := p.borderStyle.Render(body)
	if p.title != "" && len(p.title) < width-2 {
		r = util.PlaceOverlay(2, 0, p.title, r)
	}
	return r
}
