package app

type Prompt interface {
	Key() string
	Value() string
}

type PromptStatic struct {
	key   string
	value string
}

func MakePromptStatic(key, value string) PromptStatic {
	return PromptStatic{
		key:   key,
		value: value,
	}
}

// startinterface: Prompt
func (p PromptStatic) Key() string {
	return p.key
}

func (p PromptStatic) Value() string {
	return p.value
}
// endinterface: Prompt
