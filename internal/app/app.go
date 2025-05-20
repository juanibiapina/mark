package app

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"

	"mark/internal/llm"
	"mark/internal/util"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss/v2"
)

type (
	eventMsg            struct{ msg tea.Msg }
	streamStarted       struct{}
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

	uiReady bool
	width   int
	height  int
	main    *Main
	dialog  *AddContextDialog
}

func MakeApp(cwd string) (App, error) {
	// init events channel
	events := make(chan tea.Msg)

	// init app
	app := App{
		agent:   NewAgent(events),
		main:    NewMain(),
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

	// extract messages from event messages
	msg, cmd := m.processEventMessage(msg)
	cmds = append(cmds, cmd)

	// handle messages
	switch msg := msg.(type) {
	case errMsg:
		m.err = msg.err
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.handleWindowSize(msg.Width, msg.Height)

	case streamStarted:
		m.session.ClearReply()

	case streamChunkReceived:
		m.session.AppendChunk(string(msg))

	case streamFinished:
		m.session.FinishStreaming(string(msg))
	}

	// delegate to component update
	if m.dialog != nil {
		cmd := m.dialog.Update(&m, msg)
		cmds = append(cmds, cmd)
	} else {
		// delegate to main component
		cmd = m.main.Update(&m, msg)
		cmds = append(cmds, cmd)
	}

	m.renderMessagesView()

	return m, tea.Batch(cmds...)
}

func (m App) View() string {
	if !m.uiReady {
		return "Initializing..."
	}

	var view string

	view += m.main.View()

	if m.dialog != nil {
		dialogView := m.dialog.View()
		dialogWidth := lipgloss.Width(dialogView)
		dialogHeight := lipgloss.Height(dialogView)
		x := (m.width - dialogWidth) / 2
		y := (m.height - dialogHeight) / 2

		view = util.PlaceOverlay(x, y, dialogView, view)
	}

	return view
}

func (m *App) showAddContextDialog() {
	m.dialog = NewDialog()
	m.setDialogSize()
}

func (m *App) hideAddContextDialog() {
	m.dialog = nil
}

func (m *App) addContext(context string) {
	m.session.AddContext(context)
	m.main.contextItemsList.SetItemsFromSessionContext(m.session.Context())
}

func (m *App) setContextItemsFromSession() {
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

func (m *App) newSession() {
	m.agent.Cancel()

	m.session = llm.MakeSession()

	m.setContextItemsFromSession()
	m.main.input.Reset()

	m.main.focused = FocusedInput
}

func (m *App) renderMessagesView() {
	// create a new glamour renderer
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(m.main.messagesViewport.Width()-2-2), // 2 is the glamour internal gutter, extra 2 for the right side
	)
	if err != nil {
		m.err = err
		return
	}

	var content string

	// render the user message
	msg := lipgloss.NewStyle().Width(m.main.messagesViewport.Width()).Align(lipgloss.Right).Render(fmt.Sprintf("%s\n", m.session.Prompt()))

	content += msg

	// render the assistant message
	assistantMessage := m.session.Reply()
	if assistantMessage != "" {
		c, err := renderer.Render(assistantMessage)
		if err != nil {
			log.Fatal(err)
		}

		content += c
	}

	m.main.messagesViewport.SetContent(content)
}

func (m *App) submitMessage() tea.Cmd {
	v := m.main.input.Value()
	if v != "" {
		m.session.SetPrompt(v)
		m.main.input.Reset()
	}

	return complete(m)
}

func (m *App) viewMessagesInEditor() (tea.Cmd, error) {
	// build the content
	var content string

	content += m.session.Reply()

	editor := os.Getenv("EDITOR")
	if editor == "" {
		return nil, nil // TODO: handle this error
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

func (app *App) deleteContextItem(index int) {
	app.session.DeleteContextItem(index)
	app.main.contextItemsList.SetItemsFromSessionContext(app.session.Context())
}

func (m *App) handleWindowSize(width, height int) {
	m.width = width
	m.height = height

	m.main.SetSize(width, height)
	m.setDialogSize()

	if !m.uiReady {
		m.uiReady = true
	}
}

func (m *App) setDialogSize() {
	if m.dialog != nil {
		m.dialog.SetSize(m.width/2, 3) // TODO: this height is ignored
	}
}
