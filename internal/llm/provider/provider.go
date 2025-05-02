package provider

import "mark/internal/model"

type Provider interface {
	CompleteStreaming(c *model.Thread, s *model.StreamingMessage) error
}
