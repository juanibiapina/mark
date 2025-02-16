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

	con            *model.Conversation
	streaming      bool
	partialMessage string
}

func (c *Conversation) Set(con *model.Conversation, streaming bool, partialMessage string) {
	c.con = con
	c.streaming = streaming
	c.partialMessage = partialMessage
}

func (c *Conversation) renderMessages() {
	messages := c.con.Messages()

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
	if !c.streaming {
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

	if c.streaming {
		c, err := renderer.Render(c.partialMessage)
		if err != nil {
			log.Fatal(err)
		}

		content += c
	}

	c.viewport.SetContent(content)
}

func (c Conversation) Render(width, height int) string {
	c.viewport.Width = width
	c.viewport.Height = height

	c.renderMessages()

	return c.viewport.View()
}
