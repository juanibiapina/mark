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
	global       Global
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

	case newConversationMsg:
		m.conversation.Reset()
		m.conversation.Blur()
		m.input.Focus()
		return m, nil

	case cancelStreamingMsg:
		m.conversation.CancelStreaming()
		return m, nil

	case completeMsg:
		m.conversation.CancelStreaming()
		cmd := m.handleMessage()
		return m, cmd

	default:
		var cmd1 tea.Cmd
		var cmd2 tea.Cmd
		var cmd3 tea.Cmd
		m.conversation, cmd1 = m.conversation.Update(msg)
		m.input, cmd2 = m.input.Update(msg)
		m.global, cmd3 = m.global.Update(msg)
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

func receivePartialMessage(c *Conversation) tea.Cmd {
	return func() tea.Msg {
		select {
		case v := <-c.StreamingMessage.Reply:
			return replyMessage(v)
		case v := <-c.StreamingMessage.Chunks:
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

	// Create a new streaming message
	m.conversation.StreamingMessage = NewStreamingMessage()

	// Add user message to chat history
	m.conversation.AddMessage(llm.Message{Role: llm.RoleUser, Content: v})

	cmds := []tea.Cmd{
		complete(m),                            // call completions API
		receivePartialMessage(&m.conversation), // start receiving partial message
	}

	return tea.Batch(cmds...)
}
