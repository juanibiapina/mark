package domain

import (
	"strings"

	"mark/internal/icon"

	"github.com/charmbracelet/bubbles/v2/list"
	"github.com/charmbracelet/x/ansi"
)

type ContextItemText struct {
	list.Item
	text string
}

func (item ContextItemText) Icon() string {
	return icon.Text
}

func (item ContextItemText) Title() string {
	text := ansi.Strip(item.text)
	return strings.ReplaceAll(text, "\n", " ")
}

func (item ContextItemText) Message() string {
	return item.text
}

func TextItem(text string) ContextItem {
	return ContextItemText{
		text: text,
	}
}
