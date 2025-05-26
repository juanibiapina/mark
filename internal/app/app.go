package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path"

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
	addContextItemTextMsg string
	addContextItemFileMsg string
	errMsg                struct{ err error }
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

	agent    *Agent
	events   chan tea.Msg
	listener net.Listener
	err      error

	uiReady bool
	width   int
	height  int
	main    *Main
	dialog  *InputDialog
}

func MakeApp(cwd string) (App, error) {
	// init events channel
	events := make(chan tea.Msg)

	// create socket file for listening to messages
	listener, err := createSocketFile(cwd)
	if err != nil {
		return App{}, fmt.Errorf("failed to create socket file: %w", err)
	}

	// init app
	app := App{
		agent:    NewAgent(events),
		main:     NewMain(),
		session:  domain.MakeSession(),
		events:   events,
		listener: listener,
	}

	return app, nil
}

func (m App) Err() error {
	return m.err
}

func (m App) Init() tea.Cmd {
	// TODO: should handleSocketMessages be a goroutine in Program?
	return tea.Batch(processEvents(m.events), handleSocketMessages(m.listener, m.events))
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
		m.session.SetReply(string(msg))

	case addContextItemTextMsg:
		m.addContextItem(domain.TextItem(string(msg)))

	case addContextItemFileMsg:
		m.addContextItem(domain.FileItem(string(msg)))
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

type ClientRequest struct {
	Message string   `json:"message"`
	Args    []string `json:"args,omitempty"`
}

// handleSocketMessages listens for incoming messages on the socket, converts them to tea.Msg and sends them to the app.
func handleSocketMessages(listener net.Listener, events chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		for {
			// accept a connection from the socket
			conn, err := listener.Accept()
			if err != nil {
				return errMsg{err}
			}
			defer conn.Close()

			// read messages from the connection
			scanner := bufio.NewScanner(conn)
			for scanner.Scan() {
				// parse JSON
				clientRequest := ClientRequest{}
				err := json.Unmarshal(scanner.Bytes(), &clientRequest)
				if err != nil {
					return errMsg{err}
				}

				// create a tea message from the client request
				var msg tea.Msg
				switch clientRequest.Message {
				case "add_context_item_text":
					msg = addContextItemTextMsg(clientRequest.Args[0])
				case "add_context_item_file":
					msg = addContextItemFileMsg(clientRequest.Args[0])
				default:
					msg = errMsg{fmt.Errorf("unknown message: %s", clientRequest.Message)}
				}

				events <- msg
			}
		}
	}
}

func (m *App) showAddContextDialog() {
	m.dialog = NewInputDialog(func(v string) {
		m.addContextItem(domain.TextItem(v))
	})
	m.setDialogSize()
}

func (m *App) showAddContextFileDialog() {
	m.dialog = NewInputDialog(func(v string) {
		m.addContextItem(domain.FileItem(v))
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
			return errMsg{err}
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

func createSocketFile(cwd string) (net.Listener, error) {
	// determine socket path
	socketPath := path.Join(cwd, ".local", "share", "mark", "socket")

	// remove existing socket file if it exists
	os.Remove(socketPath)

	// create the directory if it doesn't exist
	if err := os.MkdirAll(path.Dir(socketPath), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create directory for socket file: %w", err)
	}

	// create socket file
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list in socket: %w", err)
	}

	return listener, nil
}

func (m *App) addContextItem(item domain.ContextItem) {
	// add item to the session context
	m.session.Context().AddItem(item)

	// update the context items list in the main view
	m.main.contextItemsList.SetItemsFromSessionContextItems(m.session.Context().Items())
}
