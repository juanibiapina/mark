package app

import (
	"fmt"
	"log"
	"os"
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

type replyMessage string
type partialMessage string

type App struct {
	// initialization
	uiReady bool

	// view models
	conversationView Conversation
	input            Input

	// clients
	ai *AIClient

	// error
	err error
}

func MakeApp() App {
	return App{
		input:            MakeInput(),
		conversationView: MakeConversation(),

		ai:           NewAIClient(),
	}
}

func (m App) Init() tea.Cmd {
	return nil
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if len(os.Getenv("DEBUG")) > 0 {
		log.Print("msg: ", reflect.TypeOf(msg), msg)
	}

	switch msg := msg.(type) {

	case errMsg:
		m.err = msg.err
		return m, tea.Quit

	case partialMessage:
		// Ignore messages if streaming has been cancelled
		if m.conversationView.StreamingMessage == nil {
			return m, nil
		}

		m.conversationView.StreamingMessage.Content += string(msg)
		m.updateConversationView()
		m.conversationView.ScrollToBottom()
		return m, receivePartialMessage(&m)

	case replyMessage:
		// Ignore messages if streaming has been cancelled
		if m.conversationView.StreamingMessage == nil {
			return m, nil
		}

		m.conversationView.StreamingMessage = nil
		m.conversationView.Messages = append(m.conversationView.Messages, Message{Role: RoleAssistant, Content: string(msg)})
		m.updateConversationView()
		m.conversationView.ScrollToBottom()
		return m, nil

	case tea.WindowSizeMsg:
		inputHeight := lipgloss.Height(m.input.View())

		if !m.uiReady {
			m.conversationView.Initialize(msg.Width, msg.Height-inputHeight)
			m.uiReady = true
		} else {
			m.conversationView.SetSize(msg.Width, msg.Height-inputHeight)
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
			m.updateConversationView()
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

	default:
		return m, nil
	}
}

func (m App) View() string {
	if !m.uiReady {
		return "Initializing UI..."
	}

	return fmt.Sprintf("%s\n%s",
		m.conversationView.View(),
		m.input.View(),
	)
}

func (m *App) updateConversationView() {
	m.conversationView.RenderMessagesTop()
}

func (m *App) cancelStreaming() {
	if m.conversationView.StreamingMessage == nil {
		return
	}

	m.conversationView.StreamingMessage.Cancel()

	// Add the partial message to the chat history
	m.conversationView.Messages = append(m.conversationView.Messages, Message{Role: RoleAssistant, Content: m.conversationView.StreamingMessage.Content})

	m.conversationView.StreamingMessage = nil
}

func complete(m *App) tea.Cmd {
	return func() tea.Msg {
		m.ai.Complete(m.conversationView.StreamingMessage.Ctx, m.conversationView.Messages, m.conversationView.StreamingMessage.Chunks, m.conversationView.StreamingMessage.Reply)

		return nil
	}
}

func receivePartialMessage(m *App) tea.Cmd {
	return func() tea.Msg {
		select {
		case v := <-m.conversationView.StreamingMessage.Reply:
			return replyMessage(v)
		case v := <-m.conversationView.StreamingMessage.Chunks:
			return partialMessage(v)
		}
	}
}

func (m *App) newConversation() {
	m.cancelStreaming()
	m.conversationView = Conversation{Messages: []Message{}}
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
	m.conversationView.Messages = append(m.conversationView.Messages, Message{Role: RoleUser, Content: v})

	// Create a new streaming message
	m.conversationView.StreamingMessage = NewStreamingMessage()

	// Render conversation view again to show the new message
	m.updateConversationView()

	cmds := []tea.Cmd{
		complete(m),              // call completions API
		receivePartialMessage(m), // start receiving partial message
	}

	return tea.Batch(cmds...)
}
