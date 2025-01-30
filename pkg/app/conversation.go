package app

import (
	"fmt"
	"log"

	"ant/pkg/llm"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type Conversation struct {
	viewport viewport.Model
}

func (c *Conversation) ScrollToBottom() {
	c.viewport.GotoBottom()
}

func (c *Conversation) render(con *llm.Conversation, streaming bool, partialMessage string) {
	messages := con.Messages()

	// create a new glamour renderer
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(c.viewport.Width - 2), // 2 is the glamour internal gutter
	)
	if err != nil {
		log.Fatal(err)
	}

	// calculate number of messages to render
	var n int
	if !streaming {
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

	if streaming {
		c, err := renderer.Render(partialMessage)
		if err != nil {
			log.Fatal(err)
		}

		content += c
	}

	c.viewport.SetContent(content)
}

func (c Conversation) View() string {
	return c.viewport.View()
}

func (c *Conversation) SetSize(w int, h int) {
	c.viewport.Width = w
	c.viewport.Height = h
}

func (c *Conversation) Height() int {
	return c.viewport.Height
}

func (c *Conversation) LineDown() {
	c.viewport.LineDown(1)
}

func (c *Conversation) LineUp() {
	c.viewport.LineUp(1)
}
