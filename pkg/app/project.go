package app

import (
	"fmt"
	"os"
	"os/exec"
)

type Project struct{}

func NewProject() *Project {
	return &Project{}
}

func (p *Project) runShellCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (p *Project) Context() (string, error) {
	var c string = "You're in a project context\n"

	readme, err := p.ReadFile("README.md")
	if err != nil {
		return "", err
	}

	gitStatus, err := p.runShellCommand("git", "status")
	if err != nil {
		return "", err
	}

	treeOutput, err := p.runShellCommand("tree")
	if err != nil {
		return "", err
	}

	gitLog, err := p.runShellCommand("git", "log", "-n", "10")
	if err != nil {
		return "", err
	}

	c += fmt.Sprintf("README.md:\n```\n%s\n```\n", readme)
	c += fmt.Sprintf("Directory tree:\n```\n%s\n```\n", treeOutput)
	c += fmt.Sprintf("Git status:\n```\n%s\n```\n", gitStatus)
	c += fmt.Sprintf("Git log (last 10 commits):\n```\n%s\n```\n", gitLog)

	return c, nil
}

func (p *Project) ReadFile(filename string) (string, error) {
    content, err := os.ReadFile(filename)
    if err != nil {
			return "", err
    }
    return string(content), nil
}

