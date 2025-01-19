package ai

type StreamingMessage struct {
	Content string
}

func NewStreamingMessage() *StreamingMessage {
	return &StreamingMessage{}
}
