package domain

import (
	"log/slog"
	"os"

	"github.com/charmbracelet/bubbles/v2/list"
)

// ListItem

type ListItem struct{}

func (i ListItem) FilterValue() string { return "" }

type ContextItem interface {
	list.Item
	Title() string
	Message() string
}

// ContextItemText

type ContextItemText struct {
	ListItem
	text string
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

// ContextItemFile

type ContextItemFile struct {
	ListItem
	path string
}

func (item ContextItemFile) Title() string {
	return "File: " + item.path
}

func (item ContextItemFile) Message() string {
	var result string
	result += "File: " + item.path + "\n"

	cwd, err := os.Getwd()
	slog.Info("CWD", "cwd", cwd)

	contents, err := os.ReadFile(item.path)
	if err != nil {
		if os.IsNotExist(err) {
			result += "File does not exist.\n"
		} else {
			result += "Error reading file: " + err.Error() + "\n"
		}
	} else {
		result += "```\n"
		result += string(contents)
		result += "```\n"
	}

	return result
}

func FileItem(path string) ContextItem {
	return ContextItemFile{
		path: path,
	}
}
