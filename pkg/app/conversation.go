package app

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type Role int

const (
	RoleUser Role = iota
	RoleAssistant
)

type Message struct {
	Role    Role
	Content string
}

type Conversation struct {
	viewport viewport.Model

	Messages         []Message
	StreamingMessage *StreamingMessage
}

func MakeConversation() Conversation {
	return Conversation{
		Messages: []Message{},
	}
}

func (c *Conversation) Initialize(w int, h int) {
	c.viewport = viewport.New(w-borderStyle.GetVerticalFrameSize(), h-borderStyle.GetHorizontalFrameSize())
}

func (c *Conversation) ScrollToBottom() {
	c.viewport.GotoBottom()
}

func (c *Conversation) RenderMessagesTop() {
	// create a new glamour renderer
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
	)

	if err != nil {
		log.Fatal(err)
	}

	// calculate number of messages to render
	var n int
	if c.StreamingMessage == nil {
		n = 2
	} else {
		n = 1
	}

	var content string

	for i := len(c.Messages) - n; i < len(c.Messages); i++ {
		if i < 0 {
			continue
		}

		var m string
		if c.Messages[i].Role == RoleUser {
			m = lipgloss.NewStyle().Width(c.viewport.Width).Align(lipgloss.Right).Render(fmt.Sprintf("%s\n", c.Messages[i].Content))
		} else {
			m, err = renderer.Render(c.Messages[i].Content)

			if err != nil {
				log.Fatal(err)
			}
		}

		content += m
	}

	if c.StreamingMessage != nil {
		c, err := renderer.Render(c.StreamingMessage.Content)

		if err != nil {
			log.Fatal(err)
		}
		content += c
	}

	c.viewport.SetContent(content)
}

func (c Conversation) View() string {
	return borderStyle.Render(c.viewport.View())
}

func (c *Conversation) SetSize(w int, h int) {
	c.viewport.Width = w - borderStyle.GetVerticalFrameSize()
	c.viewport.Height = h - borderStyle.GetHorizontalFrameSize()
}
