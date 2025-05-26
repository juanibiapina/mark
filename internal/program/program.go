package program

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path"

	"mark/internal/app"

	tea "github.com/charmbracelet/bubbletea/v2"
)

// Program wraps the bubbletea program.
type Program struct {
	TeaProgram *tea.Program
	listener   net.Listener
	events     chan tea.Msg
}

// / NewProgram creates a new Program.
func NewProgram() (*Program, error) {
	// determine the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// create socket file for listening to messages
	listener, err := createSocketFile(cwd)
	if err != nil {
		return nil, fmt.Errorf("failed to create socket file: %w", err)
	}

	// create an events channel
	events := make(chan tea.Msg)

	// initialize the App model
	m, err := app.MakeApp(cwd, events)
	if err != nil {
		return nil, err
	}

	// create the bubbletea program
	teaprogram := tea.NewProgram(m, tea.WithAltScreen(), tea.WithKeyboardEnhancements())

	// create the program
	program := &Program{
		TeaProgram: teaprogram,
		listener:   listener,
		events:     events,
	}

	return program, nil
}

// Run runs the bubbletea program.
func (p *Program) Run() error {
	go handleSocketMessages(p.listener, p.events)

	// run the tea program
	m, err := p.TeaProgram.Run()
	if err != nil {
		return err
	}

	p.listener.Close() // close the listener when the program exits
	close(p.events)

	// assert the model is of type app.App
	app := m.(app.App)

	// handle errors in the final model after bubbletea program exits
	err = app.Err()
	if err != nil {
		return err
	}

	return nil
}

func createSocketFile(cwd string) (net.Listener, error) {
	// determine socket path
	socketPath := path.Join(cwd, ".local", "share", "mark", "socket")

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

type ClientRequest struct {
	Message string   `json:"message"`
	Args    []string `json:"args,omitempty"`
}

// handleSocketMessages listens for incoming messages on the socket, converts them to tea.Msg and sends them to the events channel.
func handleSocketMessages(listener net.Listener, events chan tea.Msg) {
	for {
		// accept a connection from the socket
		conn, err := listener.Accept()
		if err != nil {
			events <- app.ErrMsg{fmt.Errorf("failed to accept socket connection: %w", err)}
			return // TODO recover from listening failure
		}
		defer conn.Close()

		// read messages from the connection
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			// parse JSON
			clientRequest := ClientRequest{}
			err := json.Unmarshal(scanner.Bytes(), &clientRequest)
			if err != nil {
				events <- app.ErrMsg{fmt.Errorf("failed to parse client request: %w", err)}
				continue // skip to the next message
			}

			// create a tea message from the client request
			var msg tea.Msg
			switch clientRequest.Message {
			case "add_context_item_text":
				msg = app.AddContextItemTextMsg(clientRequest.Args[0])
			case "add_context_item_file":
				msg = app.AddContextItemFileMsg(clientRequest.Args[0])
			default:
				msg = app.ErrMsg{fmt.Errorf("unknown message: %s", clientRequest.Message)}
			}

			events <- msg
		}
	}
}
