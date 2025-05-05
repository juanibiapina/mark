package app

import (
	"context"
	"log/slog"

	"mark/internal/logging"
	"mark/internal/llm/provider"
	"mark/internal/llm/providers"
	"mark/internal/model"

	tea "github.com/charmbracelet/bubbletea/v2"
)

type Agent struct {
	provider provider.Provider
	cancel context.CancelFunc // cancels the current streaming request
	events chan tea.Msg // sends tea.Msg to the main app
	logger *slog.Logger
}

func NewAgent(events chan tea.Msg) *Agent {
	return &Agent{
		provider: providers.NewOpenAIClient(),
		events:   events,
		logger:   logging.NewLogger("agent"),
	}
}

func (agent *Agent) CompleteStreaming(thread model.Thread) error {
	ctx, cancel := context.WithCancel(context.Background())
	agent.cancel = cancel

	streamingEvents, err := agent.provider.CompleteStreaming(ctx, thread)
	if err != nil {
		return err
	}

	for streamingEvent := range streamingEvents {
		switch e := streamingEvent.(type) {
		case provider.StreamEventChunk:
			agent.logger.Info("Received StreamEventChunk", slog.String("chunk", e.Chunk)) // TODO should be debug
			agent.events <- streamChunkReceived(e.Chunk)

		case provider.StreamEventError:
			agent.logger.Error("Received StreamEventError", slog.String("error", e.Error.Error()))
			agent.events <- errMsg{err: e.Error}

		case provider.StreamEventEnd:
			agent.logger.Info("Received StreamEventEnd")
			agent.events <- streamFinished(e.Message)
		}
	}

	return nil
}

func (agent *Agent) Cancel() {
	if agent.cancel != nil {
		agent.cancel()
		agent.cancel = nil
	}
}
