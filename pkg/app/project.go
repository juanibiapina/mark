package app

import (
	"crypto/md5"
	"fmt"
	"os"
	"os/exec"

	"github.com/neovim/go-client/nvim"
)

type Project struct {
	entries []PromptEntry
}

func NewProject() *Project {
	entries := []PromptEntry{
		&FilePromptEntry{Filename: "README.md"},
		&ShellCommandPromptEntry{Command: "tree"},
	}

	if isGitRepo() {
		entries = append(entries,
			&ShellCommandPromptEntry{Command: "git", Args: []string{"status"}},
			&ShellCommandPromptEntry{Command: "git", Args: []string{"diff"}},
			&ShellCommandPromptEntry{Command: "git", Args: []string{"diff", "--cached"}},
			&ShellCommandPromptEntry{Command: "git", Args: []string{"log", "-n", "10"}},
		)
	}

	socket, err := determineNeovimSocket()
	if err != nil {
		panic(err) // TODO handle error
	}

	// TODO: find the correct socket path automatically (default place for current directory)
	entries = append(entries, &PromptEntryNeovimBuffers{Socket: socket})

	return &Project{entries: entries}
}

func determineNeovimSocket() (string, error) {
	// get cwd
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// create md5 of cwd
	md5 := md5.New()
	md5.Write([]byte(wd))
	hash := fmt.Sprintf("%x", md5.Sum(nil))

	// create socket path
	socket := fmt.Sprintf("/tmp/nvim.%s", hash)

	return socket, nil
}

func isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()
	return err == nil
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

type PromptEntryNeovimBuffers struct {
	Socket string
}

func (p *PromptEntryNeovimBuffers) Prompt() (string, error) {
	nvim, err := nvim.Dial(p.Socket)
	if err != nil {
		return "", err
	}
	defer nvim.Close()

	buffers, err := nvim.Buffers()
	if err != nil {
		return "", err
	}

	var output string
	for _, buffer := range buffers {
		isLoaded, err := nvim.IsBufferLoaded(buffer)
		if err != nil || !isLoaded {
			continue
		}

		name, err := nvim.BufferName(buffer)
		if err != nil || name == "" {
			continue
		}

		lines, err := nvim.BufferLines(buffer, 0, -1, true)
		if err != nil {
			return "", err
		}
		output += fmt.Sprintf("Buffer: %s\n```\n%s\n```\n", name, lines)
	}

	return output, nil
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
