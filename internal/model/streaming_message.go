package model

type StreamingMessage struct {
	Chunks chan string
	Reply  chan string
}

func NewStreamingMessage() *StreamingMessage {
	return &StreamingMessage{
		Chunks: make(chan string),
		Reply:  make(chan string),
	}
}
