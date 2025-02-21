package model

type Project struct {
	Cwd  string
}

func MakeProject(cwd string) (Project, error) {
	project := Project{
		Cwd: cwd,
	}

	return project, nil
}

func (self Project) Prompt() string {
		content := "You are working on the following project:\n"
		content += "```\n" + self.Cwd + "\n```\n"

		return content
}
