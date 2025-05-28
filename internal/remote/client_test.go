package remote

import (
	"log/slog"
	"os"
	"testing"

	"mark/internal/app"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// set TMPDIR to /tmp to avoid issues with macos per user temp directories with long names.
	// example: /var/folders/ks/t5mwll9d0ys7xs_ng16n_qkc0000gn/T/
	// sockets have a limit of 104 characters.
	os.Setenv(("TMPDIR"), "/tmp/")

	v := m.Run()

	os.Exit(v)
}

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

			req := Request{
				Command: "test-message",
				Args:    []string{"arg1", "arg2"},
				Stdin:   "",
			}
			err = client.SendRequest(req)
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

			req := Request{
				Command: "add-context-item-text",
				Args:    []string{"prompt"},
				Stdin:   "stdin content",
			}
			err = client.SendRequest(req)
			require.NoError(t, err)

			msg := <-events
			slog.Info("Received message", "msg", msg)
			assert.Equal(t, app.AddContextItemTextMsg("prompt\nstdin content"), msg)
		})
	})
}
