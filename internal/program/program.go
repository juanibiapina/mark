package program

import (
	"os"

	"mark/internal/app"

	tea "github.com/charmbracelet/bubbletea/v2"
)

// Program wraps the bubbletea program.
type Program struct {
	TeaProgram *tea.Program
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
		events:     events,
	}

	return program, nil
}

// Run runs the bubbletea program.
func (p *Program) Run() error {
	// run the tea program
	m, err := p.TeaProgram.Run()
	if err != nil {
		return err
	}

	// assert the model is of type app.App
	app := m.(app.App)

	// handle errors in the final model after bubbletea program exits
	err = app.Err()
	if err != nil {
		return err
	}

	return nil
}
