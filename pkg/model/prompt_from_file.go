package model

import "os"

type PromptFromFile struct {
	name     string
	Filename string
}

func MakePromptFromFile(name, filename string) PromptFromFile {
	return PromptFromFile{
		name:     name,
		Filename: filename,
	}
}

// startinterface: Prompt

func (f PromptFromFile) Name() string {
	return f.name
}

func (f PromptFromFile) Value() (string, error) {
	// Read file content
	content, err := os.ReadFile(f.Filename)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// endinterface: Prompt
