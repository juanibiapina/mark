package app

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"

	"mark/internal/db"
	"mark/internal/model"
	"mark/internal/util"

	"github.com/charmbracelet/bubbles/v2/textarea"
	"github.com/charmbracelet/bubbles/v2/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss/v2"
)

type Focused int

const (
	FocusedInput Focused = iota
	FocusedThreadList
	FocusedThread
	FocusedEndMarker // used to determine the number of focusable items for cycling
)

const (
	inputHeight = 5
	ratio       = 0.67
)

type (
	eventMsg            struct{ msg tea.Msg }
	threadEntriesMsg    []model.ThreadEntry
	streamChunkReceived string
	streamFinished      string
	threadMsg           struct{ thread model.Thread }
	errMsg              struct{ err error }
)

var (
	textColor  = lipgloss.NoColor{}
	focusColor = lipgloss.Color("2")

	textStyle              = lipgloss.NewStyle().Foreground(textColor)
	focusedPanelTitleStyle = lipgloss.NewStyle().Foreground(focusColor).Bold(true)

	borderStyle        = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	focusedBorderStyle = borderStyle.BorderForeground(focusColor)

	highlightedEntryStyle = lipgloss.NewStyle().Background(lipgloss.Color("4"))
)

type App struct {
	// models
	thread            model.Thread
	threadListEntries []model.ThreadEntry

	// ui
	uiReady         bool
	focused         Focused
	mainPanelWidth  int
	mainPanelHeight int
	sideBarWidth    int

	input          textarea.Model
	threadViewport viewport.Model

	threadList       viewport.Model
	threadListCursor int

	// clients
	agent *Agent
	db    db.Database

	// events
	events chan tea.Msg

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
	input.Prompt = ""
	input.Styles.Focused.CursorLine = lipgloss.NewStyle() // Remove cursor line styling
	input.ShowLineNumbers = false
	input.KeyMap.InsertNewline.SetEnabled(false)

	// init active thread
	activeThread := model.MakeThread()

	// init events channel
	events := make(chan tea.Msg)

	// init app
	app := App{
		db:     db.MakeDatabase(dbdir),
		agent:  NewAgent(events),
		input:  input,
		thread: activeThread,
		events: events,
	}

	return app, nil
}

func (m App) Err() error {
	return m.err
}

// Init returns an initial command.
func (m App) Init() tea.Cmd {
	return tea.Batch(m.loadThreads(), processEvents(m.events))
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var inputHandled bool // whether the key event was handled and shouldn't be passed to the input view

	msg, cmd := m.processEventMessage(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case errMsg:
		m.err = msg.err
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.handleWindowSize(msg.Width, msg.Height)

	case streamChunkReceived:
		m.thread.AppendChunk(string(msg))

	case streamFinished:
		m.thread.FinishStreaming(string(msg))
		cmd := m.saveThread()
		cmds = append(cmds, cmd)

	case threadEntriesMsg:
		m.threadListEntries = msg

	case threadMsg:
		m.thread = msg.thread

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

		case "space":
			if m.focused == FocusedThreadList {
				inputHandled = true

				cmd := m.loadSelectedThread()
				cmds = append(cmds, cmd)
			}

		case "enter":
			inputHandled = true

			cmd := m.submitMessage()
			cmds = append(cmds, cmd)
			cmd = m.saveThread()
			cmds = append(cmds, cmd)

		case "ctrl+l":
			inputHandled = true
			m.thread.Messages = nil
			cmd := m.saveThread()
			cmds = append(cmds, cmd)

		case "ctrl+n":
			m.newThread()
			inputHandled = true

			cmd := cancelStreaming(m.agent)
			cmds = append(cmds, cmd)

		case "ctrl+c":
			inputHandled = true

			cmd := cancelStreaming(m.agent)
			cmds = append(cmds, cmd)
		}
	}

	if m.uiReady {
		if !inputHandled {
			if m.focused == FocusedInput {
				cmd := m.processInputView(msg)
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

	m.renderActiveThread()
	m.renderThreadList()

	return m, tea.Batch(cmds...)
}

func (m App) View() string {
	if !m.uiReady {
		return "Initializing..."
	}

	return m.windowView()
}

// processEventMessage checks if the message is an event message, so we can restart the
// event processing go routine. Returns the message to be processed normally.
func (m App) processEventMessage(msg tea.Msg) (tea.Msg, tea.Cmd) {
	switch msg := msg.(type) {
	case eventMsg:
		return msg.msg, processEvents(m.events)
	default:
		return msg, nil
	}
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

func (m *App) processInputView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return cmd
}

func (m *App) selectNextThread() {
	if len(m.threadListEntries) == 0 {
		return
	}

	m.threadListCursor++

	if m.threadListCursor >= len(m.threadListEntries) {
		m.threadListCursor = 0
	}
}

func (m *App) selectPrevThread() {
	if len(m.threadListEntries) == 0 {
		return
	}

	m.threadListCursor--

	if m.threadListCursor < 0 {
		m.threadListCursor = len(m.threadListEntries) - 1
	}
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
			return nil
		}
	}

	return nil
}

