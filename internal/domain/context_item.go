package domain

import "github.com/charmbracelet/bubbles/v2/list"

type ListItem struct{}

func (i ListItem) FilterValue() string { return "" }

type ContextItem interface {
	list.Item
	Title() string
	Message() string
}

type ContextItemString string

func (i ContextItemString) FilterValue() string { return "" }
func (item ContextItemString) Title() string {
	return string(item)
}
func (item ContextItemString) Message() string {
	return string(item)
}

func TextItem(text string) ContextItem {
	return ContextItemString(text)
}
