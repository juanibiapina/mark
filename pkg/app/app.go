package app

import (
	"ant/pkg/llm"
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
	conversation Conversation
	input        Input

	// LLM
	ai llm.Llm

	// error
	err error
}

func MakeApp() App {
	return App{
		input:        MakeInput(),
		conversation: MakeConversation(),

		ai: llm.NewOpenAIClient(),
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
		if m.conversation.StreamingMessage == nil {
			return m, nil
		}

		m.conversation.StreamingMessage.Content += string(msg)
		m.conversation.RenderMessagesTop()
		m.conversation.ScrollToBottom()
		return m, receivePartialMessage(&m)

	case replyMessage:
		// Ignore messages if streaming has been cancelled
		if m.conversation.StreamingMessage == nil {
			return m, nil
		}

		m.conversation.StreamingMessage = nil
		m.conversation.messages = append(m.conversation.messages, llm.Message{Role: llm.RoleAssistant, Content: string(msg)})
		m.conversation.RenderMessagesTop()
		m.conversation.ScrollToBottom()
		return m, nil

	case tea.WindowSizeMsg:
		inputHeight := lipgloss.Height(m.input.View())

		if !m.uiReady {
			m.conversation.Initialize(msg.Width, msg.Height-inputHeight)
			m.uiReady = true
		} else {
			m.conversation.SetSize(msg.Width, msg.Height-inputHeight)
		}

		m.input.SetWidth(msg.Width)

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {

		case "esc":
			return m, tea.Quit

		case "ctrl+c":
			m.conversation.CancelStreaming()
			return m, nil

		case "ctrl+n":
			m.conversation.CancelStreaming()
			m.conversation.ResetMessages()
			m.conversation.RenderMessagesTop()
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
		m.conversation.View(),
		m.input.View(),
	)
}

func complete(m *App) tea.Cmd {
	return func() tea.Msg {
		m.ai.CompleteStreaming(
			m.conversation.StreamingMessage.Ctx,
			&m.conversation,
			m.conversation.StreamingMessage.Chunks,
			m.conversation.StreamingMessage.Reply,
		)

		return nil
	}
}

func receivePartialMessage(m *App) tea.Cmd {
	return func() tea.Msg {
		select {
		case v := <-m.conversation.StreamingMessage.Reply:
			return replyMessage(v)
		case v := <-m.conversation.StreamingMessage.Chunks:
			return partialMessage(v)
		}
	}
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
	m.conversation.messages = append(m.conversation.messages, llm.Message{Role: llm.RoleUser, Content: v})

	// Create a new streaming message
	m.conversation.StreamingMessage = NewStreamingMessage()

	// Render conversation view again to show the new message
	m.conversation.RenderMessagesTop()

	cmds := []tea.Cmd{
		complete(m),              // call completions API
		receivePartialMessage(m), // start receiving partial message
	}

	return tea.Batch(cmds...)
}
