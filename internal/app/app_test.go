package app

import (
	"fmt"
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeApp(t *testing.T, cwd string) App {
	app, err := MakeApp(cwd, make(chan tea.Msg))
	require.Nil(t, err)
	return app
}

func bareApp(t *testing.T) App {
	cwd := t.TempDir()
	app := makeApp(t, cwd)
	model, _ := app.Update(tea.WindowSizeMsg{Width: 64, Height: 16})
	return model.(App)
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

func keymod(mod tea.KeyMod, code rune) tea.Msg {
	return tea.KeyPressMsg{Code: code, Text: string(code), Mod: mod}
}

func TestMain(m *testing.M) {
	// set TMPDIR to /tmp to avoid issues with macos per user temp directories with long names.
	// example: /var/folders/ks/t5mwll9d0ys7xs_ng16n_qkc0000gn/T/
	// sockets have a limit of 104 characters.
	os.Setenv(("TMPDIR"), "/tmp/")

	v := m.Run()

	snaps.Clean(m, snaps.CleanOpts{Sort: true})

	os.Exit(v)
}

func TestApp(t *testing.T) {
	t.Parallel()

	t.Run("auto scroll", func(t *testing.T) {
		t.Parallel()

		app := bareApp(t)
		model, _ := app.Update(streamStarted{})
		model, _ = model.Update(streamChunkReceived("1\n\n2\n\n3\n\n4\n\n5\n\n6"))
		v := render(t, model)
		snaps.MatchStandaloneSnapshot(t, v)

		model, _ = model.Update(streamChunkReceived("\n\n7\n\n8\n\n"))
		v = render(t, model)
		snaps.MatchStandaloneSnapshot(t, v)
	})

	t.Run("manual scroll", func(t *testing.T) {
		t.Parallel()

		app := bareApp(t)
		model, _ := app.Update(streamStarted{})
		model, _ = model.Update(streamChunkReceived("1\n\n2\n\n3\n\n4\n\n5\n\n6\n\n7\n\n8"))
		v := render(t, model)
		snaps.MatchStandaloneSnapshot(t, v)

		model, _ = model.Update(keymod(tea.ModShift, 'k'))
		v = render(t, model)
		snaps.MatchStandaloneSnapshot(t, v)

		model, _ = model.Update(keymod(tea.ModShift, 'k'))
		v = render(t, model)
		snaps.MatchStandaloneSnapshot(t, v)

		model, _ = model.Update(keymod(tea.ModShift, 'j'))
		v = render(t, model)
		snaps.MatchStandaloneSnapshot(t, v)
	})

	t.Run("messages", func(t *testing.T) {
		t.Run("streamFinished", func(t *testing.T) {
			app := bareApp(t)

			model, cmd := app.Update(streamFinished("test"))
			assert.Nil(t, cmd)
			v := render(t, model)
			snaps.MatchStandaloneSnapshot(t, v)

			model, cmd = app.Update(streamStarted{})
			assert.Nil(t, cmd)
			v = render(t, model)
			snaps.MatchStandaloneSnapshot(t, v)
		})

		t.Run("add-context-item-text", func(t *testing.T) {
			app := bareApp(t)

			model, cmd := app.Update(AddContextItemTextMsg("Test context item"))
			assert.Nil(t, cmd)
			v := render(t, model)
			snaps.MatchStandaloneSnapshot(t, v)
		})

		t.Run("add-context-item-file", func(t *testing.T) {
			app := bareApp(t)

			model, cmd := app.Update(AddContextItemFileMsg("test.txt"))
			assert.Nil(t, cmd)
			v := render(t, model)
			snaps.MatchStandaloneSnapshot(t, v)
		})

		t.Run("run", func(t *testing.T) {
			app := bareApp(t)

			model, cmd := app.Update(RunMsg{})
			assert.NotNil(t, cmd) // TODO: how to test that cmd is a tea.Cmd that runs the agent?
			v := render(t, model)
			snaps.MatchStandaloneSnapshot(t, v)
		})

		t.Run("new-session", func(t *testing.T) {
			app := bareApp(t)
			model, cmd := app.Update(AddContextItemFileMsg("test.txt"))

			model, cmd = model.Update(NewSessionMsg{})
			assert.Nil(t, cmd)
			v := render(t, model)
			snaps.MatchStandaloneSnapshot(t, v)
		})

		t.Run("ErrMsg", func(t *testing.T) {
			app := bareApp(t)

			err := fmt.Errorf("test error")
			model, cmd := app.Update(ErrMsg{Err: err})
			require.Nil(t, cmd)

			v := render(t, model)
			snaps.MatchStandaloneSnapshot(t, v)
		})
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
