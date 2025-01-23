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
	StreamingMessage *llm.StreamingMessage

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

		return m, nil

	case partialMessage:
		// Ignore message if streaming has been cancelled
		if m.StreamingMessage == nil {
			return m, nil
		}

		m.StreamingMessage.Content += string(msg)
		m.conversationView.render(&m.conversation, m.StreamingMessage)
		m.conversationView.ScrollToBottom()

		return m, processStream(&m)

	case replyMessage:
		// Ignore message if streaming has been cancelled
		if m.StreamingMessage == nil {
			return m, nil
		}

		m.StreamingMessage = nil
		m.conversation.AddMessage(llm.Message{Role: llm.RoleAssistant, Content: string(msg)})
		m.conversationView.render(&m.conversation, m.StreamingMessage)
		m.conversationView.ScrollToBottom()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {

		case "q":
			if m.conversationView.Focused() {
				return m, tea.Quit
			}

		case "ctrl+n":
			m.newConversation()
			return m, nil

		case "ctrl+c":
			m.cancelStreaming()
			return m, nil

		case "enter":
			if m.input.Focused() {
				cmd := m.submitMessage()
				return m, cmd
			}

		case "esc":
			if m.input.Focused() {
				m.input.Blur()
				m.conversationView.Focus()
				return m, nil
			}

		case "i":
			if m.conversationView.Focused() {
				m.input.Focus()
				m.conversationView.Blur()
				return m, nil
			}
		}
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.conversationView, cmd = m.conversationView.Update(msg)
	cmds = append(cmds, cmd)
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

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

func (m *App) newConversation() {
	m.cancelStreaming()
	m.conversation.Reset()
	m.conversationView.render(&m.conversation, m.StreamingMessage)
	m.conversationView.Blur()
	m.input.Focus()
}

func (m *App) cancelStreaming() {
	if m.StreamingMessage == nil {
		return
	}

	m.StreamingMessage.Cancel()

	// Add the partial message to the chat history
	m.conversation.AddMessage(llm.Message{Role: llm.RoleAssistant, Content: m.StreamingMessage.Content})

	m.StreamingMessage = nil
}

func (m *App) submitMessage() tea.Cmd {
	v := m.input.Value()
	if v == "" {
		return nil
	}

	m.input.Reset()

	m.cancelStreaming()

	// Create a new streaming message
	m.StreamingMessage = llm.NewStreamingMessage()

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
			m.StreamingMessage.Ctx,
			&m.conversation,
			m.StreamingMessage.Chunks,
			m.StreamingMessage.Reply,
		)

		return nil
	}
}

func processStream(m *App) tea.Cmd {
	return func() tea.Msg {
		select {
		case v := <-m.StreamingMessage.Reply:
			return replyMessage(v)
		case v := <-m.StreamingMessage.Chunks:
			return partialMessage(v)
		}
	}
}
