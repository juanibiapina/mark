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
	FocusedPullRequest
	FocusedThreadList
	FocusedThread
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
	project           model.Project
	thread            model.Thread
	threadListEntries []model.ThreadEntry

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

	input          textarea.Model
	threadViewport viewport.Model

	pullRequestViewport viewport.Model

	threadList       viewport.Model
	threadListCursor int

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
	input.Focus()       // focus is actually handled by the app
	input.CharLimit = 0 // no character limit
	input.MaxHeight = 0 // no max height
	input.Cursor.SetMode(cursor.CursorStatic)
	input.Prompt = ""
	input.Styles.Focused.CursorLine = lipgloss.NewStyle() // Remove cursor line styling
	input.ShowLineNumbers = false
	input.KeyMap.InsertNewline.SetEnabled(false)

	// init active thread
	activeThread := model.MakeThread()

	// init project
	project, err := model.MakeProject(cwd)
	if err != nil {
		return App{}, err
	}

	// init app
	app := App{
		project: project,
		db:      MakeFilesystemDatabase(dbdir),
		ai:      openai.NewOpenAIClient(),
		input:   input,
		thread:  activeThread,
	}

	return app, nil
}

func (m App) Err() error {
	return m.err
}

// Init returns an initial command.
func (m App) Init() (tea.Model, tea.Cmd) {
	return m, m.loadThreads()
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
		rest := msg.Height - inputHeight
		half := rest / 2
		m.pullRequestViewport.SetWidth(m.sideBarWidth - borderSize)
		m.pullRequestViewport.SetHeight(half - borderSize)
		m.threadList.SetWidth(m.sideBarWidth - borderSize)
		m.threadList.SetHeight(half - borderSize)
		highlightedEntryStyle = highlightedEntryStyle.Width(m.sideBarWidth - borderSize)

		m.threadViewport.SetWidth(m.mainPanelWidth - 2)   // 2 is the border width
		m.threadViewport.SetHeight(m.mainPanelHeight - 2) // 2 is the border width

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

		m.renderActiveThread()

	case replyMessage:
		m.streaming = false
		m.partialMessage = ""
		m.thread.AddMessage(model.Message{Role: model.RoleAssistant, Content: string(msg)})
		cmd := m.saveThread()
		cmds = append(cmds, cmd)
		m.renderActiveThread()

	case threadEntriesMsg:
		m.threadListEntries = msg
		m.renderThreadList()

	case threadMsg:
		m.thread = msg.thread
		m.renderActiveThread()
		m.renderThreadList()

	case removeThreadMsg:
		// If the active thread is the one being deleted, start a new thread
		if m.thread.ID == string(msg) {
			m.newThread()
			m.renderActiveThread()
		}

		// Remove the thread from the list of entries
		for i, entry := range m.threadListEntries {
			if entry.ID == string(msg) {
				m.threadListEntries = append(m.threadListEntries[:i], m.threadListEntries[i+1:]...)
				break
			}
		}

		// Ensure the cursor is in a valid position
		if m.threadListCursor >= len(m.threadListEntries) {
			m.threadListCursor = len(m.threadListEntries) - 1
		}

		// If we deleted the last message, focus on the input
		if len(m.threadListEntries) == 0 {
			m.focused = FocusedInput
		}

		m.renderThreadList()

	case pullRequestDescriptionMsg:
		m.thread.PullRequest.Description = string(msg)
		cmd := m.saveThread()
		cmds = append(cmds, cmd)

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
				cmd = m.saveThread()
				cmds = append(cmds, cmd)
				m.renderActiveThread()
			}

			if m.focused == FocusedThreadList {
				inputHandled = true

				cmd := m.loadSelectedThread()
				cmds = append(cmds, cmd)
			}

		case "ctrl+l":
			inputHandled = true
			m.thread.Messages = nil
			cmd := m.saveThread()
			cmds = append(cmds, cmd)
			m.renderActiveThread()

		case "ctrl+n":
			m.newThread()
			m.renderActiveThread()
			m.renderThreadList()
			inputHandled = true

		case "ctrl+c":
			m.cancelStreaming()
			m.renderActiveThread()
			inputHandled = true

		}
	}

	if m.uiReady {
		if !inputHandled {
			if m.focused == FocusedInput {
				cmd := m.processInputView(msg)
				cmds = append(cmds, cmd)
			}

			if m.focused == FocusedPullRequest {
				cmd := m.processPullRequestView(msg)
				cmds = append(cmds, cmd)
			}

			if m.focused == FocusedThreadList {
				cmd := m.processThreadList(msg)
				cmds = append(cmds, cmd)
			}

			if m.focused == FocusedThread {
				cmd := m.processThreadView(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	m.renderPullRequest()

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

	m.renderThreadList()
}

func (m *App) focusPrev() {
	m.focused -= 1
	if m.focused < 0 {
		m.focused = FocusedEndMarker - 1
	}

	m.renderThreadList()
}

func (m *App) borderInput() lipgloss.Style {
	if m.focused == FocusedInput {
		return focusedBorderStyle
	}
	return borderStyle
}

func (m *App) borderPullRequest() lipgloss.Style {
	if m.focused == FocusedPullRequest {
		return focusedBorderStyle
	}
	return borderStyle
}

func (m *App) borderThreadList() lipgloss.Style {
	if m.focused == FocusedThreadList {
		return focusedBorderStyle
	}
	return borderStyle
}

func (m *App) borderThread() lipgloss.Style {
	if m.focused == FocusedThread {
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
	m.thread.AddMessage(model.Message{Role: model.RoleAssistant, Content: m.partialMessage})

	m.partialMessage = ""
}

func (m *App) processInputView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return cmd
}

func (m *App) processPullRequestView(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "e":
			cmd, err := m.editPullRequest()
			if err != nil {
				m.err = err
				return tea.Quit
			}

			return cmd
		default:
			var cmd tea.Cmd
			m.pullRequestViewport, cmd = m.pullRequestViewport.Update(msg)
			return cmd
		}
	}

	return nil
}

func (m *App) selectNextThread() {
	if len(m.threadListEntries) == 0 {
		return
	}

	m.threadListCursor++

	if m.threadListCursor >= len(m.threadListEntries) {
		m.threadListCursor = 0
	}

	m.renderThreadList()
}

func (m *App) selectPrevThread() {
	if len(m.threadListEntries) == 0 {
		return
	}

	m.threadListCursor--

	if m.threadListCursor < 0 {
		m.threadListCursor = len(m.threadListEntries) - 1
	}

	m.renderThreadList()
}

func (m *App) processThreadList(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "j":
			m.selectNextThread()
		case "k":
			m.selectPrevThread()
		case "d":
			return m.deleteSelectedThread()
		default:
			var cmd tea.Cmd
			m.threadList, cmd = m.threadList.Update(msg)
			return cmd
		}
	}

	return nil
}

