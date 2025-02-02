package model

import "context"

type StreamingMessage struct {
	Chunks  chan string
	Reply   chan string
	Ctx     context.Context
	cancel  context.CancelFunc
}

func NewStreamingMessage() *StreamingMessage {
	ctx, cancel := context.WithCancel(context.Background())

	return &StreamingMessage{
		Chunks:  make(chan string),
		Reply:   make(chan string),
		Ctx:     ctx,
		cancel:  cancel,
	}
}

func (s *StreamingMessage) Cancel() {
	if s.cancel != nil {
		s.cancel()
	}
}
