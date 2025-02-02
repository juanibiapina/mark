package model

import (
	"crypto/md5"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/neovim/go-client/nvim"
)

type Prompt interface {
	Name() string
	Value() (string, error)
}

type PromptFile struct {
	Filename string
}

// startinterface: Prompt

func (f PromptFile) Name() string {
	return f.Filename
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

// endinterface: Prompt

type PromptShellCommand struct {
	Command string
	Args    []string
}

// startinterface: Prompt

func (s *PromptShellCommand) Name() string {
	return fmt.Sprintf("%s %v", s.Command, s.Args)
}

func (s *PromptShellCommand) Value() (string, error) {
	output, err := runShellCommand(s.Command, s.Args...)
	if err != nil {
		return "", err
	}

	// Format output
	return fmt.Sprintf("Command: %s %v\n```\n%s\n```\n", s.Command, s.Args, output), nil
}

func runShellCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// endinterface: Prompt

type PromptNeovimBuffers struct {
	Socket string
}

func NewPromptNeovimBuffers() *PromptNeovimBuffers {
	socket, err := determineNeovimSocket()
	if err != nil {
		panic(err) // TODO handle error
	}

	return &PromptNeovimBuffers{
		Socket: socket,
	}
}

// startinterface: Prompt

func (p *PromptNeovimBuffers) Name() string {
	return "Neovim Buffers: " + p.Socket
}

func (p *PromptNeovimBuffers) Value() (string, error) {
	nvim, err := nvim.Dial(p.Socket)
	if err != nil {
		return "", nil // TODO ignore when neovim is not running
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

		// Format output with line numbers
		output += fmt.Sprintf("Buffer: %s\n", name)
		output += "```\n"
		for i, line := range lines {
			output += fmt.Sprintf("%d: %s\n", i+1, line)
		}
		output += "```\n"
	}

	return output, nil
}

// endinterface: Prompt

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
