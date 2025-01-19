package app

import (
	"ant/pkg/ai"
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
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

func (c *conversation) RenderMessages(messages []ai.Message, sm *ai.StreamingMessage) {
	messageViews := make([]string, len(messages))
	for i, msg := range messages {
		messageViews[i] = fmt.Sprintf("%s", msg.Content)
	}

	// Render the messages
	var content string
	if len(messageViews) > 0 {
		content = fmt.Sprintf("%s\n", lipgloss.JoinVertical(0, messageViews...))
	}

	// Render the streaming message
	if sm != nil {
		content += fmt.Sprintf("%s", sm.Content)
	}

	c.viewport.SetContent(content)
}

func (c conversation) View() string {
	return borderStyle.Render(c.viewport.View())
}

func (c *conversation) SetSize(w int, h int) {
	c.viewport.Width = w-borderStyle.GetVerticalFrameSize()
	c.viewport.Height = h-borderStyle.GetHorizontalFrameSize()
}
