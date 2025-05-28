package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextItemText(t *testing.T) {
	t.Parallel()

	t.Run("Title", func(t *testing.T) {
		t.Parallel()

		item := TextItem("some text content")

		actual := item.Title()

		expected := "some text content"
		assert.Equal(t, expected, actual)
	})

	t.Run("Message", func(t *testing.T) {
		t.Parallel()

		item := TextItem("some text content")

		actual := item.Message()

		expected := "some text content"
		assert.Equal(t, expected, actual)
	})
}
