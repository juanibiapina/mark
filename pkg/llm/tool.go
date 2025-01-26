package llm

type Tool interface {
	Invoke() string
}

type PingTool struct{}

func (t *PingTool) Invoke() string {
	return "pong"
}
