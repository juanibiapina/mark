package app

import (
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type AddContextDialog struct {
	width  int
	height int
	input  textinput.Model
}

func NewDialog() *AddContextDialog {
	input := textinput.New()
	input.Focus()

	return &AddContextDialog{
		input: input,
	}
}

func (dialog *AddContextDialog) SetSize(width, height int) {
	dialog.width = width
	dialog.height = height
	dialog.input.SetWidth(width)
}

func (dialog *AddContextDialog) Update(app *App, msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			app.addContext(dialog.input.Value())
			app.hideAddContextDialog()
		case "esc":
			app.hideAddContextDialog()
		default:
			var cmd tea.Cmd
			dialog.input, _ = dialog.input.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return tea.Batch(cmds...)
}

func (dialog *AddContextDialog) View() string {
	return lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Render(dialog.input.View())
}
