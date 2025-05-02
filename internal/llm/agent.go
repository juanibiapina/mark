package llm

import (
	"mark/internal/model"
	"mark/internal/openai"
)

type Agent struct {
	ai *openai.OpenAI
}

func NewAgent() *Agent {
	return &Agent{
		ai: openai.NewOpenAIClient(),
	}
}

func (self *Agent) CompleteStreaming(c *model.Thread, s *model.StreamingMessage) error {
	return self.ai.CompleteStreaming(c, s)
}
