package app

import (
	"log"
	"os"
	"reflect"

	"ant/pkg/llm"
	"ant/pkg/llmopenai"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

type State int

const (
	StateInput State = iota
	StateNormal
)

type Focused int

const (
	FocusedInput Focused = iota
	FocusedEmptyPanel
)

var (
	borderStyle        = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	focusedBorderStyle = borderStyle.BorderForeground(lipgloss.Color("2"))
)

type App struct {
	// app state
	uiReady        bool
	width, height  int
	mainPanelWidth int
	state          State
	focused        Focused

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
		state:   StateInput,
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
		m.mainPanelWidth = int(float64(msg.Width) * float64(0.67))

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

		case "q":
			if m.state == StateNormal {
				return m, tea.Quit
			}

		case "esc":
			if m.state == StateInput {
				m.changeToStateNormal()
				inputHandled = true
			}

		case "enter":
			if m.state == StateInput {
				cmd := m.submitMessage()
				cmds = append(cmds, cmd)
				inputHandled = true
			}

		case "ctrl+o":
			if m.state == StateNormal {
				m.changeToStateInput()
				inputHandled = true
			}

		case "ctrl+n":
			m.newConversation()
			inputHandled = true

		case "ctrl+c":
			m.cancelStreaming()
			inputHandled = true

		case "shift+j":
			if m.state == StateNormal {
				m.conversationView.LineDown()
				inputHandled = true
			}

		case "shift+k":
			if m.state == StateNormal {
				m.conversationView.LineUp()
				inputHandled = true
			}

		}
	}

	if m.uiReady {
		m.conversationView.SetSize(m.mainPanelWidth-2, m.height-2)

		m.input.SetWidth(m.width - m.mainPanelWidth - 2)

		if m.state == StateInput && !inputHandled {
			cmd := m.processInputView(msg)
			cmds = append(cmds, cmd)
		}

		m.conversationView.render(&m.conversation, m.streaming, m.partialMessage)
	}

	return m, tea.Batch(cmds...)
}

func (m App) View() string {
	if !m.uiReady {
		return "Initializing UI..."
	}

	inputView := m.input.View()

	inputBox := m.borderInput().Render(inputView)
	emptyBox := m.borderEmptyPanel().Render(lipgloss.NewStyle().Width(lipgloss.Width(inputView)).Height(m.height - lipgloss.Height(inputView) - 4).Render(""))

	leftPanel := lipgloss.JoinVertical(lipgloss.Left, inputBox, emptyBox)
	rightPanel := borderStyle.Render(m.conversationView.View())

	return lipgloss.JoinHorizontal(lipgloss.Top,
		leftPanel,
		rightPanel,
	)
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

func (m *App) changeToStateInput() {
	m.state = StateInput
	m.focused = FocusedInput
}

func (m *App) changeToStateNormal() {
	m.state = StateNormal
	m.focused = FocusedEmptyPanel
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

	m.changeToStateInput()
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
