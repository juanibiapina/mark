package view

import "github.com/charmbracelet/lipgloss"

var (
	borderStyle        = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	focusedBorderStyle = borderStyle.BorderForeground(lipgloss.Color("2"))
)

type Focusable struct {
	focus       bool
	borderStyle lipgloss.Style
}

func MakeFocusable() Focusable {
	return Focusable{
		focus: true,
	}
}

func (f *Focusable) Focus() {
	f.focus = true
}

func (f *Focusable) Blur() {
	f.focus = false
}

func (f *Focusable) Focused() bool {
	return f.focus
}

func (f *Focusable) BorderStyle() lipgloss.Style {
	if f.focus {
		return focusedBorderStyle
	}
	return borderStyle
}
