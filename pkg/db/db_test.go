package db

import (
	"testing"

	"mark/pkg/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func threadWithID(id string) model.Thread {
	c := model.MakeThread()
	c.ID = id
	return c
}

func TestFilesystemDatabase(t *testing.T) {
	t.Parallel()

	t.Run("SaveThread", func(t *testing.T) {
		// given
		dir := t.TempDir()
		c := model.MakeThread()
		db := MakeDatabase(dir)

		// when
		err := db.SaveThread(c)
		require.Nil(t, err)

		// then
		actual, err := db.LoadThread(c.ID)
		require.Nil(t, err)
		assert.Equal(t, c.ID, actual.ID)
	})

	t.Run("DeleteThread", func(t *testing.T) {
		// given
		dir := t.TempDir()
		db := MakeDatabase(dir)
		thread := threadWithID("1")

		err := db.SaveThread(thread)
		require.Nil(t, err)

		// when
		err = db.DeleteThread(thread.ID)
		require.Nil(t, err)

		// then
		_, err = db.LoadThread(thread.ID)
		assert.NotNil(t, err)

		entries, err := db.ListThreads()
		require.Nil(t, err)
		assert.Len(t, entries, 0)
	})

	t.Run("ListThreads", func(t *testing.T) {
		// given
		dir := t.TempDir()
		db := MakeDatabase(dir)
		threads := []model.Thread{
			threadWithID("1"),
			threadWithID("2"),
		}

		for _, c := range threads {
			err := db.SaveThread(c)
			require.Nil(t, err)
		}

		// when
		entries, err := db.ListThreads()
		require.Nil(t, err)

		// then
		assert.Len(t, entries, 2)
		assert.Equal(t, threads[1].ID, entries[0].ID)
		assert.Equal(t, threads[0].ID, entries[1].ID)
	})
}
