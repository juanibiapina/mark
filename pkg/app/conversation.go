package app

import (
	"github.com/charmbracelet/bubbles/viewport"
)

type conversation struct {
	viewport viewport.Model
}

func MakeConversation() conversation {
	return conversation{}
}

func (c *conversation) Initialize(w int, h int) {
	c.viewport = viewport.New(w-borderStyle.GetVerticalFrameSize(), h-borderStyle.GetHorizontalFrameSize())
}

func (c *conversation) SetSize(w int, h int) {
	c.viewport.Width = w-borderStyle.GetVerticalFrameSize()
	c.viewport.Height = h-borderStyle.GetHorizontalFrameSize()
}

func (c conversation) View() string {
	return borderStyle.Render(c.viewport.View())
}
