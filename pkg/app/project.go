package app

import (
	"fmt"
	"os"
	"os/exec"
)

type Project struct{
	entries []PromptEntry
}

func NewProject() *Project {
	return &Project{
		entries: []PromptEntry{
			&FilePromptEntry{Filename: "README.md"},
			&ShellCommandPromptEntry{Command: "tree"},
			&ShellCommandPromptEntry{Command: "git", Args: []string{"status"}},
			&ShellCommandPromptEntry{Command: "git", Args: []string{"diff"}},
			&ShellCommandPromptEntry{Command: "git", Args: []string{"log", "-n", "10"}},
		},
	}
}

type PromptEntry interface {
	Prompt() (string, error)
}

type FilePromptEntry struct {
	Filename string
}

func (f *FilePromptEntry) Prompt() (string, error) {
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
	output += fmt.Sprintf("File: %s\n```\n%s\n```\n", f.Filename, content)

	return output, nil
}

type ShellCommandPromptEntry struct {
	Command string
	Args    []string
}

func (s *ShellCommandPromptEntry) Prompt() (string, error) {
	output, err := runShellCommand(s.Command, s.Args...)
	if err != nil {
		return "", err
	}

	// Format output
	return fmt.Sprintf("Command: %s %v\n```\n%s\n```\n", s.Command, s.Args, output), nil
}

func (p *Project) Context() (string, error) {
	var c string = "You're in a project context\n"
	var tmp string
	var err error

	for _, entry := range p.entries {
		tmp, err = entry.Prompt()
		if err != nil {
			return "", err
		}
		c += tmp
	}

	return c, nil
}

func runShellCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

