package model

import "os"

type PromptFromFile struct {
	Filename string
}

func MakePromptFromFile(filename string) PromptFromFile {
	return PromptFromFile{
		Filename: filename,
	}
}

// startinterface: Prompt

func (f PromptFromFile) Value() (string, error) {
	// Read file content
	content, err := os.ReadFile(f.Filename)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// endinterface: Prompt
