package model

import "time"

// ThreadEntry is an entry in the threads list.
type ThreadEntry struct {
	ID string
}

type Thread struct {
	ID        string    `json:"id"`
	Messages  []Message `json:"messages"`
	CreatedAt time.Time `json:"created_at"`

	streaming      bool
	partialMessage string
}

func MakeThread() Thread {
	t := time.Now()

	return Thread{
		ID:        t.Format(time.RFC3339Nano),
		CreatedAt: t,
	}
}

func (thread *Thread) AddMessage(m Message) {
	thread.Messages = append(thread.Messages, m)
}

func (thread *Thread) AppendChunk(chunk string) {
	thread.partialMessage += chunk
}

func (thread *Thread) FinishStreaming(msg string) {
	thread.AddMessage(Message{
		Role:    RoleAssistant,
		Content: msg,
	})
	thread.partialMessage = ""
}

func (thread *Thread) IsStreaming() bool {
	return thread.streaming
}

func (thread *Thread) StartStreaming() {
	thread.streaming = true
}

func (thread *Thread) CancelStreaming() {
	if !thread.streaming {
		return
	}
	thread.streaming = false
	if thread.partialMessage == "" {
		return
	}
	thread.AddMessage(Message{
		Role:    RoleAssistant,
		Content: thread.partialMessage,
	})
	thread.partialMessage = ""
}

func (thread *Thread) PartialMessage() string {
	return thread.partialMessage
}
