package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"ant/pkg/ai"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
	}
}

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func userMessage(m model, text string) tea.Cmd {
	return func() tea.Msg {
		reply, err := m.client.SendMessage(context.Background(), text)
		if err != nil {
			return errMsg{err}
		}

		return replyMessage(reply)
	}
}

type replyMessage string

type ChatMessage struct {
	title, desc string
}

func (i ChatMessage) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(ChatMessage)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s: %s", i.title, i.desc)

	fmt.Fprint(w, str)
}

type model struct {
	// view models
	list     list.Model
	textarea textarea.Model

	// ai client
	client *ai.Client

	// error
	err error
}

var docStyle = lipgloss.NewStyle()

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Message Assistant"
	ta.Focus()

	ta.Prompt = "┃ "

	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	ta.KeyMap.InsertNewline.SetEnabled(false)

	l := list.New([]list.Item{}, itemDelegate{}, 0, 0)
	l.Title = "Messages"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)

	return model{
		textarea: ta,
		list:     l,
		client:   ai.NewClient(),
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case errMsg:
		m.err = msg.err
		return m, tea.Quit

	case replyMessage:
		m.list.InsertItem(len(m.list.Items()), ChatMessage{title: "AI", desc: string(msg)})

		return m, nil

	case tea.WindowSizeMsg:
		textAreaHeight := m.textarea.Height()

		m.list.SetSize(msg.Width, msg.Height-textAreaHeight)

		m.textarea.SetWidth(msg.Width)

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {

		case "esc", "ctrl+c":
			return m, tea.Quit

		case "enter":
			v := m.textarea.Value()

			if v == "" {
				// Don't send empty messages.
				return m, nil
			}

			m.textarea.Reset()

			return m, tea.Batch(m.list.InsertItem(len(m.list.Items()), ChatMessage{title: "You", desc: v}), userMessage(m, v))

		default:
			// Send all other keypresses to the textarea.
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
		}

	case cursor.BlinkMsg:
		// Textarea should also process cursor blinks.
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd

	default:
		return m, nil
	}
}

func (m model) View() string {
	return docStyle.Render(fmt.Sprintf(
		"%s\n%s",
		m.list.View(),
		m.textarea.View(),
	))
}
