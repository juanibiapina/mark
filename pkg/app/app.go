package app

import (
	"fmt"

	"ant/pkg/ai"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

type replyMessage string
type partialMessage string

type App struct {
	// layout
	ready bool

	// view models
	viewport viewport.Model
	input    input

	// models
	conversation     ai.Conversation
	streamingMessage *ai.StreamingMessage

	// ai client
	client *ai.Client

	// error
	err error
}

func MakeApp() App {
	return App{
		input:        MakeInput(),
		client:       ai.NewClient(),
		conversation: ai.Conversation{Messages: []ai.Message{}},
	}
}

func (m App) Init() tea.Cmd {
	return tea.Batch(
		m.input.Init(),
	)
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case errMsg:
		m.err = msg.err
		return m, tea.Quit

	case partialMessage:
		// Ignore messages if streaming has been cancelled
		if m.streamingMessage == nil {
			return m, nil
		}

		m.streamingMessage.Content += string(msg)
		return m, receivePartialMessage(&m)

	case replyMessage:
		// Ignore messages if streaming has been cancelled
		if m.streamingMessage == nil {
			return m, nil
		}

		m.streamingMessage = nil
		m.conversation.Messages = append(m.conversation.Messages, ai.Message{Role: ai.Assistant, Content: string(msg)})
		return m, nil

	case tea.WindowSizeMsg:
		inputHeight := lipgloss.Height(m.input.View())

		if !m.ready {
			m.viewport = viewport.New(msg.Width-borderStyle.GetVerticalFrameSize(), msg.Height-borderStyle.GetHorizontalFrameSize()-inputHeight)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - borderStyle.GetVerticalFrameSize()
			m.viewport.Height = msg.Height - borderStyle.GetHorizontalFrameSize() - inputHeight
		}

		m.input.SetWidth(msg.Width)

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {

		case "esc":
			return m, tea.Quit

		case "ctrl+c":
			m.cancelStreaming()
			return m, nil

		case "ctrl+n":
			m.newConversation()
			return m, nil

		case "enter":
			cmd := m.handleMessage()
			return m, cmd

		default:
			// Send all other keypresses to the input component
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

	case cursor.BlinkMsg:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd

	default:
		return m, nil
	}
}

func (m App) View() string {
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
		m.input.View(),
	)
}

func (m *App) cancelStreaming() {
	if m.streamingMessage == nil {
		return
	}

	m.streamingMessage.Cancel()

	// Add the partial message to the chat history
	m.conversation.Messages = append(m.conversation.Messages, ai.Message{Role: ai.Assistant, Content: m.streamingMessage.Content})

	m.streamingMessage = nil
}

func complete(m *App) tea.Cmd {
	return func() tea.Msg {
		m.client.Complete(m.streamingMessage.Ctx, m.conversation.Messages, m.streamingMessage.Chunks, m.streamingMessage.Reply)

		return nil
	}
}

func receivePartialMessage(m *App) tea.Cmd {
	return func() tea.Msg {
		select {
		case v := <-m.streamingMessage.Reply:
			return replyMessage(v)
		case v := <-m.streamingMessage.Chunks:
			return partialMessage(v)
		}
	}
}

func (m *App) newConversation() {
	m.cancelStreaming()
	m.conversation = ai.Conversation{Messages: []ai.Message{}}
}

func (m *App) handleMessage() tea.Cmd {
	v := m.input.Value()

	// Don't send empty messages.
	if v == "" {
		return nil
	}

	// Clear the input
	m.input.Reset()

	// Add user message to chat history
	m.conversation.Messages = append(m.conversation.Messages, ai.Message{Role: ai.User, Content: v})

	// Create a new streaming message
	m.streamingMessage = ai.NewStreamingMessage()

	cmds := []tea.Cmd{
		complete(m),              // call completions API
		receivePartialMessage(m), // start receiving partial message
	}

	return tea.Batch(cmds...)
}
