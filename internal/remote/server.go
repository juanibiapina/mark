package remote

import (
	"bufio"
	"encoding/json"
	"fmt"
	"mark/internal/app"
	"mark/internal/messages"
	"net"
	"os"
	"path"

	tea "github.com/charmbracelet/bubbletea/v2"
)

type Server struct {
	listener net.Listener
	events   chan tea.Msg
}

func NewServer(cwd string, events chan tea.Msg) (*Server, error) {
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

	server := &Server{
		listener: listener,
		events:  events,
	}

	return server, nil
}

func (s *Server) Run() {
	for {
		// accept a connection from the socket
		conn, err := s.listener.Accept()
		if err != nil {
			s.events <- app.ErrMsg{Err: fmt.Errorf("failed to accept socket connection: %w", err)}
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
				s.events <- app.ErrMsg{Err: fmt.Errorf("failed to parse client request: %w", err)}
				continue // skip to the next message
			}

			msg := messages.ToTeaMsg(clientRequest.Command, clientRequest.Args)

			s.events <- msg
		}
	}
}

func (s *Server) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
