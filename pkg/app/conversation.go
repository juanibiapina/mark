package app

import (
	"ant/pkg/llm"
	"ant/pkg/view"
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type Conversation struct {
	view.Focusable

	viewport viewport.Model

	messages         []llm.Message
	StreamingMessage *StreamingMessage
}

func MakeConversation() Conversation {
	return Conversation{
		messages: []llm.Message{},
	}
}

func (c Conversation) Update(msg tea.Msg) (Conversation, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if !c.Focused() {
			return c, nil
		}

		switch msg.String() {

		case "q":
			return c, tea.Quit

		case "i":
			return c, message(focusInputMsg{})

		default:
			return c, nil
		}

	default:
		return c, nil
	}
}

func (c *Conversation) AddMessage(m llm.Message) {
	c.messages = append(c.messages, m)
	c.render()
}

func (c *Conversation) Messages() []llm.Message {
	return c.messages
}

func (c *Conversation) Reset() {
	c.CancelStreaming()
	c.messages = []llm.Message{}
	c.render()
}

func (c *Conversation) CancelStreaming() {
	if c.StreamingMessage == nil {
		return
	}

	c.StreamingMessage.Cancel()

	// Add the partial message to the chat history
	c.messages = append(c.messages, llm.Message{Role: llm.RoleAssistant, Content: c.StreamingMessage.Content})

	c.StreamingMessage = nil
}

func (c *Conversation) Initialize(w int, h int) {
	c.viewport = viewport.New(w-c.BorderStyle().GetVerticalFrameSize(), h-c.BorderStyle().GetHorizontalFrameSize())
}

func (c *Conversation) ScrollToBottom() {
	c.viewport.GotoBottom()
}

func (c *Conversation) render() {
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

	for i := len(c.messages) - n; i < len(c.messages); i++ {
		if i < 0 {
			continue
		}

		var m string
		if c.messages[i].Role == llm.RoleUser {
			m = lipgloss.NewStyle().Width(c.viewport.Width).Align(lipgloss.Right).Render(fmt.Sprintf("%s\n", c.messages[i].Content))
		} else {
			m, err = renderer.Render(c.messages[i].Content)

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
	return c.BorderStyle().Render(c.viewport.View())
}

func (c *Conversation) SetSize(w int, h int) {
	c.viewport.Width = w - c.BorderStyle().GetVerticalFrameSize()
	c.viewport.Height = h - c.BorderStyle().GetHorizontalFrameSize()
}
