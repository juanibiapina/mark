package app

import (
	"fmt"
	"os"
)

type Project struct{}

func NewProject() *Project {
	return &Project{}
}

func (p *Project) Context() (string, error) {
	var c string = "You're in a project context\n"

	readme, err := p.ReadFile("README.md")
	if err != nil {
		return "", err
	}

	c += fmt.Sprintf("README.md:\n```\n%s\n```\n", readme)

	return c, nil
}

func (p *Project) ReadFile(filename string) (string, error) {
    content, err := os.ReadFile(filename)
    if err != nil {
			return "", err
    }
    return string(content), nil
}

