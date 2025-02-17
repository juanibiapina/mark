package model

type Role int

const (
	RoleUser Role = iota
	RoleAssistant
)

type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}
