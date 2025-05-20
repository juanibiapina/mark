package provider

import (
	"context"

	"mark/internal/llm"
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
	CompleteStreaming(ctx context.Context, messages []llm.Message) (<-chan StreamingEvent, error)
}
