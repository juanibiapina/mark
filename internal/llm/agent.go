package llm

import (
	"context"

	"mark/internal/llm/provider"
	"mark/internal/llm/providers"
	"mark/internal/model"
)

type Agent struct {
	provider provider.Provider

	cancel context.CancelFunc

	// streaming
	Streaming      bool
	Stream         *model.StreamingMessage
	PartialMessage string
}

func NewAgent() *Agent {
	return &Agent{
		provider: providers.NewOpenAIClient(),
	}
}

func (self *Agent) CompleteStreaming(c *model.Thread, s *model.StreamingMessage) error {
	ctx, cancel := context.WithCancel(context.Background())
	self.cancel = cancel
	return self.provider.CompleteStreaming(ctx, c, s)
}

func (self *Agent) Cancel() {
	if self.cancel != nil {
		self.cancel()
		self.cancel = nil
	}
}
