package app

import (
	"testing"

	"mark/pkg/model"

	"github.com/stretchr/testify/assert"
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
		actual, err := db.LoadConversation(c.ID)
		require.Nil(t, err)
		assert.Equal(t, c, actual)
	})
}