func (m *App) processThreadView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.threadViewport, cmd = m.threadViewport.Update(msg)
	return cmd
}

// newThread starts a new thread
func (m *App) newThread() {
	m.cancelStreaming()

	m.thread = model.MakeThread()

	m.input.Reset()

	m.focused = FocusedInput
}

func (m *App) renderActiveThread() {
	messages := m.thread.Messages

	// create a new glamour renderer
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(m.threadViewport.Width()-2-2), // 2 is the glamour internal gutter, extra 2 for the right side
	)
	if err != nil {
		m.err = err
		return
	}

	var content string

	for i := 0; i < len(messages); i++ {
		var msg string
		if messages[i].Role == model.RoleUser {
			msg = lipgloss.NewStyle().Width(m.threadViewport.Width()).Align(lipgloss.Right).Render(fmt.Sprintf("%s\n", messages[i].Content))
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

	m.threadViewport.SetContent(content)
}

func (m *App) renderPullRequest() {
	m.pullRequestViewport.SetContent(m.thread.PullRequest.Description)
}

func (m *App) renderThreadList() {
	var content string

	for i := 0; i < len(m.threadListEntries); i++ {
		prefix := "  "
		if m.threadListEntries[i].ID == m.thread.ID {
			prefix = "* "
		}
		entryContent := prefix + m.threadListEntries[i].ID

		if i == m.threadListCursor {
			if m.focused == FocusedThreadList {
				entryContent = highlightedEntryStyle.Render(entryContent)
			}
		}

		content += entryContent + "\n"
	}

	m.threadList.SetContent(content)
}
