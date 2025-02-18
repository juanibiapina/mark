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

	t.Run("ListConversations", func(t *testing.T) {
		// given
		dir := t.TempDir()
		db := MakeFilesystemDatabase(dir)
		conversations := []model.Conversation{
			model.MakeConversation(),
			model.MakeConversation(),
		}

		for _, c := range conversations {
			err := db.SaveConversation(c)
			require.Nil(t, err)
		}

		// when
		entries, err := db.ListConversations()
		require.Nil(t, err)

		// then
		assert.Len(t, entries, 2)
		assert.Equal(t, conversations[0].ID, entries[0].ID)
		assert.Equal(t, conversations[1].ID, entries[1].ID)
	})
}
