package llm

import (
	"mark/internal/llm/provider"
	"mark/internal/llm/providers"
	"mark/internal/model"
)

type Agent struct {
	provider provider.Provider
}

func NewAgent() *Agent {
	return &Agent{
		provider: providers.NewOpenAIClient(),
	}
}

func (self *Agent) CompleteStreaming(c *model.Thread, s *model.StreamingMessage) error {
	return self.provider.CompleteStreaming(c, s)
}
