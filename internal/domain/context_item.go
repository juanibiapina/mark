package domain

import (
	"fmt"
	"os"
	"path/filepath"

	"mark/internal/icon"

	"github.com/charmbracelet/bubbles/v2/list"
)

// ListItem

type ListItem struct{}

func (i ListItem) FilterValue() string { return "" }

type ContextItem interface {
	list.Item
	Icon() string
	Title() string
	Message() string
}

// ContextItemText

type ContextItemText struct {
	ListItem
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

// ContextItemFile

type ContextItemFile struct {
	ListItem
	path string
}

func (item ContextItemFile) Icon() string {
	return icon.File
}

func (item ContextItemFile) Title() string {
	return "File: " + item.path
}

func (item ContextItemFile) Message() string {
	var result string
	result += "File: " + item.path + "\n"

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

func FileItem(path string) (ContextItem, error) {
	cwd, err := os.Getwd() // TODO: handle error
	if err != nil {
		return nil, err
	}

	if filepath.IsAbs(path) {
		var err error
		path, err = filepath.Rel(cwd, path)
		if err != nil {
			return nil, fmt.Errorf("failed to get relative path: %w", err)
		}
	}

	return ContextItemFile{
		path: path,
	}, nil
}
