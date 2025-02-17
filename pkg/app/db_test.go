package app

import (
	"os"
	"path"
	"testing"

	"mark/pkg/model"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestFilesystemDatabase(t *testing.T) {
	t.Parallel()

	t.Run("SaveConversation", func(t *testing.T) {
		// given
		dir := t.TempDir()
		c := model.MakeConversation()
		db := MakeFilesystemDatabase(dir)

		// when
		err := db.SaveConversation(c)
		require.Nil(t, err)

		// then
		filename := c.ID + ".json"
		file := path.Join(dir, filename)
		content, err := os.ReadFile(file)
		require.Nil(t, err)

		snaps.MatchSnapshot(t, string(content))
	})
}
