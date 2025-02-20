package app

import (
	"fmt"
	"log"
	"path"

	"mark/pkg/model"
	"mark/pkg/openai"

	"github.com/charmbracelet/bubbles/v2/cursor"
	"github.com/charmbracelet/bubbles/v2/textarea"
	"github.com/charmbracelet/bubbles/v2/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss/v2"
)

type Focused int

const (
	FocusedInput Focused = iota
	FocusedConversationList
	FocusedConversation
	FocusedEndMarker // used to determine the number of focusable items for cycling
)

const (
	inputHeight = 5
	ratio       = 0.67
)

var (
	borderStyle           = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	focusedBorderStyle    = borderStyle.BorderForeground(lipgloss.Color("2"))
	highlightedEntryStyle = lipgloss.NewStyle().Background(lipgloss.Color("4"))
)

type App struct {
	// models
	conversation        model.Conversation
	conversationEntries []model.ConversationEntry

	// streaming
	streaming      bool
	stream         *model.StreamingMessage
	partialMessage string

	// ui
	uiReady         bool
	focused         Focused
	mainPanelWidth  int
	mainPanelHeight int
	sideBarWidth    int

	input                textarea.Model
	conversationViewport viewport.Model

	conversationList viewport.Model
	cursorEntries    int

	// clients
	ai *openai.OpenAI
	db Database

	// error
	err error
}

func MakeApp(cwd string) (App, error) {
	// determine database directory
	dbdir := path.Join(cwd, ".mark")

	// init input
	input := textarea.New()
	input.Focus() // focus is actually handled by the app
	input.Cursor.SetMode(cursor.CursorStatic)
	input.Prompt = ""
	input.Styles.Focused.CursorLine = lipgloss.NewStyle() // Remove cursor line styling
	input.ShowLineNumbers = false
	input.KeyMap.InsertNewline.SetEnabled(false)

	// init conversation
	conversation := model.MakeConversation()

	// init app
	app := App{
		db:           MakeFilesystemDatabase(dbdir),
		ai:           openai.NewOpenAIClient(),
		input:        input,
		conversation: conversation,
	}

	return app, nil
}

func (m App) Err() error {
	return m.err
}

// Init returns an initial command.
func (m App) Init() (tea.Model, tea.Cmd) {
	return m, m.loadConversations()
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var inputHandled bool // whether the key event was handled and shouldn't be passed to the input view

	switch msg := msg.(type) {

	case errMsg:
		m.err = msg.err
		return m, tea.Quit

	case tea.WindowSizeMsg:
		borderSize := 2 // 2 times the border width

		m.mainPanelWidth = int(float64(msg.Width) * ratio)
		m.mainPanelHeight = msg.Height
		m.sideBarWidth = msg.Width - m.mainPanelWidth

		m.input.SetWidth(m.sideBarWidth - borderSize)
		m.input.SetHeight(inputHeight - borderSize)
		m.conversationList.SetWidth(m.sideBarWidth - borderSize)
		m.conversationList.SetHeight(msg.Height - inputHeight - borderSize)
		highlightedEntryStyle = highlightedEntryStyle.Width(m.sideBarWidth - borderSize)

		m.conversationViewport.SetWidth(m.mainPanelWidth - 2)   // 2 is the border width
		m.conversationViewport.SetHeight(m.mainPanelHeight - 2) // 2 is the border width

		if !m.uiReady {
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
		cmd := m.saveConversation()
		cmds = append(cmds, cmd)
		m.renderConversation()

	case conversationEntriesMsg:
		m.conversationEntries = msg
		m.renderConversationList()

	case conversationMsg:
		m.conversation = msg.conversation
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
				inputHandled = true

				v := m.input.Value()
				if v == "" {
					break
				}

				cmd := m.submitMessage()
				cmds = append(cmds, cmd)
				cmd = m.saveConversation()
				cmds = append(cmds, cmd)
				m.renderConversation()
			}

			if m.focused == FocusedConversationList {
				inputHandled = true

				cmd := m.loadSelectedConversation()
				cmds = append(cmds, cmd)
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
		if !inputHandled {
			if m.focused == FocusedInput {
				cmd := m.processInputView(msg)
				cmds = append(cmds, cmd)
			}

			if m.focused == FocusedConversationList {
				cmd := m.processConversationList(msg)
				cmds = append(cmds, cmd)
			}

			if m.focused == FocusedConversation {
				cmd := m.processConversationView(msg)
				cmds = append(cmds, cmd)
			}
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

	m.renderConversationList()
}

func (m *App) focusPrev() {
	m.focused -= 1
	if m.focused < 0 {
		m.focused = FocusedEndMarker - 1
	}

	m.renderConversationList()
}

func (m *App) borderInput() lipgloss.Style {
	if m.focused == FocusedInput {
		return focusedBorderStyle
	}
	return borderStyle
}

func (m *App) borderConversationList() lipgloss.Style {
	if m.focused == FocusedConversationList {
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

func (m *App) selectNextConversation() {
	if len(m.conversationEntries) == 0 {
		return
	}

	m.cursorEntries++

	if m.cursorEntries >= len(m.conversationEntries) {
		m.cursorEntries = 0
	}

	m.renderConversationList()
}

func (m *App) selectPrevConversation() {
	if len(m.conversationEntries) == 0 {
		return
	}

	m.cursorEntries--

	if m.cursorEntries < 0 {
		m.cursorEntries = len(m.conversationEntries) - 1
	}

	m.renderConversationList()
}

func (m *App) processConversationList(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "j":
			m.selectNextConversation()
		case "k":
			m.selectPrevConversation()
		default:
			var cmd tea.Cmd
			m.conversationList, cmd = m.conversationList.Update(msg)
			return cmd
		}
	}

	return nil
}

func (m *App) processConversationView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.conversationViewport, cmd = m.conversationViewport.Update(msg)
	return cmd
}

// newConversation starts a new conversation
func (m *App) newConversation() {
	m.cancelStreaming()

	m.conversation = model.MakeConversation()

	m.input.Reset()

	m.focused = FocusedInput
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

func (m *App) renderConversationList() {
	var content string

	for i := 0; i < len(m.conversationEntries); i++ {
		entryContent := m.conversationEntries[i].ID

		if i == m.cursorEntries {
			if m.focused == FocusedConversationList {
				entryContent = highlightedEntryStyle.Render(entryContent)
			}
		}

		content += entryContent + "\n"
	}

	m.conversationList.SetContent(content)
}
