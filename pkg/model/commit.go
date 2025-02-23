package model

type Commit struct {
	Description string `json:"description"`
}

func (self Commit) Prompt() string {
	if self.Description == "" {
		return ""
	}

	content := "You are working on the following commit:\n"
	content += "```\n" + self.Description + "\n```\n"

	return content
}
