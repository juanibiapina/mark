package model

type Prompt interface {
	Name() string
	Value() (string, error)
}
