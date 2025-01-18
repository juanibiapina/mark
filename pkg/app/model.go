package app

import (
	"context"
	"fmt"

	"ant/pkg/ai"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

type replyMessage string
type partialMessage string

var borderStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())

type Model struct {
	// layout
	ready bool

	// view models
	viewport viewport.Model
	textarea textarea.Model

	// models
	conversation     ai.Conversation
	streamingMessage *ai.StreamingMessage

	// channels
	partialMessageCh chan string

	// ai client
	client *ai.Client

	// error
	err error
}

func InitialModel() Model {
	ta := textarea.New()
	ta.Placeholder = "Message Assistant"
	ta.Focus()

	ta.Prompt = ""

	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return Model{
		textarea:         ta,
		partialMessageCh: make(chan string),
		client:           ai.NewClient(),
		conversation:     ai.Conversation{Messages: []ai.Message{}},
	}
}

func complete(m *Model) tea.Cmd {
	return func() tea.Msg {
		// create a channel to receive chunks of the response
		partialMessageCh := make(chan string)

		// create a channel to receive the final response
		replyCh := make(chan string)

		go m.client.Complete(context.Background(), m.conversation.Messages, partialMessageCh, replyCh)

		for {
			val, ok := <-partialMessageCh
			if !ok {
				break
			}
			m.partialMessageCh <- val
		}

		reply := <-replyCh
		return replyMessage(reply)
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func receivePartialMessage(m *Model) tea.Cmd {
	return func() tea.Msg {
		return partialMessage(<-m.partialMessageCh)
	}
}

func (m *Model) newConversation() {
	m.conversation = ai.Conversation{Messages: []ai.Message{}}
	m.streamingMessage = nil
}

func (m *Model) handleMessage() tea.Cmd {
	v := m.textarea.Value()

	// Don't send empty messages.
	if v == "" {
		return nil
	}

	// Clear the textarea
	m.textarea.Reset()

	// Add user message to chat history
	m.conversation.Messages = append(m.conversation.Messages, ai.Message{Role: ai.User, Content: v})

	// Create a new streaming message
	m.streamingMessage = &ai.StreamingMessage{Content: ""}

	cmds := []tea.Cmd{
		complete(m),              // call completions API
		receivePartialMessage(m), // start receiving partial message
	}

	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case errMsg:
		m.err = msg.err
		return m, tea.Quit

	case partialMessage:
		m.streamingMessage.Content += string(msg)
		return m, receivePartialMessage(&m)

	case replyMessage:
		m.streamingMessage = nil
		m.conversation.Messages = append(m.conversation.Messages, ai.Message{Role: ai.Assistant, Content: string(msg)})
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

		case "ctrl+n":
			m.newConversation()
			return m, nil

		case "enter":
			cmd := m.handleMessage()
			return m, cmd

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

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	messageViews := make([]string, len(m.conversation.Messages))
	for i, msg := range m.conversation.Messages {
		messageViews[i] = fmt.Sprintf("%s", msg.Content)
	}

	// Render the messages
	var messages string
	if len(messageViews) > 0 {
		messages = fmt.Sprintf("%s\n", lipgloss.JoinVertical(0, messageViews...))
	}

	// Render the partial message
	if m.streamingMessage != nil {
		messages += fmt.Sprintf("%s", m.streamingMessage.Content)
	}

	m.viewport.SetContent(messages)
	m.viewport.GotoBottom()

	return fmt.Sprintf("%s\n%s",
		borderStyle.Render(m.viewport.View()),
		borderStyle.Render(m.textarea.View()),
	)
}
