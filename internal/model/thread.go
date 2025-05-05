package model

import "time"

// ThreadEntry is an entry in the threads list.
type ThreadEntry struct {
	ID string
}

type Thread struct {
	ID             string    `json:"id"`
	Messages       []Message `json:"messages"`
	PartialMessage string    `json:"partial_message"`
	CreatedAt      time.Time `json:"created_at"`
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
	thread.PartialMessage += chunk
}

func (thread *Thread) FinishStreaming(msg string) {
	thread.AddMessage(Message{
		Role: RoleAssistant,
		Content: msg,
	})
	thread.PartialMessage = ""
}

func (thread *Thread) CancelStreaming() {
	if thread.PartialMessage == "" {
		return
	}
	thread.AddMessage(Message{
		Role: RoleAssistant,
		Content: thread.PartialMessage,
	})
	thread.PartialMessage = ""
}
