package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Global struct {
}

func (a Global) Update(msg tea.Msg) (Global, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+n":
			return a, message(newConversationMsg{})

		case "ctrl+c":
			return a, message(cancelStreamingMsg{})

		}
	}

	return a, nil
}
