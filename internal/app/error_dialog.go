package app

import (
	"log/slog"

	"mark/internal/util"

	"github.com/charmbracelet/bubbles/v2/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type ErrorDialog struct {
	width    int
	height   int
	hasFocus bool
	viewport viewport.Model
}

func NewErrorDialog(err error) *ErrorDialog {
	viewport := viewport.New()
	viewport.SetContent(err.Error())

	return &ErrorDialog{
		viewport: viewport,
	}
}

func (dialog *ErrorDialog) Focus() {
	dialog.hasFocus = true
}

func (dialog *ErrorDialog) Blur() {
	dialog.hasFocus = false
}

func (dialog *ErrorDialog) SetSize(width, height int) {
	dialog.width = width
	dialog.height = height
	dialog.viewport.SetWidth(width - 2)   // Subtract 2 for borders
	dialog.viewport.SetHeight(height - 2) // Subtract 2 for borders
}

func (dialog *ErrorDialog) Update(app *App, msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		default:
			app.hideDialog()
		}
	}

	return tea.Batch(cmds...)
}

func (dialog *ErrorDialog) View() string {
	slog.Info("Rendering error dialog", "width", dialog.width, "height", dialog.height)
	return util.RenderBorderWithTitle(
		dialog.viewport.View(),
		dialog.BorderStyle(),
		"Error",
		dialog.TitleStyle(),
	)
}

func (dialog *ErrorDialog) BorderStyle() lipgloss.Style {
	if dialog.hasFocus {
		return focusedBorderStyle
	}
	return borderStyle
}

func (dialog *ErrorDialog) TitleStyle() lipgloss.Style {
	if dialog.hasFocus {
		return focusedPanelTitleStyle
	}
	return textStyle
}
