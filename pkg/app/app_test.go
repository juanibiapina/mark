package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
)

func makeApp(t *testing.T) App {
	app, err := MakeApp()
	assert.Nil(t, err)
	return app
}

func TestApp(t *testing.T) {
	app := makeApp(t)

	v := app.View()
	assert.Equal(t, "Initializing...", v)

	model, _ := app.Update(tea.WindowSizeMsg{Width: 64, Height: 16})

	v = model.View()
	snaps.MatchStandaloneSnapshot(t, v)
}
