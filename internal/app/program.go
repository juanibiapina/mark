package app

import (
	"os"

	tea "github.com/charmbracelet/bubbletea/v2"
)

// Program wraps the bubbletea program.
type Program struct {
	TeaProgram *tea.Program
}

// / NewProgram creates a new Program.
func NewProgram() (*Program, error) {
	// determine the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// initialize the App model
	m, err := MakeApp(cwd)
	if err != nil {
		return nil, err
	}

	// create the bubbletea program
	teaprogram := tea.NewProgram(m, tea.WithAltScreen(), tea.WithKeyboardEnhancements())

	// create the program
	program := &Program{
		TeaProgram: teaprogram,
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

	// handle errors in the final model after bubbletea program exits
	app := m.(App)
	err = app.Err()
	if err != nil {
		return err
	}

	return nil
}
