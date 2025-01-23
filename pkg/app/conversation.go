package app

import (
	"ant/pkg/llm"
	"ant/pkg/view"
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type Conversation struct {
	view.Focusable

	viewport viewport.Model
}

func (c *Conversation) Initialize(w int, h int) {
	c.viewport = viewport.New(w-c.BorderStyle().GetVerticalFrameSize(), h-c.BorderStyle().GetHorizontalFrameSize())
}

func (c *Conversation) ScrollToBottom() {
	c.viewport.GotoBottom()
}

func (c *Conversation) render(con *llm.Conversation, sm *llm.StreamingMessage) {
	messages := con.Messages()

	// create a new glamour renderer
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
	)

	if err != nil {
		log.Fatal(err)
	}

	// calculate number of messages to render
	var n int
	if sm == nil {
		n = 2
	} else {
		n = 1
	}

	var content string

	for i := len(messages) - n; i < len(messages); i++ {
		if i < 0 {
			continue
		}

		var m string
		if messages[i].Role == llm.RoleUser {
			m = lipgloss.NewStyle().Width(c.viewport.Width).Align(lipgloss.Right).Render(fmt.Sprintf("%s\n", messages[i].Content))
		} else {
			m, err = renderer.Render(messages[i].Content)

			if err != nil {
				log.Fatal(err)
			}
		}

		content += m
	}

	if sm != nil {
		c, err := renderer.Render(sm.Content)

		if err != nil {
			log.Fatal(err)
		}
		content += c
	}

	c.viewport.SetContent(content)
}

func (c Conversation) View() string {
	return c.BorderStyle().Render(c.viewport.View())
}

func (c *Conversation) SetSize(w int, h int) {
	c.viewport.Width = w - c.BorderStyle().GetVerticalFrameSize()
	c.viewport.Height = h - c.BorderStyle().GetHorizontalFrameSize()
}
