package app

import (
	"log/slog"

	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type InputDialog struct {
	width    int
	height   int
	input    textinput.Model
	callback func(v string) error
}

func NewInputDialog(callback func(v string) error) *InputDialog {
	input := textinput.New()
	input.Focus()

	return &InputDialog{
		input: input,
		callback: callback,
	}
}

func (dialog *InputDialog) SetSize(width, height int) {
	dialog.width = width
	dialog.height = height
	dialog.input.SetWidth(width)
}

func (dialog *InputDialog) Update(app *App, msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			err := dialog.callback(dialog.input.Value())
			if err != nil {
				slog.Error("Error") // TODO handle error
				return nil
			}
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

func (dialog *InputDialog) View() string {
	return lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Render(dialog.input.View())
}