func (m *App) processThreadView(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "e":
			cmd, err := m.viewThreadInEditor()
			if err != nil {
				m.err = err
				return tea.Quit
			}

			return cmd
		default:
			var cmd tea.Cmd
			m.threadViewport, cmd = m.threadViewport.Update(msg)
			return cmd
		}
	}

	return nil
}

// newThread starts a new thread
func (m *App) newThread() {
	m.thread = model.MakeThread()

	m.input.Reset()

	m.focused = FocusedInput
}

func (m *App) renderActiveThread() {
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

	for _, message := range m.thread.Messages {
		var msg string
		if message.Role == model.RoleUser {
			msg = lipgloss.NewStyle().Width(m.threadViewport.Width()).Align(lipgloss.Right).Render(fmt.Sprintf("%s\n", message.Content))
		} else {
			msg, err = renderer.Render(message.Content)
			if err != nil {
				log.Fatal(err)
			}
		}

		content += msg
	}

	if m.thread.PartialMessage != "" {
		c, err := renderer.Render(m.thread.PartialMessage)
		if err != nil {
			log.Fatal(err)
		}

		content += c
	}

	m.threadViewport.SetContent(content)
}

func (m *App) renderThreadList() {
	var content string

	for i, entry := range m.threadListEntries {
		prefix := "  "
		if entry.ID == m.thread.ID {
			prefix = "* "
		}
		entryContent := prefix + entry.ID

		if i == m.threadListCursor {
			if m.focused == FocusedThreadList {
				entryContent = highlightedEntryStyle.Render(entryContent)
			}
		}

		content += entryContent + "\n"
	}

	m.threadList.SetContent(content)
}

func (m *App) submitMessage() tea.Cmd {
	m.thread.CancelStreaming()

	v := m.input.Value()
	if v != "" {
		m.thread.AddMessage(model.Message{Role: model.RoleUser, Content: v})
		m.input.Reset()
	}

	return complete(m)
}

func (m *App) viewThreadInEditor() (tea.Cmd, error) {
	// build the content
	var content string
	for _, msg := range m.thread.Messages {
		content += "---\n"
		content += msg.Content + "\n"
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		return nil, nil
	}

	tmpdir, err := os.MkdirTemp("", "mark-*")
	if err != nil {
		return nil, err
	}

	tmpFile := path.Join(tmpdir, "mark-thread")
	err = os.WriteFile(tmpFile, []byte(content), 0o644)
	if err != nil {
		return nil, err
	}

	c := exec.Command(editor, tmpFile)

	return tea.ExecProcess(c, func(err error) tea.Msg {
		defer os.RemoveAll(tmpdir)

		if err != nil {
			return errMsg{err}
		}

		return nil
	}), nil
}

func (m *App) windowView() string {
	return lipgloss.JoinHorizontal(lipgloss.Top, m.sidebarView(), m.mainView())
}

func (m *App) sidebarView() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.inputView(), m.threadListView())
}

func (m *App) mainView() string {
	return util.RenderBorderWithTitle(
		m.threadView(),
		m.borderIfFocused(FocusedThread),
		"Thread",
		m.panelTitleStyleIfFocused(FocusedThread),
	)
}

func (m *App) inputView() string {
	return util.RenderBorderWithTitle(
		m.input.View(),
		m.borderIfFocused(FocusedInput),
		"Message Assistant",
		m.panelTitleStyleIfFocused(FocusedInput),
	)
}

func (m *App) threadListView() string {
	return util.RenderBorderWithTitle(
		m.threadList.View(),
		m.borderIfFocused(FocusedThreadList),
		"Threads",
		m.panelTitleStyleIfFocused(FocusedThreadList),
	)
}

func (m *App) panelTitleStyleIfFocused(focused Focused) lipgloss.Style {
	if m.focused == focused {
		return focusedPanelTitleStyle
	}
	return textStyle
}

func (m *App) borderIfFocused(focused Focused) lipgloss.Style {
	if m.focused == focused {
		return focusedBorderStyle
	}
	return borderStyle
}

func (m *App) threadView() string {
	return m.threadViewport.View()
}

func (m *App) handleWindowSize(width, height int) {
	borderSize := 2 // 2 times the border width

	m.mainPanelWidth = int(float64(width) * ratio)
	m.mainPanelHeight = height
	m.sideBarWidth = width - m.mainPanelWidth

	m.input.SetWidth(m.sideBarWidth - borderSize)
	m.input.SetHeight(inputHeight - borderSize)
	rest := height - inputHeight
	m.threadList.SetWidth(m.sideBarWidth - borderSize)
	m.threadList.SetHeight(rest - borderSize)
	highlightedEntryStyle = highlightedEntryStyle.Width(m.sideBarWidth - borderSize)

	m.threadViewport.SetWidth(m.mainPanelWidth - 2)   // 2 is the border width
	m.threadViewport.SetHeight(m.mainPanelHeight - 2) // 2 is the border width

	if !m.uiReady {
		m.uiReady = true
	}
}
