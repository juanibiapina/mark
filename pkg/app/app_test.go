package app

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testConfig() Config {
	return Config{
		promptsDir: "testdata/prompts",
	}
}

func makeApp(t *testing.T) App {
	app, err := MakeApp(testConfig())
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

		err := fmt.Errorf("test error")

		model, cmd := model.Update(errMsg{err: err})

		assert.Equal(t, tea.QuitMsg{}, cmd())
		app = model.(App)
		require.ErrorIs(t, app.Err(), err)
	})
}
