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

func complete(m model) tea.Cmd {
	return func() tea.Msg {
		// create a channel to receive chunks of the response
		partialMessageCh := make(chan string)

		// create a channel to receive the final response
		replyCh := make(chan string)

		go m.client.Complete(context.Background(), m.messages, partialMessageCh, replyCh)

		for {
			select {
			case reply := <-replyCh:
				return replyMessage(reply)
			case partial := <-partialMessageCh:
				m.partialMessageCh <- partial
			}
		}
	}
}

type replyMessage string
type partialMessage string

type model struct {
	// layout
	ready bool

	// view models
	viewport viewport.Model
	textarea textarea.Model

	// models
	messages []ai.Message
	partialMessage *ai.Message

	// channels
	partialMessageCh chan string

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
		partialMessageCh: make(chan string),
		client:   ai.NewClient(),
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func receivePartialMessage(m model) tea.Cmd {
	return func() tea.Msg {
		return partialMessage(<-m.partialMessageCh)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case errMsg:
		m.err = msg.err
		return m, tea.Quit

	case partialMessage:
		if m.partialMessage == nil {
			m.partialMessage = &ai.Message{Role: ai.Assistant, Content: string(msg)}
		} else {
			m.partialMessage.Content += string(msg)
		}
		return m, receivePartialMessage(m)

	case replyMessage:
		m.partialMessage = nil
		m.messages = append(m.messages, ai.Message{Role: ai.Assistant, Content: string(msg)})
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
		switch msg.String() {

		case "esc", "ctrl+c":
			return m, tea.Quit

		case "enter":
			v := m.textarea.Value()

			if v == "" {
				// Don't send empty messages.
				return m, nil
			}

			// Clear the textarea
			m.textarea.Reset()

			// Add user message to chat history
			m.messages = append(m.messages, ai.Message{Role: ai.User, Content: v})

			cmds := []tea.Cmd{
				complete(m), // call completions API
				receivePartialMessage(m), // start receiving partial message
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
		messageViews[i] = fmt.Sprintf("%s", msg.Content)
	}

	// Render the messages
	var messages string
	if len(messageViews) > 0 {
		messages = fmt.Sprintf("%s\n", lipgloss.JoinVertical(0, messageViews...))
	}

	// Render the partial message
	if m.partialMessage != nil {
		messages += fmt.Sprintf("%s", m.partialMessage.Content)
	}

	m.viewport.SetContent(messages)
	m.viewport.GotoBottom()

	return docStyle.Render(fmt.Sprintf("%s\n%s",
		borderStyle.Render(m.viewport.View()),
		borderStyle.Render(m.textarea.View()),
	))
}
