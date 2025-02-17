package app

import (
	"fmt"
	"log"
	"os"
	"reflect"

	"mark/pkg/model"
	"mark/pkg/openai"

	"github.com/charmbracelet/bubbles/v2/textarea"
	"github.com/charmbracelet/bubbles/v2/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type Focused int

const (
	FocusedInput Focused = iota
	FocusedPromptList
	FocusedConversation
	FocusedEndMarker // used to determine the number of focusable items for cycling
)

const (
	inputHeight = 5
	ratio       = 0.67
)

var (
	borderStyle        = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	focusedBorderStyle = borderStyle.BorderForeground(lipgloss.Color("2"))
)

type App struct {
	// config
	config Config

	// models
	prompts      []model.Prompt
	conversation model.Conversation

	// streaming
	streaming      bool
	stream         *model.StreamingMessage
	partialMessage string

	// ui
	uiReady              bool
	focused              Focused
	mainPanelWidth       int
	mainPanelHeight      int
	sideBarWidth         int
	inputWidth           int
	promptListWidth      int
	promptListHeight     int
	conversationViewport viewport.Model
	input                textarea.Model
	promptList           viewport.Model
	selectedPromptIndex  int

	// llm
	ai *openai.OpenAI

	// error
	err error
}

func MakeApp(config Config) (App, error) {
	app := App{
		config: config,
		ai:     openai.NewOpenAIClient(),
	}

	return app, nil
}

func (m App) Err() error {
	return m.err
}

// Init initializes the App model and possibly returns an initial command.
func (m App) Init() (tea.Model, tea.Cmd) {
	err := m.initPrompts()
	if err != nil {
		m.err = err
		return m, tea.Quit
	}

	m.initInput()
	m.newConversation()

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
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.mainPanelWidth = int(float64(msg.Width) * ratio)
		m.mainPanelHeight = msg.Height
		m.sideBarWidth = msg.Width - m.mainPanelWidth

		m.inputWidth = m.sideBarWidth
		m.input.SetWidth(m.inputWidth - 2)                   // 2 is the border width
		m.input.SetHeight(inputHeight - 2)                   // 2 is the border width
		m.promptList.SetWidth(m.sideBarWidth - 2)            // 2 is the border width
		m.promptList.SetHeight(msg.Height - inputHeight - 2) // 2 is the border width

		m.promptListWidth = m.sideBarWidth
		m.promptListHeight = msg.Height - inputHeight
		m.conversationViewport.SetWidth(m.mainPanelWidth - 2)   // 2 is the border width
		m.conversationViewport.SetHeight(m.mainPanelHeight - 2) // 2 is the border width

		if !m.uiReady {
			m.renderPrompts()

			m.uiReady = true
		}

	case partialMessage:
		// Ignore message if streaming has been cancelled
		if !m.streaming {
			return m, nil
		}

		m.partialMessage += string(msg)

		cmds = append(cmds, processStream(&m))

		m.renderConversation()

	case replyMessage:
		m.streaming = false
		m.partialMessage = ""
		m.conversation.AddMessage(model.Message{Role: model.RoleAssistant, Content: string(msg)})
		m.renderConversation()

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
				m.renderConversation()
				inputHandled = true
			}

			if m.focused == FocusedPromptList {
				cmd := m.selectCurrentPrompt()
				cmds = append(cmds, cmd)
				inputHandled = true
			}

		case "ctrl+n":
			m.newConversation()
			m.renderConversation()
			inputHandled = true

		case "ctrl+c":
			m.cancelStreaming()
			m.renderConversation()
			inputHandled = true

		}
	}

	if m.uiReady {
		if m.focused == FocusedInput && !inputHandled {
			cmd := m.processInputView(msg)
			cmds = append(cmds, cmd)
		}

		if m.focused == FocusedPromptList {
			cmd := m.processPromptListView(msg)
			cmds = append(cmds, cmd)
		}

		if m.focused == FocusedConversation {
			cmd := m.processConversationView(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m App) View() string {
	if !m.uiReady {
		return "Initializing..."
	}

	return m.windowView()
}

func (m *App) focusNext() {
	m.focused += 1
	if m.focused == FocusedEndMarker {
		m.focused = 0
	}
}

func (m *App) focusPrev() {
	m.focused -= 1
	if m.focused < 0 {
		m.focused = FocusedEndMarker - 1
	}
}

func (m *App) borderInput() lipgloss.Style {
	if m.focused == FocusedInput {
		return focusedBorderStyle
	}
	return borderStyle
}

func (m *App) borderPromptList() lipgloss.Style {
	if m.focused == FocusedPromptList {
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
	m.stream = model.NewStreamingMessage()
	m.streaming = true
}

func (m *App) cancelStreaming() {
	if m.stream != nil {
		m.stream.Cancel()
	}
	m.stream = nil

	m.streaming = false

	// Add the partial message to the chat history
	m.conversation.AddMessage(model.Message{Role: model.RoleAssistant, Content: m.partialMessage})

	m.partialMessage = ""
}

func (m *App) processInputView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return cmd
}

func (m *App) processPromptListView(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j":
			m.selectNextPrompt()
			m.renderPrompts()
		case "k":
			m.selectPrevPrompt()
			m.renderPrompts()
		}
	}

	return nil
}

