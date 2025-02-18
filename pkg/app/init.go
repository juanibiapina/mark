package app

import (
	"github.com/charmbracelet/bubbles/v2/cursor"
	"github.com/charmbracelet/bubbles/v2/textarea"
	"github.com/charmbracelet/lipgloss/v2"
)

func (m *App) initInput() {
	m.input = textarea.New()
	m.input.Focus() // focus is actually handled by the app

	m.input.Cursor.SetMode(cursor.CursorStatic)

	m.input.Prompt = ""

	// Remove cursor line styling
	m.input.Styles.Focused.CursorLine = lipgloss.NewStyle()

	m.input.ShowLineNumbers = false

	m.input.KeyMap.InsertNewline.SetEnabled(false)
}
