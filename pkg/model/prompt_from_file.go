package model

import "os"

type PromptFromFile struct {
	name     string
	filename string
}

func MakePromptFromFile(name, filename string) PromptFromFile {
	return PromptFromFile{
		name:     name,
		filename: filename,
	}
}

// startinterface: Prompt

func (f PromptFromFile) Name() string {
	return f.name
}

func (f PromptFromFile) Value() (string, error) {
	// Read file content
	content, err := os.ReadFile(f.filename)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// endinterface: Prompt
