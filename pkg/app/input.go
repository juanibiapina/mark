package app

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type input struct {
	textarea textarea.Model
}

func MakeInput() input {
	ta := textarea.New()
	ta.Placeholder = "Message Assistant"
	ta.Focus()

	ta.Prompt = ""

	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return input{
		textarea: ta,
	}
}

func (i input) Init() tea.Cmd {
	return textarea.Blink
}

func (i input) Update(msg tea.Msg) (input, tea.Cmd) {
	var cmd tea.Cmd
	i.textarea, cmd = i.textarea.Update(msg)
	return i, cmd
}

func (i input) View() string {
	return borderStyle.Render(i.textarea.View())
}

func (i *input) Value() string {
	return i.textarea.Value()
}

func (i *input) Reset() {
	i.textarea.Reset()
}

func (i *input) SetWidth(w int) {
	i.textarea.SetWidth(w - borderStyle.GetVerticalFrameSize())
}
