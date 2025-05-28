package domain

import (
	"github.com/charmbracelet/bubbles/v2/list"
)

type ContextItem interface {
	list.Item
	Icon() string
	Title() string
	Message() string
}