func (m *App) processConversationView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.conversationViewport, cmd = m.conversationViewport.Update(msg)
	return cmd
}

func (m *App) selectNextPrompt() {
	if len(m.prompts) == 0 {
		return
	}

	m.selectedPromptIndex += 1
	if m.selectedPromptIndex >= len(m.prompts) {
		m.selectedPromptIndex = 0
	}
}

func (m *App) selectPrevPrompt() {
	if len(m.prompts) == 0 {
		return
	}

	m.selectedPromptIndex -= 1
	if m.selectedPromptIndex < 0 {
		m.selectedPromptIndex = len(m.prompts) - 1
	}
}

// newConversation starts a new conversation without a prompt
func (m *App) newConversation() {
	m.cancelStreaming()

	m.conversation = model.MakeConversation()

	m.input.Reset()

	m.focused = FocusedInput
}

func (m *App) selectCurrentPrompt() tea.Cmd {
	m.conversation.Prompt = m.prompts[m.selectedPromptIndex]
	return nil
}

func (m *App) submitMessage() tea.Cmd {
	m.cancelStreaming()

	v := m.input.Value()
	if v != "" {
		m.conversation.AddMessage(model.Message{Role: model.RoleUser, Content: v})
		m.input.Reset()
	}

	// maybe update the prompt here

	m.startStreaming()

	cmds := []tea.Cmd{
		complete(m),      // call completions API
		processStream(m), // start receiving partial message
	}
	return tea.Batch(cmds...)
}

func complete(m *App) tea.Cmd {
	return func() tea.Msg {
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

func (m *App) renderConversation() {
	messages := m.conversation.Messages

	// create a new glamour renderer
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(m.conversationViewport.Width()-2-2), // 2 is the glamour internal gutter, extra 2 for the right side
	)
	if err != nil {
		m.err = err
		return
	}


	var content string

	for i := 0; i < len(messages); i++ {
		var msg string
		if messages[i].Role == model.RoleUser {
			msg = lipgloss.NewStyle().Width(m.conversationViewport.Width()).Align(lipgloss.Right).Render(fmt.Sprintf("%s\n", messages[i].Content))
		} else {
			msg, err = renderer.Render(messages[i].Content)
			if err != nil {
				log.Fatal(err)
			}
		}

		content += msg
	}

	if m.streaming {
		c, err := renderer.Render(m.partialMessage)
		if err != nil {
			log.Fatal(err)
		}

		content += c
	}

	m.conversationViewport.SetContent(content)
}

func (m *App) renderPrompts() {
	var content string

	for index, prompt := range m.prompts {
		name := prompt.Name()

		var prefix string
		if index == m.selectedPromptIndex {
			prefix = " "
		} else {
			prefix = "  "
		}

		content += prefix + name + "\n"

	}

	m.promptList.SetContent(content)
}
