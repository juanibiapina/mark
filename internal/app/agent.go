package app

import (
	"context"
	"log/slog"

	"mark/internal/domain"
	"mark/internal/llm"
	"mark/internal/llm/provider"
	"mark/internal/llm/providers"
	"mark/internal/logging"

	tea "github.com/charmbracelet/bubbletea/v2"
)

type Agent struct {
	provider  provider.Provider
	cancel    context.CancelFunc // cancels the current streaming request
	streaming bool               // true if the agent is currently streaming
	events    chan tea.Msg       // sends tea.Msg to the main app
	logger    *slog.Logger
}

func NewAgent(events chan tea.Msg) *Agent {
	return &Agent{
		provider: providers.NewOpenAIClient(),
		events:   events,
		logger:   logging.NewLogger("agent"),
	}
}

func convertSessionToMessages(session domain.Session) []llm.Message {
	var messages []llm.Message

	// add context message
	messages = append(messages, llm.Message{
		Role:    llm.RoleUser,
		Content: session.Context().Message(),
	})

	// add prompt
	messages = append(messages, llm.Message{
		Role:    llm.RoleUser,
		Content: session.Prompt(),
	})

	return messages
}

func (agent *Agent) Run(session domain.Session) error {
	if agent.streaming {
		agent.Cancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	agent.cancel = cancel

	agent.streaming = true

	messages := convertSessionToMessages(session)

	streamingEvents, err := agent.provider.CompleteStreaming(ctx, messages)
	if err != nil {
		return err
	}

	agent.events <- streamStarted{}

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

	agent.streaming = false

	return nil
}

func (agent *Agent) Cancel() {
	agent.streaming = false

	if agent.cancel != nil {
		agent.cancel()
		agent.cancel = nil
	}
}
