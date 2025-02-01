package app

import (
	"log"
	"os"
	"reflect"

	"mark/pkg/llm"
	"mark/pkg/llmopenai"
	"mark/pkg/view"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

type Focused int

const (
	FocusedInput Focused = iota
	FocusedEmptyPanel
	FocusedConversation
	FocusedEndMarker // used to determine the number of focused items
)

var (
	borderStyle        = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	focusedBorderStyle = borderStyle.BorderForeground(lipgloss.Color("2"))
)

type App struct {
	// app state
	uiReady       bool
	width, height int
	focused       Focused

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
		focused: FocusedInput,
		input:   MakeInput(),
		ai:      llmopenai.NewOpenAIClient(),
	}

	app.project = NewProject()
	app.newConversation()

	return app
}

// Init Init is required by the bubbletea interface. It is called once when the
// program starts. It is used to send an initial command to the update function.
// Note: Modifications to the model here are lost since there's no way to return
// the updated model like in Update.
func (m App) Init() (tea.Model, tea.Cmd) {
	return m, nil
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if len(os.Getenv("DEBUG")) > 0 {
		log.Print("msg: ", reflect.TypeOf(msg), msg)
	}

	var cmds []tea.Cmd
	var inputHandled bool // whether the key event was handled and shouldn't be passed to the input view

	switch msg := msg.(type) {

	case errMsg:
		m.err = msg.err
		log.Panic(msg.err)
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.uiReady {
			m.uiReady = true
		}

	case ReplaceLine:
		// replace the line in the file
		err := msg.Invoke()
		if err != nil {
			panic(err)
		}

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

	case tea.KeyPressMsg:
		switch msg.String() {

		case "esc":
			return m, tea.Quit

		case "tab":
			m.focusNext()
			inputHandled = true

		case "shift+tab":
			m.focusPrev()
			inputHandled = true

		case "enter":
			if m.focused == FocusedInput {
				cmd := m.submitMessage()
				cmds = append(cmds, cmd)
				inputHandled = true
			}

		case "ctrl+n":
			m.newConversation()
			inputHandled = true

		case "ctrl+c":
			m.cancelStreaming()
			inputHandled = true

		}
	}

	if m.uiReady {
		if m.focused == FocusedInput && !inputHandled {
			cmd := m.processInputView(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m App) View() string {
	if !m.uiReady {
		return "Initializing UI..."
	}

	// TODO still weird that I need to do this in a view method
	m.conversationView.Set(&m.conversation, m.streaming, m.partialMessage)

	main := view.Main{
		Left: view.Sidebar{
			Input: view.NewPane(m.input, m.borderInput(), "Message Assistant"),
			Empty: view.NewPane(view.Space{}, m.borderEmptyPanel(), ""),
		},
		Right: view.NewPane(m.conversationView, m.borderConversation(), "Conversation"),
		Ratio: 0.67,
	}

	return main.Render(m.width, m.height)
}

func (m *App) focusNext() {
	m.focused += 1
	if m.focused == FocusedEndMarker {
		m.focused = FocusedInput
	}
}

func (m *App) focusPrev() {
	m.focused -= 1
	if m.focused == FocusedEndMarker {
		m.focused = FocusedInput
	}
}

func (m *App) borderInput() lipgloss.Style {
	if m.focused == FocusedInput {
		return focusedBorderStyle
	}
	return borderStyle
}

func (m *App) borderEmptyPanel() lipgloss.Style {
	if m.focused == FocusedEmptyPanel {
		return focusedBorderStyle
	}
	return borderStyle
}

func (m *App) borderConversation() lipgloss.Style {
	if m.focused == FocusedConversation {
		return focusedBorderStyle
	}
	return borderStyle
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

	m.conversation.SetContext("assistant", "You're a TUI companion app called Mark. You are direct and to the point. Behave like a staff software engineer. Do not offer any assistance, suggestions, or follow-up questions. Only provide information that is directly requested.")
	m.conversation.SetContext("user", "My name is Juan. Refer to me by name. I'm a software developer with a Computer Science degree. Assume I know advanced computer science concepts and programming languages. DO NOT EXPLAIN BASIC CONCEPTS.")
	m.conversation.SetContext("prompt", "I'm currently working on a software project. I'm in the project's root directory. If there are changes, explain the git diff.")

	c, err := m.project.Context()
	if err != nil {
		m.err = err
		log.Panic(err)
	}

	m.conversation.SetContext("project", c)

	m.input.Reset()

	m.focused = FocusedInput
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

	c, err := m.project.Context()
	if err != nil {
		m.err = err
		log.Panic(err)
	}

	m.conversation.SetContext("project", c)

	cmds := []tea.Cmd{
		complete(m),      // call completions API
		processStream(m), // start receiving partial message
	}
	return tea.Batch(cmds...)
}

func complete(m *App) tea.Cmd {
	return func() tea.Msg {
		//rs := llm.ResponseSchema{
		//	Name: "replace_line",
		//	Description: "Replace a line in a file",
		//	Schema: ReplaceLineResponseSchema,
		//}

		// replaceLine := ReplaceLine{}

		err := m.ai.CompleteStreaming(&m.conversation, m.stream)
		if err != nil {
			return errMsg{err}
		}

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
