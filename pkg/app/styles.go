package app

import (
	"github.com/charmbracelet/lipgloss"
)

var borderStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
var focusedBorderStyle = borderStyle.BorderForeground(lipgloss.Color("2"))
