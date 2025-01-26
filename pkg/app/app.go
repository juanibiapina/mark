package app

import (
	"fmt"
	"log"
	"os"
	"reflect"

	"ant/pkg/llm"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type App struct {
	// app state
	uiReady bool

	// models
	conversation llm.Conversation
	project      *Project

	// view models
	conversationView Conversation
	input            Input

	// llm
	ai llm.Llm

	// error
	err error
}

func MakeApp() App {
	app := App{
		input: MakeInput(),
		ai:    llm.NewOpenAIClient(),
	}

	app.newConversation()

	return app
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

		case "ctrl+a":
			err := m.addProjectContextToConversation()
			if err != nil {
				m.err = err
				log.Panic(err)
				return m, tea.Quit
			}

		case "ctrl+c":
			m.conversation.CancelStreaming()

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

	m.conversationView.render(&m.conversation)
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

// addProjectContextToConversation adds project context to the current model
// and conversation. Will duplicate context if called multiple times. This is
// just a proof of concept.
func (m *App) addProjectContextToConversation() error {
	m.project = NewProject()

	c, err := m.project.Context()
	if err != nil {
		return err
	}

	m.conversation.AddContext(c)

	return nil
}

func (m *App) processInputView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return cmd
}

func (m *App) newConversation() {
	m.conversation.CancelStreaming()

	m.conversation = llm.MakeConversation()

	cwd, err := os.Getwd()
	if err != nil {
		m.err = err
		return
	}

	m.conversation.AddContext("You're a TUI companion app called Mark. You are direct and to the point. Behave like a professional software engineer. Do not offer any assistance, suggestions, or follow-up questions. Only provide information that is directly requested.")
	m.conversation.AddContext("My name is Juan. Refer to me by name. I'm a software developer with a Computer Science degree. Assume I know advanced computer science concepts and programming languages. DO NOT EXPLAIN BASIC CONCEPTS.")
	m.conversation.AddContext("I'm currently working on a project. I'm in the project's root directory.")
	m.conversation.AddContext("My goal is: 'on app start, already trigger the completion API so the AI actually greets me before I input anything'")

	m.conversation.AddVariable("cwd", cwd)

	m.conversationView.Blur()
	m.input.Focus()
}

func (m *App) submitMessage() tea.Cmd {
	v := m.input.Value()
	if v == "" {
		return nil
	}

	m.input.Reset()

	m.conversation.CancelStreaming()

	// Add user message to chat history
	m.conversation.AddMessage(llm.Message{Role: llm.RoleUser, Content: v})

	// Create a new streaming message
	m.conversation.StreamingMessage = llm.NewStreamingMessage()

	cmds := []tea.Cmd{
		complete(m),      // call completions API
		processStream(m), // start receiving partial message
	}
	return tea.Batch(cmds...)
}

func complete(m *App) tea.Cmd {
	return func() tea.Msg {
		m.ai.CompleteStreaming(&m.conversation)

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
