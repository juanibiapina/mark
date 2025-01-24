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

type App struct {
	// app state
	uiReady bool

	// models
	conversation     llm.Conversation

	// view models
	conversationView Conversation
	input            Input

	// llm
	ai llm.Llm

	// error
	err error
}

func MakeApp() App {
	return App{
		input: MakeInput(),
		ai:    llm.NewOpenAIClient(),
	}
}

func (m App) Init() tea.Cmd {
	return nil
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if len(os.Getenv("DEBUG")) > 0 {
		log.Print("msg: ", reflect.TypeOf(msg), msg)
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case errMsg:
		m.err = msg.err
		return m, tea.Quit

	case tea.WindowSizeMsg:
		inputHeight := lipgloss.Height(m.input.View())

		if !m.uiReady {
			m.conversationView.Initialize(msg.Width, msg.Height-inputHeight)
			m.uiReady = true
		} else {
			m.conversationView.SetSize(msg.Width, msg.Height-inputHeight)
		}

		m.input.SetWidth(msg.Width)

	case partialMessage:
		// Ignore message if streaming has been cancelled
		if m.conversation.StreamingMessage == nil {
			return m, nil
		}

		m.conversation.StreamingMessage.Content += string(msg)

		cmds = append(cmds, processStream(&m))

	case replyMessage:
		// Ignore message if streaming has been cancelled
		if m.conversation.StreamingMessage == nil {
			return m, nil
		}

		m.conversation.StreamingMessage = nil
		m.conversation.AddMessage(llm.Message{Role: llm.RoleAssistant, Content: string(msg)})

	case tea.KeyMsg:
		switch msg.String() {

		case "q":
			if m.conversationView.Focused() {
				return m, tea.Quit
			}

		case "ctrl+n":
			m.newConversation()

		case "ctrl+c":
			m.cancelStreaming()

		case "enter":
			if m.input.Focused() {
				cmd := m.submitMessage()
				cmds = append(cmds, cmd)
			}

		case "esc":
			if m.input.Focused() {
				m.input.Blur()
				m.conversationView.Focus()
			}

		case "i":
			if m.conversationView.Focused() {
				m.input.Focus()
				m.conversationView.Blur()
			}
		}
	}

	cmd := m.processInputView(msg)
	cmds = append(cmds, cmd)

	m.conversationView.render(&m.conversation, m.conversation.StreamingMessage)
	m.conversationView.ScrollToBottom()

	return m, tea.Batch(cmds...)
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

func (m *App) processInputView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return cmd
}

func (m *App) newConversation() {
	m.cancelStreaming()
	m.conversation.Reset()
	m.conversationView.Blur()
	m.input.Focus()
}

func (m *App) cancelStreaming() {
	if m.conversation.StreamingMessage == nil {
		return
	}

	m.conversation.StreamingMessage.Cancel()

	// Add the partial message to the chat history
	m.conversation.AddMessage(llm.Message{Role: llm.RoleAssistant, Content: m.conversation.StreamingMessage.Content})

	m.conversation.StreamingMessage = nil
}

func (m *App) submitMessage() tea.Cmd {
	v := m.input.Value()
	if v == "" {
		return nil
	}

	m.input.Reset()

	m.cancelStreaming()

	// Create a new streaming message
	m.conversation.StreamingMessage = llm.NewStreamingMessage()

	// Add user message to chat history
	m.conversation.AddMessage(llm.Message{Role: llm.RoleUser, Content: v})

	cmds := []tea.Cmd{
		complete(m),      // call completions API
		processStream(m), // start receiving partial message
	}
	return tea.Batch(cmds...)
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

func processStream(m *App) tea.Cmd {
	return func() tea.Msg {
		select {
		case v := <-m.conversation.StreamingMessage.Reply:
			return replyMessage(v)
		case v := <-m.conversation.StreamingMessage.Chunks:
			return partialMessage(v)
		}
	}
}
