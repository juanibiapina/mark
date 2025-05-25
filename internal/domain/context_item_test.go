package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextItemFile(t *testing.T) {
	t.Parallel()

	t.Run("Title", func(t *testing.T) {
		t.Parallel()

		item := FileItem("filename")

		actual := item.Title()

		expected := "File: filename"
		assert.Equal(t, expected, actual)
	})

	t.Run("Message", func(t *testing.T) {
		t.Parallel()

		t.Run("when file exists", func(t *testing.T) {
			t.Parallel()

			item := FileItem("testdata/file.txt")

			actual := item.Message()

			expected := "File: testdata/file.txt\n```\nFile contents\n```\n"

			assert.Equal(t, expected, actual)
		})

		t.Run("when file does not exist", func(t *testing.T) {
			t.Parallel()

			item := FileItem("testdata/nonexistent.txt")

			actual := item.Message()

			expected := "File: testdata/nonexistent.txt\nFile does not exist.\n"

			assert.Equal(t, expected, actual)
		})
	})
}
