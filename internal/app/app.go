package app

import (
	"fmt"
	"log"

	"mark/internal/domain"
	"mark/internal/util"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss/v2"
)

type (
	eventMsg              struct{ msg tea.Msg }
	streamStarted         struct{}
	streamChunkReceived   string
	streamFinished        string
	AddContextItemTextMsg string
	AddContextItemFileMsg string
	RunMsg                struct{}
	NewSessionMsg         struct{}
	ErrMsg                struct{ Err error }
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

// TODO: rename App to Model
type App struct {
	session domain.Session

	agent  *Agent
	events chan tea.Msg
	err    error

	uiReady bool
	width   int
	height  int
	main    *Main
	dialog  *InputDialog
}

func MakeApp(cwd string, events chan tea.Msg) (App, error) {
	// init app
	app := App{
		agent:   NewAgent(events),
		main:    NewMain(),
		session: domain.MakeSession(),
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
	case ErrMsg:
		m.err = msg.Err
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.handleWindowSize(msg.Width, msg.Height)

	case streamStarted:
		m.session.ClearReply()

	case streamChunkReceived:
		m.session.AppendChunk(string(msg))

	case streamFinished:
		m.session.SetReply(string(msg))

	case AddContextItemTextMsg:
		m.addContextItem(domain.TextItem(string(msg)))

	case AddContextItemFileMsg:
		item, err := domain.FileItem(string(msg))
		if err != nil {
			m.err = err
			return m, tea.Quit
		}
		m.addContextItem(item)

	case RunMsg:
		cmds = append(cmds, runAgent(&m))

	case NewSessionMsg:
		m.newSession()
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
	m.dialog = NewInputDialog(func(v string) error {
		m.addContextItem(domain.TextItem(v))
		return nil
	})
	m.setDialogSize()
}

func (m *App) showAddContextFileDialog() {
	m.dialog = NewInputDialog(func(v string) error {
		item, err := domain.FileItem(v)
		if err != nil {
			return err
		}
		m.addContextItem(item)
		return nil
	})
	m.setDialogSize()
}

func (m *App) hideAddContextDialog() {
	m.dialog = nil
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

	m.session = domain.MakeSession()

	m.main.contextItemsList.SetItemsFromSessionContextItems(m.session.Context().Items())
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
	if v == "" {
		return nil
	}

	m.session.SetPrompt(v)
	m.main.input.Reset()

	return runAgent(m)
}

func (app *App) deleteContextItem(index int) {
	app.session.Context().DeleteItem(index)
	app.main.contextItemsList.SetItemsFromSessionContextItems(app.session.Context().Items())
}

func runAgent(m *App) tea.Cmd {
	return func() tea.Msg {
		err := m.agent.Run(m.session)
		if err != nil {
			return ErrMsg{err}
		}

		return nil
	}
}

func processEvents(events chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		v := <-events
		return eventMsg{v}
	}
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

func (m *App) addContextItem(item domain.ContextItem) {
	// add item to the session context
	m.session.Context().AddItem(item)

	// update the context items list in the main view
	m.main.contextItemsList.SetItemsFromSessionContextItems(m.session.Context().Items())
}
