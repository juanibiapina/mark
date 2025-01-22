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

type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
)

type App struct {
	// app state
	uiReady bool
	mode    Mode

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
		mode: ModeInsert,

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

	case partialMessage, replyMessage:
		var cmd tea.Cmd
		m.conversation, cmd = m.conversation.Update(msg)
		return m, cmd

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
		// global keybindings
		switch msg.String() {
			case "ctrl+n":
				m.conversation.Reset()
				return m, nil

			case "ctrl+c":
				m.conversation.CancelStreaming()
				return m, nil
		}

		// normal mode keybindings
		if m.mode == ModeNormal {
			switch msg.String() {

			case "i":
				m.mode = ModeInsert
				return m, nil

			case "q":
				return m, tea.Quit

			default:
				return m, nil
			}
		}

		// insert mode keybindings
		if m.mode == ModeInsert {
			switch msg.String() {

			case "esc":
				m.mode = ModeNormal
				return m, nil

			case "enter":
				m.conversation.CancelStreaming()
				cmd := m.handleMessage()
				return m, cmd

			default:
				// Send keypresses to the input component
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}
		}

		return m, nil

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
