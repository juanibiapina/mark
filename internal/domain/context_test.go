package domain

import (
	"testing"

	"mark/internal/icon"

	"github.com/charmbracelet/bubbles/v2/list"
	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	t.Run("DeleteItem", func(t *testing.T) {
		c := NewContext()

		c.AddItem(mockItem("item 1"))
		assert.Equal(t, 1, len(c.Items()))

		c.AddItem(mockItem("item 2"))
		assert.Equal(t, 2, len(c.Items()))

		c.DeleteItem(0)
		assert.Equal(t, 1, len(c.Items()))

		c.DeleteItem(0)
		assert.Equal(t, 0, len(c.Items()))

		c.DeleteItem(0) // Deleting from an empty context should not panic
	})
}

func mockItem(text string) ContextItem {
	return &MockItem{text: text}
}

type MockItem struct {
	list.Item
	text string
}

func (item MockItem) Icon() string {
	return icon.Text
}

func (item MockItem) Title() string {
	return item.text
}

func (item MockItem) Message() string {
	return item.text
}
