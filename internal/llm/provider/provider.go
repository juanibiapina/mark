package provider

import (
	"context"

	"mark/internal/model"
)

type StreamingEvent any

type StreamingEventChunk struct {
	Chunk string
}

type StreamingEventError struct {
	Error error
}

type StreamingCancelled struct {}

type StreamingEventEnd struct {
	Message string
}

type Provider interface {
	CompleteStreaming(ctx context.Context, c model.Thread) (<-chan StreamingEvent, error)
}
