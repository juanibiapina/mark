package provider

import (
	"context"

	"mark/internal/model"
)

type Provider interface {
	CompleteStreaming(ctx context.Context, c model.Thread, s *model.StreamingMessage) error
}
