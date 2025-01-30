package view

import "github.com/charmbracelet/lipgloss"

type Pane struct {
	content     string
	borderStyle lipgloss.Style
}

func NewPane(content string, borderStyle lipgloss.Style) Pane {
	return Pane{
		content:     content,
		borderStyle: borderStyle,
	}
}

func (p Pane) Render() string {
	return p.borderStyle.Render(p.content)
}
