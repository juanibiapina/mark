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

	// streaming
	streaming      bool
	stream         *llm.StreamingMessage
	partialMessage string

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

	app.project = NewProject()
	app.newConversation()
	app.startStreaming() // need to start streaming here because Init can't make changes to the model

	return app
}

// Init Init is required by the bubbletea interface. It is called once when the
// program starts. It is used to send an initial command to the update function.
// Note: Modifications to the model here are lost since there's no way to return
// the updated model like in Update.
func (m App) Init() tea.Cmd {
	return tea.Batch(
		complete(&m),      // call completions API
		processStream(&m), // start receiving partial message
	)
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
		if !m.streaming {
			return m, nil
		}

		m.partialMessage += string(msg)

		cmds = append(cmds, processStream(&m))

	case replyMessage:
		// Ignore message if streaming has been cancelled
		if !m.streaming {
			return m, nil
		}

		m.streaming = false
		m.partialMessage = ""
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

	m.conversationView.render(&m.conversation, m.streaming, m.partialMessage)
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

func (m *App) startStreaming() {
	m.stream = llm.NewStreamingMessage()
	m.streaming = true
}

func (m *App) cancelStreaming() {
	if m.stream != nil {
		m.stream.Cancel()
	}
	m.stream = nil

	m.streaming = false

	// Add the partial message to the chat history
	m.conversation.AddMessage(llm.Message{Role: llm.RoleAssistant, Content: m.partialMessage})

	m.partialMessage = ""
}

func (m *App) processInputView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return cmd
}

func (m *App) newConversation() {
	m.cancelStreaming()

	m.conversation = llm.MakeConversation()

	m.conversation.AddContext("You're a TUI companion app called Mark. You are direct and to the point. Behave like a staff software engineer. Do not offer any assistance, suggestions, or follow-up questions. Only provide information that is directly requested.")
	m.conversation.AddContext("My name is Juan. Refer to me by name. I'm a software developer with a Computer Science degree. Assume I know advanced computer science concepts and programming languages. DO NOT EXPLAIN BASIC CONCEPTS.")
	m.conversation.AddContext("I'm currently working on a software project. I'm in the project's root directory.")
	m.conversation.AddContext("Greet me and explain the current state of the project.")

	c, err := m.project.Context()
	if err != nil {
		m.err = err
		log.Panic(err)
	}

	m.conversation.AddContext(c)


	m.conversationView.Blur()
	m.input.Focus()
}

func (m *App) submitMessage() tea.Cmd {
	v := m.input.Value()
	if v == "" {
		return nil
	}

	m.input.Reset()

	m.cancelStreaming()

	// Add user message to chat history
	m.conversation.AddMessage(llm.Message{Role: llm.RoleUser, Content: v})

	m.startStreaming()

	cmds := []tea.Cmd{
		complete(m),      // call completions API
		processStream(m), // start receiving partial message
	}
	return tea.Batch(cmds...)
}

func complete(m *App) tea.Cmd {
	return func() tea.Msg {
		m.ai.CompleteStreaming(&m.conversation, m.stream)

		return nil
	}
}

func processStream(m *App) tea.Cmd {
	return func() tea.Msg {
		select {
		case v := <-m.stream.Reply:
			return replyMessage(v)
		case v := <-m.stream.Chunks:
			return partialMessage(v)
		}
	}
}
