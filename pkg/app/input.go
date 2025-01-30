package app

import (
	"github.com/charmbracelet/bubbles/v2/cursor"
	"github.com/charmbracelet/bubbles/v2/textarea"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type Input struct {
	textarea textarea.Model
}

func MakeInput() Input {
	ta := textarea.New()
	ta.Placeholder = "Message Assistant"
	ta.Focus() // focus is actually handled by the app

	ta.Cursor.SetMode(cursor.CursorStatic)

	ta.Prompt = ""

	ta.SetHeight(3)

	// Remove cursor line styling
	ta.Styles.Focused.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return Input{
		textarea: ta,
	}
}

func (i Input) Init() (Input, tea.Cmd) {
	return i, nil
}

func (i Input) Update(msg tea.Msg) (Input, tea.Cmd) {
	var cmd tea.Cmd
	i.textarea, cmd = i.textarea.Update(msg)
	return i, cmd
}

func (i Input) View() string {
	return i.textarea.View()
}

func (i Input) Render(width, height int) string {
	i.textarea.SetWidth(width)
	i.textarea.SetHeight(height)
	return i.textarea.View()
}

func (i *Input) SetWidth(w int) {
	i.textarea.SetWidth(w)
}

func (i *Input) Value() string {
	return i.textarea.Value()
}

func (i *Input) Reset() {
	i.textarea.Reset()
}

func (i *Input) Width() int {
	return i.textarea.Width()
}
