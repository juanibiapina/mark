package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextItemFile(t *testing.T) {
	t.Parallel()

	t.Run("Title", func(t *testing.T) {
		t.Parallel()

		item, err := FileItem("testdata/file.txt")
		require.Nil(t, err)

		actual := item.Title()

		expected := "File: testdata/file.txt"
		assert.Equal(t, expected, actual)
	})

	t.Run("Message", func(t *testing.T) {
		t.Parallel()

		t.Run("when file exists", func(t *testing.T) {
			t.Parallel()

			item, err := FileItem("testdata/file.txt")
			require.NoError(t, err)

			actual := item.Message()

			expected := "File: testdata/file.txt\n```\nFile contents\n```\n"

			assert.Equal(t, expected, actual)
		})

		t.Run("when file does not exist", func(t *testing.T) {
			t.Parallel()

			item, err := FileItem("testdata/nonexistent.txt")
			require.NoError(t, err)

			actual := item.Message()

			expected := "File: testdata/nonexistent.txt\nFile does not exist.\n"

			assert.Equal(t, expected, actual)
		})
	})
}
