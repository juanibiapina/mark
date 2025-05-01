package app

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"

	"mark/internal/db"
	"mark/internal/model"
	"mark/internal/openai"
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
	threadEntriesMsg []model.ThreadEntry
	partialMessage   string
	replyMessage     string
	threadMsg        struct{ thread model.Thread }
	errMsg           struct{ err error }
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

	threadList       viewport.Model
	threadListCursor int

	// clients
	ai *openai.OpenAI
	db db.Database

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

	// init app
	app := App{
		db:     db.MakeDatabase(dbdir),
		ai:     openai.NewOpenAIClient(),
		input:  input,
		thread: activeThread,
	}

	return app, nil
}

func (m App) Err() error {
	return m.err
}

// Init returns an initial command.
func (m App) Init() tea.Cmd {
	return m.loadThreads()
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var inputHandled bool // whether the key event was handled and shouldn't be passed to the input view

	switch msg := msg.(type) {

	case errMsg:
		m.err = msg.err
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.handleWindowSize(msg.Width, msg.Height)

	case partialMessage:
		// Ignore message if streaming has been cancelled
		if !m.streaming {
			return m, nil
		}

		m.partialMessage += string(msg)

		cmds = append(cmds, processStream(&m))

	case replyMessage:
		m.streaming = false
		m.partialMessage = ""
		m.thread.AddMessage(model.Message{Role: model.RoleAssistant, Content: string(msg)})
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

		case "ctrl+c":
			m.cancelStreaming()
			inputHandled = true

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

func (m *App) saveThread() tea.Cmd {
	return func() tea.Msg {
		err := m.db.SaveThread(m.thread)
		if err != nil {
			return errMsg{err}
		}

		return m.loadThreads()()
	}
}

func (m *App) loadSelectedThread() tea.Cmd {
	if len(m.threadListEntries) == 0 {
		return nil
	}

	selectedEntry := m.threadListEntries[m.threadListCursor]

	return func() tea.Msg {
		thread, err := m.db.LoadThread(selectedEntry.ID)
		if err != nil {
			return errMsg{err}
		}

		return threadMsg{thread}
	}
}

func (m *App) loadThreads() tea.Cmd {
	return func() tea.Msg {
		threads, err := m.db.ListThreads()
		if err != nil {
			return errMsg{err}
		}

		return threadEntriesMsg(threads)
	}
}

func (m *App) submitMessage() tea.Cmd {
	m.cancelStreaming()

	v := m.input.Value()
	if v != "" {
		m.thread.AddMessage(model.Message{Role: model.RoleUser, Content: v})
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

func (m *App) deleteSelectedThread() tea.Cmd {
	if len(m.threadListEntries) == 0 {
		return nil
	}

	selectedEntryID := m.threadListEntries[m.threadListCursor].ID

	// Remove the thread from the list of entries
	for i, entry := range m.threadListEntries {
		if entry.ID == selectedEntryID {
			m.threadListEntries = append(m.threadListEntries[:i], m.threadListEntries[i+1:]...)
			break
		}
	}

	// Ensure the cursor is in a valid position
	if len(m.threadListEntries) == 0 {
		m.threadListCursor = 0
	} else {
		m.threadListCursor = util.Clamp(m.threadListCursor, 0, len(m.threadListEntries)-1)
	}

	m.renderActiveThread()
	m.renderThreadList()

	return func() tea.Msg {
		err := m.db.DeleteThread(selectedEntryID)
		if err != nil {
			return errMsg{err}
		}

		return nil
	}
}

func complete(m *App) tea.Cmd {
	return func() tea.Msg {
		err := m.ai.CompleteStreaming(&m.thread, m.stream)
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
