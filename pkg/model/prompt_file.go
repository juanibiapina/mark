package model

import (
	"fmt"
	"os"
	"strings"
)

type PromptFile struct {
	Filename string
}

// startinterface: Prompt

func (f PromptFile) Name() string {
	return "File: " + f.Filename
}

func (f PromptFile) Value() (string, error) {
	var output string

	// Return empty prompt if file does not exist
	if _, err := os.Stat(f.Filename); os.IsNotExist(err) {
		return "", nil
	}

	// Read file content
	content, err := os.ReadFile(f.Filename)
	if err != nil {
		return "", err
	}

	// Format output
	output += fmt.Sprintf("File: %s\n", f.Filename)
	output += "```\n"
	// Format output with line numbers
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		output += fmt.Sprintf("%d: %s\n", i+1, line)
	}
	output += "```\n"

	return output, nil
}

