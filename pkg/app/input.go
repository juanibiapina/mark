package app

import (
	"ant/pkg/view"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Input struct {
	view.Focusable

	textarea textarea.Model
}

func MakeInput() Input {
	ta := textarea.New()
	ta.Placeholder = "Message Assistant"
	ta.Focus()

	ta.Cursor.SetMode(cursor.CursorStatic)

	ta.Prompt = ""

	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return Input{
		Focusable: view.MakeFocusable(),
		textarea:  ta,
	}
}

func (i Input) Init() tea.Cmd {
	return nil
}

func (i Input) Update(msg tea.Msg) (Input, tea.Cmd) {
	if !i.Focused() {
		return i, nil
	}

	var cmd tea.Cmd
	i.textarea, cmd = i.textarea.Update(msg)
	return i, cmd
}

func (i Input) View() string {
	return i.BorderStyle().Render(i.textarea.View())
}

func (i *Input) SetWidth(w int) {
	i.textarea.SetWidth(w - i.BorderStyle().GetVerticalFrameSize())
}

func (i *Input) Value() string {
	return i.textarea.Value()
}

func (i *Input) Reset() {
	i.textarea.Reset()
}
