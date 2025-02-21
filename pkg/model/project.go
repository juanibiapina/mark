package model

import (
	"os/exec"
)

type Project struct {
	Cwd string
}

func MakeProject(cwd string) (Project, error) {
	project := Project{
		Cwd: cwd,
	}

	return project, nil
}

func (self Project) Prompt() (string, error) {
	content := "You are working on the following project:\n"
	content += "```\n" + self.Cwd + "\n```\n"

	part, err := gitstatus()
	if err != nil {
		return "", err
	}
	content += part

	part, err = gitdiff()
	if err != nil {
		return "", err
	}

	content += part

	return content, nil
}

func gitstatus() (string, error) {
	output, err := runShellCommand("git", "status")
	if err != nil {
		return "", err
	}

	return taggedCodeBlock("Git status", output), nil
}

func gitdiff() (string, error) {
	output, err := runShellCommand("git", "diff")
	if err != nil {
		return "", err
	}

	return taggedCodeBlock("Git diff", output), nil
}

func taggedCodeBlock(tag string, content string) string {
	var result string
	result += tag + ":\n"
	result += "```\n" + content + "\n```\n"
	return result
}

func runShellCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
