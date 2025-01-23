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

	// view models
	conversation Conversation
	input        Input

	// llm
	ai llm.Llm

	// error
	err error
}

func MakeApp() App {
	return App{
		input:        MakeInput(),
		conversation: MakeConversation(),
		ai:           llm.NewOpenAIClient(),
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
			m.conversation.Initialize(msg.Width, msg.Height-inputHeight)
			m.uiReady = true
		} else {
			m.conversation.SetSize(msg.Width, msg.Height-inputHeight)
		}

		m.input.SetWidth(msg.Width)

		return m, nil

	case focusInputMsg:
		m.input.Focus()
		m.conversation.Blur()
		return m, nil

	case focusConversationMsg:
		m.input.Blur()
		m.conversation.Focus()
		return m, nil

	case partialMessage:
		// Ignore message if streaming has been cancelled
		if m.conversation.StreamingMessage == nil {
			return m, nil
		}

		m.conversation.StreamingMessage.Content += string(msg)
		m.conversation.render()
		m.conversation.ScrollToBottom()

		return m, receivePartialMessage(&m)

	case replyMessage:
		// Ignore message if streaming has been cancelled
		if m.conversation.StreamingMessage == nil {
			return m, nil
		}

		m.conversation.StreamingMessage = nil
		m.conversation.messages = append(m.conversation.messages, llm.Message{Role: llm.RoleAssistant, Content: string(msg)})
		m.conversation.render()
		m.conversation.ScrollToBottom()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+n":
			m.newConversation()
			return m, nil

		case "ctrl+c":
			m.cancelStreaming()
			return m, nil

		case "enter":
			cmd := m.submitMessage()
			return m, cmd

		default:
			var cmd1 tea.Cmd
			var cmd2 tea.Cmd
			var cmd3 tea.Cmd
			m.conversation, cmd1 = m.conversation.Update(msg)
			m.input, cmd2 = m.input.Update(msg)
			return m, tea.Batch(cmd1, cmd2, cmd3)
		}

	default:
		var cmd1 tea.Cmd
		var cmd2 tea.Cmd
		var cmd3 tea.Cmd
		m.conversation, cmd1 = m.conversation.Update(msg)
		m.input, cmd2 = m.input.Update(msg)
		return m, tea.Batch(cmd1, cmd2, cmd3)

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

func (m *App) newConversation() {
	m.conversation.Reset()
	m.conversation.Blur()
	m.input.Focus()
}

func (m *App) cancelStreaming() {
	m.conversation.CancelStreaming()
}

func (m *App) submitMessage() tea.Cmd {
	if !m.input.Focused() {
		return nil
	}

	v := m.input.Value()
	if v == "" {
		return nil
	}

	m.input.Reset()

	m.cancelStreaming()

	// Create a new streaming message
	m.conversation.StreamingMessage = NewStreamingMessage()

	// Add user message to chat history
	m.conversation.AddMessage(llm.Message{Role: llm.RoleUser, Content: v})

	cmds := []tea.Cmd{
		complete(m),              // call completions API
		receivePartialMessage(m), // start receiving partial message
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
