package app

import (
	"fmt"
	"os"
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

// handleMessage processes messages like in a tea.Program.
// This is needed for handling internal messages like tea.BatchMsg.
func handleMessage(t *testing.T, model tea.Model, msg tea.Msg) tea.Model {
	switch msg := msg.(type) {
	case tea.QuitMsg:
		t.Fatal("unexpected quit message")
	case tea.BatchMsg:
		for _, cmd := range msg {
			m := cmd()
			model, _ = model.Update(m)
		}
	default:
		model, _ = model.Update(msg)
	}
	return model
}

func update(app App, msg tea.Msg) App {
	model, _ := app.Update(msg)
	return model.(App)
}

// render a tea.Model to a string assuming it implements tea.ViewModel.
func render(t *testing.T, model tea.Model) string {
	vmodel, ok := model.(tea.ViewModel)
	require.True(t, ok)

	v := vmodel.View()
	return v
}

func key(code rune) tea.Msg {
	return tea.KeyPressMsg{Code: code, Text: string(code)}
}

func TestMain(m *testing.M) {
	v := m.Run()

	snaps.Clean(m, snaps.CleanOpts{Sort: true})

	os.Exit(v)
}

func TestApp(t *testing.T) {
	t.Parallel()

	t.Run("Initialization", func(t *testing.T) {
		cwd := path.Join("testdata", "cwd")
		app := makeApp(t, cwd)

		v := app.View()
		assert.Equal(t, "Initializing...", v)

		cmd := app.Init()
		assert.NotNil(t, cmd)

		model, _ := app.Update(tea.WindowSizeMsg{Width: 64, Height: 16})

		msg := cmd()
		model = handleMessage(t, model, msg)

		v = render(t, model)

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

			v := render(t, model)
			snaps.MatchStandaloneSnapshot(t, v)
		})
	})

	t.Run("focus", func(t *testing.T) {
		app := bareApp(t)
		focuses := []Focused{FocusedInput, FocusedThreadList, FocusedThread, FocusedInput}

		for _, expectedFocus := range focuses {
			require.Equal(t, expectedFocus, app.focused)
			v := app.View()
			snaps.MatchSnapshot(t, v)
			app = update(app, key(tea.KeyTab))
		}
	})

	t.Run("input", func(t *testing.T) {
		app := bareApp(t)

		app = update(app, key('h'))
		app = update(app, key('e'))
		app = update(app, key('y'))

		v := app.View()
		snaps.MatchSnapshot(t, v)

		app = update(app, key(tea.KeyEnter))
		v = app.View()
		snaps.MatchSnapshot(t, v)
	})
}
