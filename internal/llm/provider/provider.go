package provider

import (
	"context"

	"mark/internal/model"
)

type StreamingEvent any

type StreamEventChunk struct {
	Chunk string
}

type StreamEventError struct {
	Error error
}

type StreamEventEnd struct {
	Message string
}

type Provider interface {
	CompleteStreaming(ctx context.Context, c model.Thread) (<-chan StreamingEvent, error)
}
