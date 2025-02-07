package app

import (
	"fmt"
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
	t.Parallel()

	t.Run("Initialization", func(t *testing.T) {
		app := makeApp(t)

		v := app.View()
		assert.Equal(t, "Initializing...", v)

		model, _ := app.Update(tea.WindowSizeMsg{Width: 64, Height: 16})

		v = model.View()
		snaps.MatchStandaloneSnapshot(t, v)
	})

	t.Run("Error", func(t *testing.T) {
		app := makeApp(t)
		model, _ := app.Update(tea.WindowSizeMsg{Width: 64, Height: 16})
		model, cmd := model.Update(errMsg{err: fmt.Errorf("test error")})

		assert.Equal(t, tea.QuitMsg{}, cmd())

		v := model.View()
		assert.Equal(t, "Error: test error", v)
	})
}
