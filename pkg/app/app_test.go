package app

import (
	"fmt"
	"path"
	"testing"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeApp(t *testing.T, cwd string) App {
	app, err := MakeApp(cwd)
	assert.Nil(t, err)
	return app
}

func bareApp(t *testing.T) App {
	cwd := t.TempDir()
	app := makeApp(t, cwd)
	model, _ := app.Update(tea.WindowSizeMsg{Width: 64, Height: 16})
	return model.(App)
}

func TestApp(t *testing.T) {
	t.Parallel()

	t.Run("Initialization", func(t *testing.T) {
		cwd := path.Join("testdata", "cwd")
		app := makeApp(t, cwd)

		v := app.View()
		assert.Equal(t, "Initializing...", v)

		model, cmd := app.Init()
		assert.NotNil(t, cmd)

		model, _ = model.Update(tea.WindowSizeMsg{Width: 64, Height: 16})

		msg := cmd()
		assert.NotNil(t, msg)
		model, _ = model.Update(msg)

		v = model.View()
		snaps.MatchStandaloneSnapshot(t, v)
	})

	t.Run("Error", func(t *testing.T) {
		app := bareApp(t)

		err := fmt.Errorf("test error")

		model, cmd := app.Update(errMsg{err: err})

		assert.Equal(t, tea.QuitMsg{}, cmd())
		app = model.(App)
		require.ErrorIs(t, app.Err(), err)
	})

	t.Run("messages", func(t *testing.T) {
		t.Run("replyMessage", func(t *testing.T) {
			app := bareApp(t)

			model, cmd := app.Update(replyMessage("test"))
			assert.NotNil(t, cmd)

			v := model.View()
			snaps.MatchStandaloneSnapshot(t, v)
		})
	})

	t.Run("focus", func(t *testing.T) {
		app := bareApp(t)

		// default focus
		require.Equal(t, FocusedInput, app.focused)
		v := app.View()
		snaps.MatchSnapshot(t, v)

		// tab once
		model, _ := app.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		app = model.(App)
		require.Equal(t, FocusedConversationList, app.focused)
		v = model.View()
		snaps.MatchSnapshot(t, v)

		// tab twice
		model, _ = app.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		app = model.(App)
		require.Equal(t, FocusedConversation, app.focused)
		v = model.View()
		snaps.MatchSnapshot(t, v)

		// tab thrice (back to input)
		model, _ = app.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		app = model.(App)
		require.Equal(t, FocusedInput, app.focused)
		v = model.View()
		snaps.MatchSnapshot(t, v)
	})
}
