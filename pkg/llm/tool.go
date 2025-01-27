package llm

type Tool interface {
	Invoke() string
}
