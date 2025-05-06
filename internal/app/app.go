package app

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"

	"mark/internal/llm"
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
	FocusedMessages
	FocusedEndMarker // used to determine the number of focusable items for cycling
)

const (
	inputHeight = 5
)

type (
	eventMsg            struct{ msg tea.Msg }
	streamChunkReceived string
	streamFinished      string
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
	session llm.Session

	agent  *Agent
	events chan tea.Msg
	err    error

	// ui
	uiReady         bool
	focused         Focused
	mainPanelWidth  int
	mainPanelHeight int
	sideBarWidth    int

	input            textarea.Model
	messagesViewport viewport.Model
}

func MakeApp(cwd string) (App, error) {
	// init input
	input := textarea.New()
	input.Focus()       // focus is actually handled by the app
	input.CharLimit = 0 // no character limit
	input.MaxHeight = 0 // no max height
	input.Prompt = ""
	input.Styles.Focused.CursorLine = lipgloss.NewStyle() // Remove cursor line styling
	input.ShowLineNumbers = false
	input.KeyMap.InsertNewline.SetEnabled(false)

	// init events channel
	events := make(chan tea.Msg)

	// init app
	app := App{
		agent:   NewAgent(events),
		input:   input,
		session: llm.MakeSession(),
		events:  events,
	}

	return app, nil
}

func (m App) Err() error {
	return m.err
}

func (m App) Init() tea.Cmd {
	return processEvents(m.events)
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
		if m.session.IsStreaming() {
			m.session.AppendChunk(string(msg))
		}

	case streamFinished:
		m.session.FinishStreaming(string(msg))

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
			inputHandled = true

			cmd := m.submitMessage()
			cmds = append(cmds, cmd)

		case "ctrl+n":
			m.newSession()
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

			if m.focused == FocusedMessages {
				cmd := m.processMessagesView(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	m.renderMessagesView()

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

func (m *App) processMessagesView(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "e":
			cmd, err := m.viewMessagesInEditor()
			if err != nil {
				m.err = err
				return tea.Quit
			}

			return cmd
		default:
			var cmd tea.Cmd
			m.messagesViewport, cmd = m.messagesViewport.Update(msg)
			return cmd
		}
	}

	return nil
}

func (m *App) newSession() {
	m.session = llm.MakeSession()

	m.input.Reset()

	m.focused = FocusedInput
}

func (m *App) renderMessagesView() {
	// create a new glamour renderer
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(m.messagesViewport.Width()-2-2), // 2 is the glamour internal gutter, extra 2 for the right side
	)
	if err != nil {
		m.err = err
		return
	}

	var content string

	for _, message := range m.session.Messages {
		var msg string
		if message.Role == llm.RoleUser {
			msg = lipgloss.NewStyle().Width(m.messagesViewport.Width()).Align(lipgloss.Right).Render(fmt.Sprintf("%s\n", message.Content))
		} else {
			msg, err = renderer.Render(message.Content)
			if err != nil {
				log.Fatal(err)
			}
		}

		content += msg
	}

	if m.session.IsStreaming() {
		c, err := renderer.Render(m.session.PartialMessage())
		if err != nil {
			log.Fatal(err)
		}

		content += c
	}

	m.messagesViewport.SetContent(content)
}

func (m *App) submitMessage() tea.Cmd {
	m.session.CancelStreaming()

	v := m.input.Value()
	if v != "" {
		m.session.AddMessage(llm.Message{Role: llm.RoleUser, Content: v})
		m.input.Reset()
	}

	m.session.StartStreaming()

	return complete(m)
}

func (m *App) viewMessagesInEditor() (tea.Cmd, error) {
	// build the content
	var content string
	for _, msg := range m.session.Messages {
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

	tmpFile := path.Join(tmpdir, "mark-messages.md")
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
	return lipgloss.JoinVertical(lipgloss.Top, m.mainView(), m.inputView())
}

func (m *App) mainView() string {
	return util.RenderBorderWithTitle(
		m.messagesView(),
		m.borderIfFocused(FocusedMessages),
		"Messages",
		m.panelTitleStyleIfFocused(FocusedMessages),
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

func (m *App) messagesView() string {
	return m.messagesViewport.View()
}

func (m *App) handleWindowSize(width, height int) {
	borderSize := 2 // 2 times the border width

	m.mainPanelWidth = width
	m.mainPanelHeight = height - inputHeight

	m.input.SetWidth(width - borderSize)
	m.input.SetHeight(inputHeight - borderSize)

	m.messagesViewport.SetWidth(m.mainPanelWidth - 2)   // 2 is the border width
	m.messagesViewport.SetHeight(m.mainPanelHeight - 2) // 2 is the border width

	if !m.uiReady {
		m.uiReady = true
	}
}
