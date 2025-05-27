package app

import tea "github.com/charmbracelet/bubbletea/v2"

type Component interface {
	Focus()
	Blur()
	SetSize(width, height int)
	Update(app *App, msg tea.Msg) tea.Cmd
	View() string
}
