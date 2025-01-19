package ai

type StreamingMessage struct {
	Content string
	Chunks  chan string
	Reply   chan string
}

func NewStreamingMessage() *StreamingMessage {
	return &StreamingMessage{
		Content: "",
		Chunks:  make(chan string),
		Reply:   make(chan string),
	}
}
