package model

import "time"

// ThreadEntry is an entry in the threads list.
type ThreadEntry struct {
	ID string
}

type Thread struct {
	ID          string      `json:"id"`
	Messages    []Message   `json:"messages"`
	CreatedAt   time.Time   `json:"created_at"`
	PullRequest PullRequest `json:"pull_request"`
}

func MakeThread() Thread {
	t := time.Now()

	return Thread{
		ID:        t.Format(time.RFC3339Nano),
		CreatedAt: t,
	}
}

func (c *Thread) AddMessage(m Message) {
	c.Messages = append(c.Messages, m)
}
