package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextItemText(t *testing.T) {
	t.Parallel()

	t.Run("Title", func(t *testing.T) {
		t.Run("simple", func(t *testing.T) {
			t.Parallel()

			item := TextItem("some text content")

			actual := item.Title()

			expected := "some text content"
			assert.Equal(t, expected, actual)
		})

		t.Run("with newlines", func(t *testing.T) {
			t.Parallel()

			item := TextItem("some\ntext\ncontent")

			actual := item.Title()

			expected := "some text content"
			assert.Equal(t, expected, actual)
		})

		t.Run("with ansi escape sequences", func(t *testing.T) {
			t.Parallel()

			item := TextItem("\x1b[31msome\x1b[0m text \x1b[1mcontent\x1b[0m")

			actual := item.Title()

			expected := "some text content"
			assert.Equal(t, expected, actual)
		})
	})

	t.Run("Message", func(t *testing.T) {
		t.Parallel()

		item := TextItem("some text content")

		actual := item.Message()

		expected := "some text content"
		assert.Equal(t, expected, actual)
	})
}
