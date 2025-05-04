package app

import (
	"context"
	"log/slog"

	"mark/internal/llm/provider"
	"mark/internal/llm/providers"
	"mark/internal/model"

	tea "github.com/charmbracelet/bubbletea/v2"
)

type Agent struct {
	provider provider.Provider

	cancel context.CancelFunc

	// events is a channel for sending messages to the main program
	events chan tea.Msg

	// streaming
	Streaming      bool
	PartialMessage string
}

func NewAgent(events chan tea.Msg) *Agent {
	return &Agent{
		provider: providers.NewOpenAIClient(),
		events:   events,
	}
}

func (self *Agent) CompleteStreaming(c model.Thread) error {
	self.Cancel()

	self.Streaming = true

	ctx, cancel := context.WithCancel(context.Background())
	self.cancel = cancel

	streamingEvents, err := self.provider.CompleteStreaming(ctx, c)
	if err != nil {
		return err
	}

	for streamingEvent := range streamingEvents {
		switch e := streamingEvent.(type) {
		case provider.StreamingEventChunk:
			slog.Info("Received chunk", slog.String("chunk", e.Chunk)) // TODO should be debug
			self.events <- partialMessage(e.Chunk)

		case provider.StreamingEventError:
			self.events <- errMsg{e.Error}

		case provider.StreamingEventEnd:
			slog.Info("Received end event")
			self.Streaming = false
			self.PartialMessage = ""
			self.events <- replyMessage(e.Message)
		}
	}

	return nil
}

func (self *Agent) Cancel() {
	if self.cancel != nil {
		self.cancel()
		self.cancel = nil
	}
}
