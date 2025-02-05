package model

import (
	"os"
	"strings"
	"text/template"

	"mark/pkg/util"
)

type PromptMarkdown struct {
	name     string
	filename string
}

func MakePromptFromFile(name, filename string) PromptMarkdown {
	return PromptMarkdown{
		name:     name,
		filename: filename,
	}
}

type PromptContext struct{}

func (p PromptContext) ShellCommand(cmd string, args ...string) (string, error) {
	output, err := util.RunShellCommand(cmd, args...)
	if err != nil {
		return "", err
	}
	return output, nil
}

// startinterface: Prompt

func (f PromptMarkdown) Name() string {
	return f.name
}

func (f PromptMarkdown) Value() (string, error) {
	promptContext := PromptContext{}

	// Read file content
	content, err := os.ReadFile(f.filename)
	if err != nil {
		return "", err
	}

	// Parse template
	tmpl, err := template.New("prompt").Parse(string(content))
	if err != nil {
		return "", err
	}

	// Execute template
	builder := strings.Builder{}

	err = tmpl.Execute(&builder, promptContext)
	if err != nil {
		return "", err
	}

	return builder.String(), nil
}

// endinterface: Prompt
