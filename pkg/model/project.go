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
