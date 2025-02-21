package model

type PullRequest struct {
	Description string `json:"title"`
}

func (self PullRequest) Prompt() string {
	if self.Description == "" {
		return ""
	}

	content := "You are working on the following Pull Request:\n"
	content += "```\n" + self.Description + "\n```\n"

	return content
}
