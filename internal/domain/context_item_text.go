package domain

import (
	"mark/internal/icon"

	"github.com/charmbracelet/bubbles/v2/list"
)

type ContextItemText struct {
	list.Item
	text string
}

func (item ContextItemText) Icon() string {
	return icon.Text
}

func (item ContextItemText) Title() string {
	return item.text
}

func (item ContextItemText) Message() string {
	return item.text
}

func TextItem(text string) ContextItem {
	return ContextItemText{
		text: text,
	}
}
