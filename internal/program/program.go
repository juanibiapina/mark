package program

import (
	"fmt"
	"os"

	"mark/internal/app"
	"mark/internal/remote"

	tea "github.com/charmbracelet/bubbletea/v2"
)

// Program wraps the bubbletea program.
type Program struct {
	TeaProgram *tea.Program
	server     *remote.Server
	events     chan tea.Msg
}

// / NewProgram creates a new Program.
func NewProgram() (*Program, error) {
	// determine the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// create an events channel
	events := make(chan tea.Msg)

	// create server for listening to messages
	server, err := remote.NewServer(cwd, events)
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

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
		server:     server,
		events:     events,
	}

	return program, nil
}

// Run runs the bubbletea program.
func (p *Program) Run() error {
	go p.server.Run()

	// run the tea program
	m, err := p.TeaProgram.Run()
	if err != nil {
		return err
	}

	p.server.Close()
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
