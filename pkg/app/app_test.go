package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea/v2"
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

	model, cmd = model.Update(tea.WindowSizeMsg{Width: 160, Height: 90})

	app = model.(App)

	assert.Nil(t, cmd)
	assert.Equal(t, 160, app.width)
	assert.Equal(t, 90, app.height)
	assert.True(t, app.uiReady)
}
