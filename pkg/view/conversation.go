package view

import (
	"ant/pkg/ai"
	"ant/pkg/model"
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type Conversation struct {
	viewport viewport.Model
}

func MakeConversation() Conversation {
	return Conversation{}
}

func (c *Conversation) Initialize(w int, h int) {
	c.viewport = viewport.New(w-borderStyle.GetVerticalFrameSize(), h-borderStyle.GetHorizontalFrameSize())
}

func (c *Conversation) ScrollToBottom() {
	c.viewport.GotoBottom()
}

func (c *Conversation) RenderMessages(messages []model.Message, sm *ai.StreamingMessage) {
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

func (c Conversation) View() string {
	return borderStyle.Render(c.viewport.View())
}

func (c *Conversation) SetSize(w int, h int) {
	c.viewport.Width = w-borderStyle.GetVerticalFrameSize()
	c.viewport.Height = h-borderStyle.GetHorizontalFrameSize()
}
