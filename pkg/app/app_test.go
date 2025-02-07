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

	model, cmd := app.Init()
	assert.Nil(t, cmd)
	assert.Equal(t, app, model)

	model, cmd = model.Update(tea.WindowSizeMsg{Width: 64, Height: 16})

	app = model.(App)

	assert.Nil(t, cmd)
	assert.True(t, app.uiReady)

	v := model.View()
	snaps.MatchStandaloneSnapshot(t, v)
}
