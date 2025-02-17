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

func initApp(t *testing.T) App {
	app := makeApp(t)
	model, _ := app.Init()
	model, _ = model.Update(tea.WindowSizeMsg{Width: 64, Height: 16})
	return model.(App)
}

func TestApp(t *testing.T) {
	t.Parallel()

	t.Run("Initialization", func(t *testing.T) {
		app := makeApp(t)

		v := app.View()
		assert.Equal(t, "Initializing...", v)

		model, _ := app.Init()
		model, _ = model.Update(tea.WindowSizeMsg{Width: 64, Height: 16})

		v = model.View()
		snaps.MatchStandaloneSnapshot(t, v)
	})

	t.Run("Error", func(t *testing.T) {
		app := initApp(t)

		err := fmt.Errorf("test error")

		model, cmd := app.Update(errMsg{err: err})

		assert.Equal(t, tea.QuitMsg{}, cmd())
		app = model.(App)
		require.ErrorIs(t, app.Err(), err)
	})

	t.Run("messages", func(t *testing.T) {

		t.Run("replyMessage", func(t *testing.T) {
			app := initApp(t)

			model, cmd := app.Update(replyMessage("test"))
			assert.NotNil(t, cmd)

			v := model.View()
			snaps.MatchStandaloneSnapshot(t, v)
		})

	})
}
