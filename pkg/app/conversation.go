package app

import (
	"fmt"
	"log"

	"mark/pkg/model"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type Conversation struct {
	viewport viewport.Model
}

func (c *Conversation) renderMessages(m App) {
	messages := m.conversation.Messages()

	// create a new glamour renderer
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(c.viewport.Width-2-2), // 2 is the glamour internal gutter, extra 2 for the right side
	)
	if err != nil {
		log.Fatal(err)
	}

	// calculate number of messages to render
	var n int
	if !m.streaming {
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
		if messages[i].Role == model.RoleUser {
			m = lipgloss.NewStyle().Width(c.viewport.Width).Align(lipgloss.Right).Render(fmt.Sprintf("%s\n", messages[i].Content))
		} else {
			m, err = renderer.Render(messages[i].Content)
			if err != nil {
				log.Fatal(err)
			}
		}

		content += m
	}

	if m.streaming {
		c, err := renderer.Render(m.partialMessage)
		if err != nil {
			log.Fatal(err)
		}

		content += c
	}

	c.viewport.SetContent(content)
}

func (c Conversation) Render(m App, width, height int) string {
	c.viewport.Width = width
	c.viewport.Height = height

	c.renderMessages(m)

	return c.viewport.View()
}
