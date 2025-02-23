package model

import (
	"os/exec"
	"strings"
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

	part, err := shellCommand("git", "status")
	if err != nil {
		return "", err
	}
	content += part

	part, err = shellCommand("git", "diff")
	if err != nil {
		return "", err
	}
	content += part

	part, err = shellCommand("git", "log", "--oneline", "...main")
	if err != nil {
		return "", err
	}
	content += part

	return content, nil
}

func shellCommand(command string, args ...string) (string, error) {
	output, err := runShellCommand(command, args...)
	if err != nil {
		return "", err
	}

	cmd := "`" + command + " " + strings.Join(args, " ") + "`"
	return taggedCodeBlock(cmd, output), nil
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
