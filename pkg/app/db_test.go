package app

import (
	"testing"

	"mark/pkg/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func conversationWithID(id string) model.Conversation {
	c := model.MakeConversation()
	c.ID = id
	return c
}

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
		assert.Equal(t, c.ID, actual.ID)
	})

	t.Run("DeleteConversation", func(t *testing.T) {
		// given
		dir := t.TempDir()
		db := MakeFilesystemDatabase(dir)
		conversation := conversationWithID("1")

		err := db.SaveConversation(conversation)
		require.Nil(t, err)

		// when
		err = db.DeleteConversation(conversation.ID)
		require.Nil(t, err)

		// then
		_, err = db.LoadConversation(conversation.ID)
		assert.NotNil(t, err)

		entries, err := db.ListConversations()
		require.Nil(t, err)
		assert.Len(t, entries, 0)
	})

	t.Run("ListConversations", func(t *testing.T) {
		// given
		dir := t.TempDir()
		db := MakeFilesystemDatabase(dir)
		conversations := []model.Conversation{
			conversationWithID("1"),
			conversationWithID("2"),
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
		assert.Equal(t, conversations[1].ID, entries[0].ID)
		assert.Equal(t, conversations[0].ID, entries[1].ID)
	})
}
