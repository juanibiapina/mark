package main

import (
	"context"
	"fmt"
	"os"

	"ant/pkg/ai"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
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

type Message struct {
	text   string
}

type model struct {
	// layout
	ready bool

	// view models
	viewport viewport.Model
	textarea textarea.Model

	// models
	messages []Message

	// ai client
	client *ai.Client

	// error
	err error
}

var docStyle = lipgloss.NewStyle()
var borderStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())

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

	return model{
		textarea: ta,
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
		m.messages = append(m.messages, Message{text: string(msg)})
		//renderer, err := glamour.NewTermRenderer(glamour.WithAutoStyle())
		//if err != nil {
		//	m.err = err
		//	return m, tea.Quit
		//}

		//str, err := renderer.Render(string(msg))
		//if err != nil {
		//	m.err = err
		//	return m, tea.Quit
		//}

		//m.viewport.SetContent(str)
		return m, nil

	case tea.WindowSizeMsg:
		textAreaHeight := lipgloss.Height(borderStyle.Render(m.textarea.View()))

		if !m.ready {
			m.viewport = viewport.New(msg.Width-borderStyle.GetVerticalFrameSize(), msg.Height-borderStyle.GetHorizontalFrameSize()-textAreaHeight)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - borderStyle.GetVerticalFrameSize()
			m.viewport.Height = msg.Height - borderStyle.GetHorizontalFrameSize() - textAreaHeight
		}

		m.textarea.SetWidth(msg.Width - borderStyle.GetVerticalFrameSize())

		return m, nil

	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyEsc, tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEnter:
			v := m.textarea.Value()

			if v == "" {
				// Don't send empty messages.
				return m, nil
			}

			m.textarea.Reset()
			m.messages = append(m.messages, Message{text: v})

			cmds := []tea.Cmd{
				userMessage(m, v), // send user message to AI
			}

			return m, tea.Batch(cmds...)

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
	if !m.ready {
		return "Initializing..."
	}

	messageViews := make([]string, len(m.messages))
	for i, msg := range m.messages {
		messageViews[i] = fmt.Sprintf("%s", msg.text)
	}

	// Render the messages
	var messages string
	if len(messageViews) > 0 {
		messages = fmt.Sprintf("%s\n", lipgloss.JoinVertical(0, messageViews...))
	}

	m.viewport.SetContent(messages)
	m.viewport.GotoBottom()

	return docStyle.Render(fmt.Sprintf("%s\n%s",
		borderStyle.Render(m.viewport.View()),
		borderStyle.Render(m.textarea.View()),
	))
}
