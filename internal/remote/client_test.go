package remote

import (
	"log/slog"
	"mark/internal/app"
	"testing"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	t.Parallel()

	t.Run("NewClient", func(t *testing.T) {
		t.Parallel()

		_, err := NewClient("testdata")
		require.NoError(t, err)
	})

	t.Run("SendMessage", func(t *testing.T) {
		t.Run("socket path doesn't exist", func(t *testing.T) {
			client, err := NewClient("testdata/nonexistent")
			require.NoError(t, err)

			err = client.SendMessage("test-message", []string{"arg1", "arg2"})
			require.Error(t, err)
			assert.Equal(t, "Couldn't find socket path: testdata/nonexistent/.local/share/mark/socket", err.Error())
		})

		t.Run("socket path exists", func(t *testing.T) {
			t.Parallel()

			events := make(chan tea.Msg)
			cwd := t.TempDir()
			server, err := NewServer(cwd, events)
			require.NoError(t, err)
			go server.Run()
			defer server.Close()

			client, err := NewClient(cwd)
			require.NoError(t, err)

			err = client.SendMessage("add-context-item-text", []string{"prompt"})
			require.NoError(t, err)

			msg := <-events
			slog.Info("Received message", "msg", msg)
			assert.Equal(t, app.AddContextItemTextMsg("prompt"), msg)
		})
	})
}
