package model

import (
	"os/exec"
)

type PromptGitRepository struct {
	entries []Prompt
}

func NewPromptGitRepository() *PromptGitRepository {
	entries := []Prompt{}

	if isGitRepo() {
		entries = append(entries,
			&PromptShellCommand{Command: "git", Args: []string{"status"}},
			&PromptShellCommand{Command: "git", Args: []string{"diff"}},
			&PromptShellCommand{Command: "git", Args: []string{"diff", "--cached"}},
			&PromptShellCommand{Command: "git", Args: []string{"log", "-n", "10"}},
		)
	}

	return &PromptGitRepository{entries: entries}
}

func isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()
	return err == nil
}

// startinterface: Prompt

func (p *PromptGitRepository) Name() string {
	return "Git Repository"
}

func (p *PromptGitRepository) Value() (string, error) {
	var c string
	var tmp string
	var err error

	for _, entry := range p.entries {
		tmp, err = entry.Value()
		if err != nil {
			return "", err
		}
		c += tmp
	}

	return c, nil
}

// endinterface: Prompt
